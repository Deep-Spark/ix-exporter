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
	"time"

	"gitee.com/deep-spark/go-ixml/pkg/ixml"
	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type subCollector interface {
	collect(ctx *ixContext)
}

type iluvatarCollector struct {
	opts            *Options
	collectorConfig []collectorConfig
	resources       map[string]*prometheus.Desc
	gpus            iluvatarGPU
	labels          []string
	ctx             *ixContext
}

func initIXMLAndCheckDrivers(info *iluvatarGPU) error {
	var ret ixml.Return

	ret = ixml.Init()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Errorf("Unable to initialize IXML: %v", ret)
		return fmt.Errorf("unable to initialize IXML: %v", ret)
	}

	info.count, ret = ixml.DeviceGetCount()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Errorln("Unable to get device count: %v", ret)
		return fmt.Errorf("Unable to get device count: %v", ret)
	}

	info.driverVersion, ret = ixml.SystemGetDriverVersion()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get driver version: %v", ret)
	}

	info.cudaVersion, ret = ixml.SystemGetCudaDriverVersion()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get cuda driver version: %v", ret)
	}
	return nil
}

func processDeviceAtIndex(info *iluvatarGPU, index uint, chipList []chip, chipmap map[chip]bool) error {
	var device ixml.Device
	gpu := gpuInfo{
		index: index,
	}

	ret := ixml.DeviceGetHandleByIndex(index, &device)
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Errorf("Unable to get device at index %d: %v", index, ret)
		return fmt.Errorf("Unable to get device at index %d: %v", index, ret)
	}

	gpu.name, ret = device.GetName()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get name %v", ret)
	}

	uuid, ret := device.GetUUID()
	if ret != ixml.SUCCESS {
		logger.IluvatarLog.Logger.Warningf("Unable to get device uuid %v", ret)
	}

	pos, ret := device.GetBoardPosition()
	if ret != ixml.SUCCESS {
		if ret == ixml.ERROR_NOT_SUPPORTED {
			logger.IluvatarLog.Logger.Infof("GPU %s not support splitboard.\n", gpu.name)
			info.pairChips[uuid] = uuid
		} else {
			logger.IluvatarLog.Logger.Warningf("Unable to get BoardPosition %v", ret)
		}
	} else {
		logger.IluvatarLog.Logger.Infof("GPU %s on board %d.\n", gpu.name, pos)
		key := chip{
			uuid:      uuid,
			operation: device,
		}

		chipmap[key] = false
		chipList = append(chipList, key)
	}

	info.gpus[uuid] = gpu
	return nil
}

func collectChipData(info *iluvatarGPU, chipmap map[chip]bool) []chip {
	var chipList []chip

	for index := uint(0); index < info.count; index++ {
		if err := processDeviceAtIndex(info, index, chipList, chipmap); err != nil {
			break
		}
	}

	return chipList
}

func getDeviceInfo() iluvatarGPU {
	var info iluvatarGPU
	info.pairChips = make(map[string]string)
	chipmap := make(map[chip]bool)
	info.gpus = make(map[string]gpuInfo)

	if err := initIXMLAndCheckDrivers(&info); err != nil {
		return info
	}

	chipList := collectChipData(&info, chipmap)
	if len(chipList) == 0 {
		logger.IluvatarLog.Logger.Errorf("No chips detected")
		return info
	}
	for i, first := range chipList {
		if chipmap[first] {
			continue
		}
		for j := i + 1; j < len(chipList); j++ {
			second := chipList[j]
			onSameBoard, ret := ixml.GetOnSameBoard(first.operation, chipList[j].operation)
			if ret != ixml.SUCCESS {
				if ret == ixml.ERROR_NOT_SUPPORTED {
					logger.IluvatarLog.Logger.Warningf("GetOnSameBoard: Not supported\n")
				} else {
					logger.IluvatarLog.Logger.Errorf("Unable to get OnSameBoard %v", ret)
				}
				continue
			}
			if onSameBoard == 1 {
				chipmap[first] = true
				chipmap[second] = true
				info.pairChips[first.uuid] = second.uuid
				info.pairChips[second.uuid] = first.uuid
				break
			}
		}
	}

	return info
}

