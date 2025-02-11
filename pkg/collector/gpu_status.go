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
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"gitee.com/deep-spark/go-ixdcgm/pkg/ixdcgm"
	"gitee.com/deep-spark/go-ixml/pkg/ixml"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/utils"
)

type IXGpuStatus struct {
	ixdcgm.DeviceStatus
	ixdcgm.DeviceProfStatus
}

type ProcessInfo struct {
	*ixml.Info
}

var StatusMetrics = []string{
	IxMemTotal,
	IxMemFree,
	IxMemUsed,
	IxTemperature,
	IxPowerUsage,
	IxFanSpeed,
	IxSmClock,
	IxMemClock,
	IxMemUtilization,
	IxGpuUtilization,
	IxEccSbeVolStatus,
	IxEccDbeVolStatus,
	IxSmUtilization,
	IxSmOccupancy,
	IxDrawBwUtilization,
	IxPcieRxThroughput,
	IxPcieTxThroughput,
	IxPcieReplayCounter,
}

type gpuStatusGather struct {
	gpuIds    []uint
	Fields    []ixdcgm.Short
	FieldGrp  ixdcgm.FieldGrpHandle
	GpuGrp    ixdcgm.GroupHandle
	StatusMap map[uint]IXGpuStatus
}

const (
	IdxPower int = iota
	IdxGpuTemp
	IdxGpuUtil
	IdxMemUtil
	IdxSmClock
	IdxMemClock
	IdxPcieRxThroughput
	IdxPcieTxThroughput
	IdxPcieReplayCounter
	IdxFanSpeed
	IdxEccSbeVolDev
	IdxEccDbeVolDev
	IdxMemTotal
	IdxMemUsed
	IdxMemFree
	IdxSmActive
	IdxSmOccupancy
	IdxDramActive
	FieldsCount
)

func NewGpuStatusGather(gpuIds []uint) (*gpuStatusGather, error) {
	fields := make([]ixdcgm.Short, FieldsCount)
	fields[IdxPower] = ixdcgm.DCGM_FI_DEV_POWER_USAGE
	fields[IdxGpuTemp] = ixdcgm.DCGM_FI_DEV_GPU_TEMP
	fields[IdxGpuUtil] = ixdcgm.DCGM_FI_DEV_GPU_UTIL
	fields[IdxMemUtil] = ixdcgm.DCGM_FI_DEV_MEM_COPY_UTIL
	fields[IdxSmClock] = ixdcgm.DCGM_FI_DEV_SM_CLOCK
	fields[IdxMemClock] = ixdcgm.DCGM_FI_DEV_MEM_CLOCK
	fields[IdxPcieRxThroughput] = ixdcgm.DCGM_FI_DEV_PCIE_RX_THROUGHPUT
	fields[IdxPcieTxThroughput] = ixdcgm.DCGM_FI_DEV_PCIE_TX_THROUGHPUT
	fields[IdxPcieReplayCounter] = ixdcgm.DCGM_FI_DEV_PCIE_REPLAY_COUNTER
	fields[IdxFanSpeed] = ixdcgm.DCGM_FI_DEV_FAN_SPEED
	fields[IdxEccSbeVolDev] = ixdcgm.DCGM_FI_DEV_ECC_SBE_VOL_DEV
	fields[IdxEccDbeVolDev] = ixdcgm.DCGM_FI_DEV_ECC_DBE_VOL_DEV
	fields[IdxMemTotal] = ixdcgm.DCGM_FI_DEV_FB_TOTAL
	fields[IdxMemUsed] = ixdcgm.DCGM_FI_DEV_FB_USED
	fields[IdxMemFree] = ixdcgm.DCGM_FI_DEV_FB_FREE
	fields[IdxSmActive] = ixdcgm.DCGM_FI_PROF_SM_ACTIVE
	fields[IdxSmOccupancy] = ixdcgm.DCGM_FI_PROF_SM_OCCUPANCY
	fields[IdxDramActive] = ixdcgm.DCGM_FI_PROF_DRAM_ACTIVE

	fieldGrpName := fmt.Sprintf("IXFieldGroup%d", rand.Uint64())
	fieldGrpHdl, err := ixdcgm.FieldGroupCreate(fieldGrpName, fields)
	if err != nil {
		return nil, err
	}

	gpuGrpName := fmt.Sprintf("IXGpuGroup%d", rand.Uint64())
	gpuGrpHdl, err := ixdcgm.WatchFields(gpuIds, fieldGrpHdl, gpuGrpName)
	if err != nil {
		err = ixdcgm.FieldGroupDestroy(fieldGrpHdl)
		return nil, err
	}
	// delay some time before gathering exact lastest prof values
	time.Sleep(2 * time.Second)

	return &gpuStatusGather{
		gpuIds:    gpuIds,
		Fields:    fields,
		FieldGrp:  fieldGrpHdl,
		GpuGrp:    gpuGrpHdl,
		StatusMap: make(map[uint]IXGpuStatus, len(gpuIds)),
	}, nil
}

func (gather *gpuStatusGather) DcgmGroupDestroy() (err error) {
	err = ixdcgm.FieldGroupDestroy(gather.FieldGrp)
	if err != nil {
		return
	}
	err = ixdcgm.DestroyGroup(gather.GpuGrp)
	return
}

