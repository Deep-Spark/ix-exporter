/*
Copyright (c) 2024, Shanghai Iluvatar CoreX Semiconductor Co., Ltd.
All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package exporter

import (
	"strconv"
	"sync"

	"github.com/golang/glog"

	"gitee.com/deep-spark/ixexporter/pkg/ixml"
)

const (
	temperature    resourceType = "ix_temperature"
	fanSpeed                    = "ix_fan_speed"
	smClock                     = "ix_sm_clock"
	memClock                    = "ix_mem_clock"
	memTotal                    = "ix_mem_total"
	memUsed                     = "ix_mem_used"
	memFree                     = "ix_mem_free"
	memUtilization              = "ix_mem_utilization"
	gpuUtilization              = "ix_gpu_utilization"
	powerUsage                  = "ix_power_usage"
)

// gpuCollector
type gpuCollector struct {
	gpus    iluvatarGPU
	metrics []Metric
	devices map[string]ixml.Device
	once    sync.Once
}

func (gc *gpuCollector) collect(ctx *ixContext) {
	gc.once.Do(func() {
		for uuid, _ := range gc.gpus.gpus {
			device, err := ixml.NewDeviceByUUID(uuid)
			if err != nil {
				glog.Errorln(err)
				continue
			}
			gc.devices[uuid] = device
		}
	})

	labels := []string{"gpu", "name", "uuid"}
	ctx.updateLabels(labels)

	for {
		select {
		case <-ctx.done():
			return
		case <-ctx.signal():
			gc.collectMetrics(ctx)
		}
	}
}

func (gc *gpuCollector) collectMetrics(ctx *ixContext) error {
	metrics := make(map[string][]metric)

	for uuid, gpu := range gc.gpus.gpus {
		var m metric
		var ms []metric
		values := gc.getLabelsValues(uuid, gpu)
		for _, mm := range gc.metrics {
			switch mm.Name {
			case temperature:
				m.name = temperature
				m.value = gc.collectTemperature(gc.devices[uuid])
				m.labels = values
			case fanSpeed:
				m.name = fanSpeed
				m.value = gc.collectFanSpeed(gc.devices[uuid])
				m.labels = values
			case smClock:
				m.name = smClock
				m.value = gc.collectSmClock(gc.devices[uuid])
				m.labels = values
			case memClock:
				m.name = memClock
				m.value = gc.collectMemClock(gc.devices[uuid])
				m.labels = values
			case memTotal:
				m.name = memTotal
				m.value = gc.collectTotalMemory(gc.devices[uuid])
				m.labels = values
			case memUsed:
				m.name = memUsed
				m.value = gc.collectUsedMemory(gc.devices[uuid])
				m.labels = values
			case memFree:
				m.name = memFree
				m.value = gc.collectFreeMemory(gc.devices[uuid])
				m.labels = values
			case memUtilization:
				m.name = memUtilization
				m.value = gc.collectMemUtilization(gc.devices[uuid])
				m.labels = values
			case gpuUtilization:
				m.name = gpuUtilization
				m.value = gc.collectGPUUtilization(gc.devices[uuid])
				m.labels = values
			case powerUsage:
				m.name = powerUsage
				m.value = gc.collectPowerUsage(gc.devices[uuid])
				m.labels = values
			}

			ms = append(ms, m)
		}

		metrics[uuid] = ms
	}

	ctx.updateMetrics(metrics)

	return nil
}

func (gc *gpuCollector) collectTemperature(device ixml.Device) float64 {
	temperature, err := device.DeviceGetTemperature()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(temperature)
}

func (gc *gpuCollector) collectFanSpeed(device ixml.Device) float64 {
	speed, err := device.DeviceGetFanSpeed()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(speed)
}

func (gc *gpuCollector) collectSmClock(device ixml.Device) float64 {
	clock, err := device.DeviceGetClockInfo()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(clock.Sm)
}

func (gc *gpuCollector) collectMemClock(device ixml.Device) float64 {
	clock, err := device.DeviceGetClockInfo()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(clock.Mem)
}

func (gc *gpuCollector) collectTotalMemory(device ixml.Device) float64 {
	mem, err := device.DeviceGetMemoryInfo()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(mem.Total)
}

func (gc *gpuCollector) collectUsedMemory(device ixml.Device) float64 {
	mem, err := device.DeviceGetMemoryInfo()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(mem.Used)
}

func (gc *gpuCollector) collectFreeMemory(device ixml.Device) float64 {
	mem, err := device.DeviceGetMemoryInfo()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(mem.Free)
}

func (gc *gpuCollector) collectPowerUsage(device ixml.Device) float64 {
	usage, err := device.DeviceGetPowerUsage()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(usage)
}

func (gc *gpuCollector) collectMemUtilization(device ixml.Device) float64 {
	utilization, err := device.DeviceGetUtilization()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(utilization.Mem)
}

func (gc *gpuCollector) collectGPUUtilization(device ixml.Device) float64 {
	utilization, err := device.DeviceGetUtilization()
	if err != nil {
		glog.Errorln(err)
		return 0
	}

	return float64(utilization.GPU)
}

func (gc *gpuCollector) getLabelsValues(uuid string, info gpuInfo) labelType {
	return labelType{
		"gpu":  strconv.FormatUint(uint64(info.index), 10),
		"name": info.name,
		"uuid": uuid}
}

func registerGpuCollector(ctx *ixContext, metrics []Metric, gpus iluvatarGPU) {
	var collector subCollector

	collector = &gpuCollector{
		gpus:    gpus,
		metrics: metrics,
		devices: make(map[string]ixml.Device),
	}
	ctx.registerCollector(collector)

	go collector.collect(ctx)
}
