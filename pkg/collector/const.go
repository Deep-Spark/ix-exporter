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

import "time"

const (
	Iluvatar = "iluvatar"

	IxTemperature       = "ix_temperature" // Deprecated
	IxGpuTemperature    = "ix_gpu_temperature"
	IxMemTemperature    = "ix_mem_temperature"
	IxFanSpeed          = "ix_fan_speed"
	IxSmClock           = "ix_sm_clock"
	IxMemClock          = "ix_mem_clock"
	IxMemTotal          = "ix_mem_total"
	IxMemUsed           = "ix_mem_used"
	IxMemFree           = "ix_mem_free"
	IxMemUtilization    = "ix_mem_utilization"
	IxGpuUtilization    = "ix_gpu_utilization"
	IxPowerUsage        = "ix_power_usage"
	IxProcessInfo       = "ix_process_info"
	IxXidErrors         = "ix_xid_errors"
	IxEccSbeVolStatus   = "ix_ecc_sbe_vol_status"
	IxEccDbeVolStatus   = "ix_ecc_dbe_vol_status"
	IxSmUtilization     = "ix_sm_utilization"
	IxSmOccupancy       = "ix_sm_occupancy"
	IxDrawBwUtilization = "ix_dram_bw_util"
	IxPcieRxThroughput  = "ix_pcie_rx_throughput"
	IxPcieTxThroughput  = "ix_pcie_tx_throughput"
	IxPcieReplayCounter = "ix_pcie_replay_counter"
)

const (
	CollectWaitTime    = 10 * time.Second
	CacheExpiryTime    = 2 * time.Second
	KubeletConnTimeout = 5 * time.Second
)

const (
	LabelProcessPid  = "process_pid"
	LabelProcessName = "process_name"
)
