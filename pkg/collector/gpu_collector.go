/*
Copyright (c) 2024, Shanghai Iluvatar CoreX Semiconductor Co., Ltd.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"); you may
not use this file except in compliance with the License. You may obtain
a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/deep-spark/go-ixml/pkg/ixml"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/utils"
)

var metricCollectors = map[string]func(device ixml.Device) interface{}{
	Temperature:     collectTemperature,
	FanSpeed:        collectFanSpeed,
	SmClock:         collectSmClock,
	MemClock:        collectMemClock,
	MemTotal:        collectTotalMemory,
	MemUsed:         collectUsedMemory,
	MemFree:         collectFreeMemory,
	MemUtilization:  collectMemUtilization,
	GpuUtilization:  collectGPUUtilization,
	PowerUsage:      collectPowerUsage,
	ProcessInfo:     collectProcessInfo,
	XidErrors:       collectXidErrors,
	EccSbeVolStatus: collectEccSbeVolStatus,
	EccDbeVolStatus: collectEccDbeVolStatus,
	SmUtilization:   collectSmUtilization,
}

type gpuCollector struct {
	mutex            sync.Mutex
	gpus             iluvatarGPU
	once             sync.Once
	devices          map[string]ixml.Device
	collectorConfigs []collectorConfig
}

func registerGpuCollector(ctx *ixContext, collectorConfigs []collectorConfig, gpus iluvatarGPU) {
	var collector subCollector

	collector = &gpuCollector{
		gpus:             gpus,
		collectorConfigs: collectorConfigs,
		devices:          make(map[string]ixml.Device),
	}
	ctx.registerCollector(collector)

	go collector.collect(ctx)
}

func (gc *gpuCollector) collect(ctx *ixContext) {
	gc.once.Do(func() {
		for uuid, _ := range gc.gpus.gpus {
			device, ret := ixml.GetHandleByUUID(uuid)
			if ret != ixml.SUCCESS {
				logger.IluvatarLog.Logger.Errorf("Unable to get Handle by uuid %v", ret)
				continue
			}
			gc.devices[uuid] = device
		}
	})

	for {
		select {
		case <-ctx.done():
			return
		case <-ctx.signal():
			logger.IluvatarLog.Infoln("Start to collect gpu metrics")
			gc.collectMetrics(ctx)
		}
	}
}

func (gc *gpuCollector) setLabelValue(key, value string) map[string]string {
	return map[string]string{
		key: value,
	}
}

func (gc *gpuCollector) collectMetrics(ctx *ixContext) {
	metrics := make(map[string][]metric)

	for uuid, gpu := range gc.gpus.gpus {
		device, ok := gc.devices[uuid]
		if !ok {
			logger.IluvatarLog.Logger.Errorf("Device not found for uuid: %s", uuid)
			continue
		}

		baseLabels := map[string]string{
			LabelUuid: uuid,
			LabelName: gpu.name,
			LabelGPU:  strconv.FormatUint(uint64(gpu.index), 10),
		}

		for _, config := range gc.collectorConfigs {
			if collectFunc, ok := metricCollectors[config.Name]; ok {
				var value float64
				var collectedValue interface{}
				var isProcessInfo bool

				if config.Name == ProcessInfo {
					isProcessInfo = true
					collectedValue = collectFunc(device)
				} else {
					if config.Name == SmUtilization {
						gpuQuerySupport, ret := device.GpmQueryDeviceSupport()
						if ret != ixml.SUCCESS || gpuQuerySupport.IsSupportedDevice == 0 {
							continue
						}
					}
					collectedValue = collectFunc(device)
					value, ok = collectedValue.(float64)
					if !ok {
						logger.IluvatarLog.Logger.Errorln("collectFunc returned non-float64")
						continue
					}
				}

				if isProcessInfo {
					infos, ok := collectedValue.([]ixml.Info)
					if !ok {
						logger.IluvatarLog.Logger.Errorln("collectFunc returned non-ProcessInfo")
						continue
					}
					if len(infos) == 0 {
						pidLabels := make(map[string]string, len(baseLabels)+2)
						for k, v := range baseLabels {
							pidLabels[k] = v
						}
						pidLabels[LabelProcessPid] = ""
						pidLabels[LabelProcessName] = ""
						metrics[uuid] = append(metrics[uuid], metric{
							name:   config.Name,
							labels: pidLabels,
							value:  0,
						})
					}
					for _, info := range infos {
						pidLabels := make(map[string]string, len(baseLabels)+2)
						for k, v := range baseLabels {
							pidLabels[k] = v
						}
						pidLabels[LabelProcessPid] = strconv.FormatUint(uint64(info.Pid), 10)
						pidLabels[LabelProcessName] = getProcessNameByPid(info.Pid)
						value = float64(info.UsedGpuMemory / 1024 / 1024) // to MiB
						metrics[uuid] = append(metrics[uuid], metric{
							name:   config.Name,
							labels: pidLabels,
							value:  value,
						})
					}
				} else {
					metrics[uuid] = append(metrics[uuid], metric{
						name:   config.Name,
						labels: baseLabels,
						value:  value,
					})
				}
			}
		}
	}
	ctx.updateMetrics(metrics)
}

func collectTemperature(device ixml.Device) interface{} {
	temperature, ret := device.GetTemperature()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU temperature of device: %v", ret)

	}
	return float64(temperature)
}

func collectFanSpeed(device ixml.Device) interface{} {
	speed, ret := device.GetFanSpeed()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU FanSpeed of device: %v", ret)
		return 0
	}

	return float64(speed)
}

func collectSmClock(device ixml.Device) interface{} {
	clock, ret := device.GetClockInfo()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU SmClock of device: %v", ret)
		return 0
	}

	return float64(clock.Sm)
}

func collectMemClock(device ixml.Device) interface{} {
	clock, ret := device.GetClockInfo()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU MemsClock of device: %v", ret)
		return 0
	}

	return float64(clock.Mem)
}

func collectTotalMemory(device ixml.Device) interface{} {
	mem, ret := device.GetMemoryInfo()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU MemoryInfo of device: %v", ret)
		return 0
	}

	return float64(mem.Total)
}

func collectUsedMemory(device ixml.Device) interface{} {
	mem, ret := device.GetMemoryInfo()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU MemoryInfo of device: %v", ret)
		return 0
	}

	return float64(mem.Used)
}

func collectFreeMemory(device ixml.Device) interface{} {
	mem, ret := device.GetMemoryInfo()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU MemoryInfo of device: %v", ret)
		return 0
	}

	return float64(mem.Free)
}

func collectPowerUsage(device ixml.Device) interface{} {
	usage, ret := device.GetPowerUsage()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get usage %v", ret)
		return 0
	}

	return float64(usage)
}

func collectMemUtilization(device ixml.Device) interface{} {
	utilization, ret := device.GetUtilizationRates()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU Memory utilizationRates of device %v", ret)
		return 0
	}

	return float64(utilization.Memory)
}

func collectGPUUtilization(device ixml.Device) interface{} {
	utilization, ret := device.GetUtilizationRates()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get GPU utilizationRates of device %v", ret)
		return 0
	}

	return float64(utilization.Gpu)
}

func getProcessNameByPid(pid uint32) string {
	var cmdlinePath string
	if utils.IsDocker() {
		cmdlinePath = fmt.Sprintf("/host-proc/%d/cmdline", pid)
	} else {
		cmdlinePath = fmt.Sprintf("/proc/%d/cmdline", pid)
	}

	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		logger.IluvatarLog.Logger.Errorf("Error reading cmdline file for pid %d: %v", pid, err)
		return ""
	}
	return strings.TrimSuffix(string(data), "\x00")
}

func collectProcessInfo(device ixml.Device) interface{} {
	processInfos, ret := device.GetComputeRunningProcesses()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get processInfos: %v", ret)
		return nil
	}
	return processInfos
}

func collectXidErrors(device ixml.Device) interface{} {
	clocksThrottleReasons, ret := device.GetCurrentClocksThrottleReasons()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get clocksThrottleReasons: %v", ret)
		return nil
	}
	return float64(clocksThrottleReasons)
}

func collectEccSbeVolStatus(device ixml.Device) interface{} {
	singleErr, _, ret := device.GetEccErros()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get ECC SBE Volatile: %v", ret)
		return nil
	}
	return float64(singleErr)
}

func collectEccDbeVolStatus(device ixml.Device) interface{} {
	_, doubleErr, ret := device.GetEccErros()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get ECC DBE Volatile: %v", ret)
		return nil
	}
	return float64(doubleErr)
}

func collectSmUtilization(device ixml.Device) interface{} {
	sample1, ret := ixml.GpmSampleAlloc()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to allocate GPM sample: %v", ret)
		return nil
	}
	defer func() {
		_ = sample1.Free()
	}()
	sample2, ret := ixml.GpmSampleAlloc()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to allocate GPM sample: %v", ret)
		return nil
	}
	defer func() {
		_ = sample2.Free()
	}()

	if ret := device.GpmSampleGet(sample1); ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("could not get GPM sample: %w", ret)
		return nil
	}
	time.Sleep(1 * time.Second)
	if ret := device.GpmSampleGet(sample2); ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("could not get GPM sample: %w", ret)
		return nil
	}

	gpmMetric := ixml.GpmMetricsGetType{
		NumMetrics: 1,
		Sample1:    sample1,
		Sample2:    sample2,
		Metrics: [98]ixml.GpmMetric{
			{
				MetricId: uint32(ixml.GPM_METRIC_SM_UTIL),
			},
		},
	}
	ret = ixml.GpmMetricsGet(&gpmMetric)
	if ret != ixml.SUCCESS {
		return fmt.Errorf("failed to get gpm metric: %w", ret)
	}

	return float64(gpmMetric.Metrics[0].Value)
}
