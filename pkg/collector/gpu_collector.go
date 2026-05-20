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
	"slices"
	"strconv"

	"gitee.com/deep-spark/go-ixdcgm/pkg/ixdcgm"
	"gitee.com/deep-spark/go-ixml/pkg/ixml"

	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
)

type gpuCollector struct {
	sysinfo    SysInfo
	devices    map[string]ixml.Device
	cfgMetrics []config.MetricItem

	devStatuses  map[uint]*IXGpuStatus
	statusGather *gpuStatusGather
}

func InitGpuCollector(ic *IXCollector) {
	logger.IXLog.Infof("Init gpu collector.")

	gc := &gpuCollector{
		sysinfo:     ic.sysinfo,
		cfgMetrics:  ic.cfgMetrics,
		devices:     make(map[string]ixml.Device),
		devStatuses: make(map[uint]*IXGpuStatus),
	}

	gpuIds := make([]uint, 0)
	for uuid, gpuInfo := range gc.sysinfo.GPUs {
		gpuIds = append(gpuIds, gpuInfo.index)
		device, ret := ixml.GetHandleByUUID(uuid)
		if ret != ixml.SUCCESS {
			logger.IXLog.Errorf("Unable to get Handle by uuid %v", ret)
			continue
		}
		gc.devices[uuid] = device
	}

	var err error
	gc.statusGather, err = NewGpuStatusGather(gpuIds)
	if err != nil {
		logger.IXLog.Fatalf("Error while creating gpu status gather: %v", err)
	}

	ic.ctx.registerCollector(gc)
	go gc.collect(ic.ctx)
}

func (gc *gpuCollector) collect(ctx *ixContext) {
	for {
		select {
		case <-ctx.done(): // Triggered by prometheus.Unregister()
			gc.statusGather.DcgmGroupDestroy()
			return
		case <-ctx.signalCh:
			logger.IXLog.Infoln("Start to collect gpu metrics")
			gc.collectMetrics(ctx)
			ctx.updateStat()
			logger.IXLog.Infoln("Gpu collector has updated all metrics.")
		}
	}
}

func (gc *gpuCollector) collectMetrics(ctx *ixContext) {
	var err error
	gc.devStatuses, err = gc.statusGather.GetGpuStatus()
	if err != nil {
		logger.IXLog.Errorf("Error while gathering all gpu status: %v", err)
		return
	}

	metrics := make(MetricsMap, gc.sysinfo.GPUCount)
	for uuid, gpu := range gc.sysinfo.GPUs {
		metrics[uuid] = gc.collectGpuMetrics(ctx, &gpu)
	}

	logger.IXLog.Infof("Store gpu device metrics.")
	for uuid, ms := range metrics {
		ctx.metricss[uuid] = ms
	}
}

func (gc *gpuCollector) collectGpuMetrics(ctx *ixContext, gpu *GpuInfo) []*Metric {

	metrics := make([]*Metric, 0)
	baseLabels := map[string]string{
		config.LabelGpuUuid:   gpu.uuid,
		config.LabelGpuName:   gpu.name,
		config.LabelGpuIdx:    strconv.FormatUint(uint64(gpu.index), 10),
		config.LabelGpuSerial: gpu.serial,
		config.LabelNodeName:  ctx.nodeName,
		config.LabelIxml:      gc.sysinfo.ixmlVersion,
		config.LabelDriver:    gc.sysinfo.driverVersion,
	}
	metrics = append(metrics, gc.getStatusMetrics(gpu, baseLabels)...)

	for _, config := range gc.cfgMetrics {
		if slices.Contains(StatusMetrics, config.Name) {
			continue
		}

		switch config.Name {
		case IxProcessInfo:
			infos, err := ixdcgm.GetDeviceRunningProcesses(gpu.index)
			if err != nil {
				logger.IXLog.Warnf("Failed to get process info for dev %d: %v", gpu.index, err)
				continue
			}
			if len(infos) == 0 {
				pidLabels := make(map[string]string, len(baseLabels)+2)
				for k, v := range baseLabels {
					pidLabels[k] = v
				}
				pidLabels[LabelProcessPid] = ""
				pidLabels[LabelProcessName] = ""
				metrics = append(metrics, &Metric{
					name:   IxProcessInfo,
					value:  0,
					labels: pidLabels,
				})
			}
			for _, info := range infos {
				pidLabels := make(map[string]string, len(baseLabels)+2)
				for k, v := range baseLabels {
					pidLabels[k] = v
				}
				pidLabels[LabelProcessPid] = strconv.FormatUint(uint64(info.Pid), 10)
				pidLabels[LabelProcessName] = getProcessNameByPid(uint32(info.Pid))
				metrics = append(metrics, &Metric{
					name:   IxProcessInfo,
					value:  float64(info.UsedGpuMemory),
					labels: pidLabels,
				})
			}

		default:
			logger.IXLog.Warnf("Unknown metric '%s'", config.Name)
			continue
		}
	}

	return metrics
}