func (gather *gpuStatusGather) GetGpuStatus() (map[uint]*IXGpuStatus, error) {
	devStatuses := make(map[uint]*IXGpuStatus, len(gather.gpuIds))

	for _, gpuId := range gather.gpuIds {
		values, err := ixdcgm.GetLatestValuesForFields(gpuId, gather.Fields)
		if err != nil {
			_ = ixdcgm.FieldGroupDestroy(gather.FieldGrp)
			_ = ixdcgm.DestroyGroup(gather.GpuGrp)
			return nil, fmt.Errorf("error while gathering status for dev %d, %v", gpuId, err)
		}

		gs := new(IXGpuStatus)
		gs.Id = gpuId
		gs.Power = ixdcgm.GetFieldValueStr(values[IdxPower], "float64")
		gs.Temperature = ixdcgm.GetFieldValueStr(values[IdxGpuTemp], "int64")
		gs.Utilization = ixdcgm.UtilizationInfo{
			Gpu: values[IdxGpuUtil].Int64(),
			Mem: values[IdxMemUtil].Int64(),
		}
		gs.Clocks = ixdcgm.ClockInfo{
			Sm:  values[IdxSmClock].Int64(),
			Mem: values[IdxMemClock].Int64(),
		}
		gs.PCI = ixdcgm.PCIStatusInfo{
			Rx:            values[IdxPcieRxThroughput].Int64(),
			Tx:            values[IdxPcieTxThroughput].Int64(),
			ReplayCounter: values[IdxPcieReplayCounter].Int64(),
		}
		gs.MemUsage = ixdcgm.MemoryUsage{
			Total: values[IdxMemTotal].Int64(),
			Free:  values[IdxMemFree].Int64(),
			Used:  values[IdxMemUsed].Int64(),
		}
		gs.FanSpeed = ixdcgm.GetFieldValueStr(values[IdxFanSpeed], "int64")
		gs.EccSbeVolDev = ixdcgm.GetFieldValueStr(values[IdxEccSbeVolDev], "int64")
		gs.EccDbeVolDev = ixdcgm.GetFieldValueStr(values[IdxEccDbeVolDev], "int64")

		gs.SmActive = ixdcgm.GetFieldValueStr(values[IdxSmActive], "float64")
		gs.SmOccupancy = ixdcgm.GetFieldValueStr(values[IdxSmOccupancy], "float64")
		gs.DramActive = ixdcgm.GetFieldValueStr(values[IdxDramActive], "float64")

		devStatuses[gpuId] = gs

	}

	return devStatuses, nil
}

func (gc *gpuCollector) getStatusMetrics(gpu *GpuInfo, labels map[string]string) []*Metric {
	logger.IXLog.Debugf("try to get status metrics for dev: %s", gpu.uuid)
	metrics := make([]*Metric, 0)

	devStatus, ok := gc.devStatuses[gpu.index]
	if !ok {
		logger.IXLog.Errorf("failed to get status metrics for dev: %s", gpu.uuid)
		return metrics
	}

	for _, config := range gc.cfgMetrics {
		if slices.Contains(StatusMetrics, config.Name) {
			value := float64(0)
			valueStrFlag := false
			valueStr := "0"

			switch config.Name {
			case IxMemTotal:
				value = float64(devStatus.MemUsage.Total)
			case IxMemFree:
				value = float64(devStatus.MemUsage.Free)
			case IxMemUsed:
				value = float64(devStatus.MemUsage.Used)
			case IxSmClock:
				value = float64(devStatus.Clocks.Sm)
			case IxMemClock:
				value = float64(devStatus.Clocks.Mem)
			case IxGpuUtilization:
				value = float64(devStatus.Utilization.Gpu)
			case IxMemUtilization:
				value = float64(devStatus.Utilization.Mem)
			case IxPcieRxThroughput:
				value = float64(devStatus.PCI.Rx)
			case IxPcieTxThroughput:
				value = float64(devStatus.PCI.Tx)
			case IxPcieReplayCounter:
				value = float64(devStatus.PCI.ReplayCounter)

			case IxTemperature:
				valueStrFlag = true
				valueStr = devStatus.Temperature
			case IxPowerUsage:
				valueStrFlag = true
				valueStr = devStatus.Power
			case IxFanSpeed:
				valueStrFlag = true
				valueStr = devStatus.FanSpeed
			case IxEccDbeVolStatus:
				valueStrFlag = true
				valueStr = devStatus.EccDbeVolDev
			case IxEccSbeVolStatus:
				valueStrFlag = true
				valueStr = devStatus.EccSbeVolDev
			case IxSmUtilization:
				valueStrFlag = true
				valueStr = devStatus.SmActive
			case IxSmOccupancy:
				valueStrFlag = true
				valueStr = devStatus.SmOccupancy
			case IxDrawBwUtilization:
				valueStrFlag = true
				valueStr = devStatus.DramActive

			default:
				logger.IXLog.Errorf("metric %s is not a status metric", config.Name)
				continue
			}

			if valueStrFlag {
				if valueStr == "N/A" {
					logger.IXLog.Warnf("status metric %s is not supported for dev %d", config.Name, gpu.index)
					continue
				} else {
					value, _ = strconv.ParseFloat(valueStr, 64)
				}
			}

			metrics = append(metrics, &Metric{
				name:   config.Name,
				value:  value,
				labels: labels,
			})
		}
	}
	logger.IXLog.Debugf("succeeded to get status metrics for dev %d", gpu.index)
	return metrics
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
		logger.IXLog.Errorf("Error reading cmdline file for pid %d: %v", pid, err)
		return ""
	}
	data = bytes.ReplaceAll(data, []byte{0}, []byte{' '})
	return strings.TrimSpace(strings.TrimSuffix(string(data), "\x00"))
}

func collectXidErrors(device ixml.Device) (uint64, error) {
	clocksThrottleReasons, ret := device.GetCurrentClocksThrottleReasons()
	if ret != ixml.SUCCESS {
		return 0, fmt.Errorf("%v", ret)
	}
	return clocksThrottleReasons, nil
}