func getMetricConfig(mcs config.ExporterConfig) []collectorConfig {
	m := make([]collectorConfig, len(mcs.Metrics))
	for i, mc := range mcs.Metrics {
		m[i].Name = mc.Name
		m[i].Help = mc.Help
	}
	return m
}

func NewIluvatarCollector(opts *Options) (*iluvatarCollector, error) {

	cfg := config.Config{
		ConfigFile: opts.MetricsConfig,
		IxExporter: make(map[string]config.ExporterConfig),
	}
	if err := cfg.ParseConfig(); err != nil {
		logger.IluvatarLog.Errorf("Error parsing config: %s", err)
		return nil, err
	}
	iluvatarConfig, ok := cfg.IxExporter[Iluvatar]
	if !ok {
		logger.IluvatarLog.Errorf("Iluvatar configuration not found")
		return nil, fmt.Errorf("iluvatar configuration not found")
	}
	ml := getMetricConfig(iluvatarConfig)

	var labels []string
	if opts.EnableKube {
		labels = LabelAllList
	} else {
		labels = LabelList
	}

	return &iluvatarCollector{
		opts:            opts,
		gpus:            getDeviceInfo(),
		resources:       make(map[string]*prometheus.Desc),
		collectorConfig: ml,
		labels:          labels,
		ctx:             nil,
	}, nil
}

// Describe is the implementation of the interface of 'prometheus.Collecter.Describe()', once
// 'prometheus.MustRegtister()' or 'prometheus.Unregister()' was called, it will be triggered.
func (ic *iluvatarCollector) Describe(ch chan<- *prometheus.Desc) {

	logger.IluvatarLog.Info("Describe() called...")
	if ic.ctx == nil {
		ic.ctx = newContext()
		registerGpuCollector(ic.ctx, ic.collectorConfig, ic.gpus)
		if ic.opts.EnableKube {
			registerKubeCollector(ic.ctx, ic.gpus)
		}
		for _, mc := range ic.collectorConfig {
			var labelsForDesc []string
			if mc.Name == ProcessInfo {
				labelsForDesc = append(labelsForDesc, ic.labels...)
				labelsForDesc = append(labelsForDesc, LabelProcessPid)
				labelsForDesc = append(labelsForDesc, LabelProcessName)
			} else {
				labelsForDesc = ic.labels
			}
			desc := prometheus.NewDesc(mc.Name, mc.Help, labelsForDesc, nil)
			ic.resources[mc.Name] = desc
			ch <- desc

			logger.IluvatarLog.Infof("Register gpu resource '%s'", mc.Name)
		}
	} else {
		ic.ctx.cancel()
		ic.ctx = nil
		for key, _ := range ic.resources {
			delete(ic.resources, key)
			logger.IluvatarLog.Infof("Unregister gpu resource '%s'", string(key))
		}
	}
}

// Collect is the implementation of the interface of 'prometheus.Collector.Collect()', once
// there is a request from client, it will be triggered, then collect the metrics.
func (ic *iluvatarCollector) Collect(ch chan<- prometheus.Metric) {
	logger.IluvatarLog.Info("Collect() called...")
	collectMetrics := func(ch chan<- prometheus.Metric) {
		metrics := ic.ctx.getMetrics()
		for _, ms := range metrics {
			for _, m := range ms {
				labelForValues := make([]string, len(ic.labels))
				for i, label := range ic.labels {
					labelForValues[i] = m.labels[label]
				}
				if m.name == ProcessInfo {
					labelForValues = append(labelForValues, m.labels[LabelProcessPid])
					labelForValues = append(labelForValues, m.labels[LabelProcessName])
				}
				if desc, ok := ic.resources[m.name]; ok {
					ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, m.value, labelForValues...)
				}
			}
		}
	}

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer close(done)
		collectMetrics(ch)
	}()

	select {
	case <-time.After(25 * time.Second):
		logger.IluvatarLog.Errorf("Collect metrics timeout")
		return
	case <-done:
		logger.IluvatarLog.Infof("Task completed within the timeout period.")
	}

	logger.IluvatarLog.Infof("Collect metrics took %v", time.Since(start))
}
