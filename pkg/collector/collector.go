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

	"strconv"
	"time"

	"gitee.com/deep-spark/go-ixml/pkg/ixml"
	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type subCollector interface {
	collect(ctx *ixContext)
}

// IXCollector implements the Collector interface.
type IXCollector struct {
	opts       *Options
	resources  map[string]*prometheus.Desc
	sysinfo    SysInfo
	cfgMetrics []config.MetricItem
	cfgLabels  []string
	ctx        *ixContext
}

func NewIXCollector(opts *Options) (*IXCollector, error) {
	cfg := config.MetricConfig{
		ConfigFile:    opts.MetricsConfig,
		MetricsLabels: &config.MetricsLabels{},
	}
	if err := cfg.ParseMetricConfig(); err != nil {
		logger.IXLog.Errorf("Error parsing config: %s", err)
		return nil, err
	}

	labels, _ := cfg.ParseLabelConfig(opts.EnableK8s)

	return &IXCollector{
		opts:       opts,
		sysinfo:    getDeviceInfo(),
		resources:  make(map[string]*prometheus.Desc),
		cfgMetrics: cfg.MetricsLabels.Metrics,
		cfgLabels:  labels,
		ctx:        nil,
	}, nil
}

// Describe is the implementation of the interface of 'prometheus.Collecter.Describe()', once
// 'prometheus.MustRegtister()' or 'prometheus.Unregister()' was called, it will be triggered.
func (ic *IXCollector) Describe(ch chan<- *prometheus.Desc) {
	logger.IXLog.Info("Describe() called ...")

	if ic.ctx == nil {
		ic.ctx = newContext()
		InitGpuCollector(ic)
		InitKubeCollector(ic)

		for _, mc := range ic.cfgMetrics {
			var labelsForDesc []string
			if mc.Name == IxProcessInfo {
				labelsForDesc = append(labelsForDesc, ic.cfgLabels...)
				labelsForDesc = append(labelsForDesc, LabelProcessPid)
				labelsForDesc = append(labelsForDesc, LabelProcessName)
			} else {
				labelsForDesc = ic.cfgLabels
			}

			desc := prometheus.NewDesc(mc.Name, mc.Help, labelsForDesc, nil)
			ic.resources[mc.Name] = desc
			ch <- desc
			logger.IXLog.Infof("Register gpu resource '%s' to prometheus.", mc.Name)
		}
	} else {
		ic.ctx.cancel()
		ic.ctx = nil
		for key := range ic.resources {
			delete(ic.resources, key)
		}
	}
}

// Collect is the implementation of the interface of 'prometheus.Collector.Collect()', once
// there is a request from client, it will be triggered, then collect the metrics.
func (ic *IXCollector) Collect(ch chan<- prometheus.Metric) {
	logger.IXLog.Info("Collect() called...")

	start := time.Now()

	metricss := ic.ctx.getMetrics()
	for _, ms := range metricss {
		for _, m := range ms {
			labelForValues := make([]string, len(ic.cfgLabels))
			for i, label := range ic.cfgLabels {
				labelForValues[i] = m.labels[label]
			}
			if m.name == IxProcessInfo {
				labelForValues = append(labelForValues, m.labels[LabelProcessPid])
				labelForValues = append(labelForValues, m.labels[LabelProcessName])
			}
			if desc, ok := ic.resources[m.name]; ok {
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, m.value, labelForValues...)
			}
		}
	}
	logger.IXLog.Infof("Collect metrics took %v", time.Since(start))
}

func collectDriverVers(info *SysInfo) error {
	var ret ixml.Return

	info.GPUCount, ret = ixml.DeviceGetCount()
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get device count, ret: %v", ret)
	}
	logger.IXLog.Infof("GPU count: %d", info.GPUCount)

	info.driverVersion, ret = ixml.SystemGetDriverVersion()
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get device version, ret: %v", ret)
	}
	logger.IXLog.Infof("Driver version: %s", info.driverVersion)

	info.ixmlVersion, ret = ixml.SystemGetNVMLVersion()
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get ixml version, ret: %v", ret)
	}
	logger.IXLog.Infof("IXML version: %s", info.ixmlVersion)

	info.cudaVersion, ret = ixml.SystemGetCudaDriverVersion()
	if ret != ixml.SUCCESS {
		logger.IXLog.Warningf("Unable to get cuda driver version: %v", ret)
	} else {
		cudaVersionInt, err := strconv.Atoi(info.cudaVersion)
		if err != nil {
			logger.IXLog.Warningf("Unable to convert cuda version %s: %v", info.cudaVersion, err)
		} else {
			major, minor := uint(cudaVersionInt)/1000, uint(cudaVersionInt)%1000/10
			logger.IXLog.Infof("Cuda version:  %d.%d", major, minor)
		}
	}
	return nil
}

func processDeviceAtIndex(info *SysInfo, index uint, boardChips *[]chip) error {
	var device ixml.Device
	gpu := GpuInfo{
		index: index,
	}

	ret := ixml.DeviceGetHandleByIndex(index, &device)
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get device handle by index, ret: %v", ret)
	}

	name, ret := device.GetName()
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get device name, ret: %v", ret)
	}
	gpu.name = name

	uuid, ret := device.GetUUID()
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get device uuid, ret: %v", ret)
	}
	gpu.uuid = uuid

	serial, ret := device.GetSerial()
	if ret != ixml.SUCCESS {
		return fmt.Errorf("unable to get device serial, ret: %v", ret)
	}
	gpu.serial = serial

	pos, ret := device.GetBoardPosition()
	if ret != ixml.SUCCESS {
		if ret == ixml.ERROR_NOT_SUPPORTED {
			logger.IXLog.Infof("GPU %s not support splitboard.", gpu.uuid)
			info.pairChips[uuid] = uuid
		} else {
			return fmt.Errorf("failed to get board position of %s, ret: %v", gpu.uuid, ret)
		}
	} else {
		logger.IXLog.Infof("The board position of %s is: %d", gpu.uuid, pos)
		key := chip{
			uuid:      uuid,
			operation: device,
		}

		*boardChips = append(*boardChips, key)
	}

	info.GPUs[uuid] = gpu
	return nil
}

func collectChipData(info *SysInfo) []chip {
	var boardChips []chip

	for index := uint(0); index < info.GPUCount; index++ {
		if err := processDeviceAtIndex(info, index, &boardChips); err != nil {
			logger.IXLog.Errorf("error to process device at index %d: %v", index, err)
			continue
		}
	}

	return boardChips
}

func getDeviceInfo() SysInfo {
	var info SysInfo
	info.pairChips = make(map[string]string)
	info.GPUs = make(map[string]GpuInfo)

	if err := collectDriverVers(&info); err != nil {
		logger.IXLog.Error(err)
		return info
	}

	boardChips := collectChipData(&info)
	if len(boardChips) == 0 {
		logger.IXLog.Warningln("No chip support splitboard.")
		return info
	}

	chipmap := make(map[string]bool, len(boardChips))
	for _, bc := range boardChips {
		chipmap[bc.uuid] = false
	}

	for i, first := range boardChips {
		if chipmap[first.uuid] {
			continue
		}
		for j := i + 1; j < len(boardChips); j++ {
			second := boardChips[j]
			onSameBoard, ret := ixml.GetOnSameBoard(first.operation, boardChips[j].operation)
			if ret != ixml.SUCCESS {
				if ret == ixml.ERROR_NOT_SUPPORTED {
					logger.IXLog.Warningf("GetOnSameBoard: Not supported\n")
				} else {
					logger.IXLog.Errorf("Unable to get OnSameBoard %v", ret)
				}
				continue
			}
			if onSameBoard == 1 { // 1 means on same board
				chipmap[first.uuid] = true
				chipmap[second.uuid] = true
				info.pairChips[first.uuid] = second.uuid
				info.pairChips[second.uuid] = first.uuid
				logger.IXLog.Infof("%s and %s on same board", first.uuid, second.uuid)
				break
			}
		}
	}

	return info
}
