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

const (
	Iluvatar = "iluvatar"

	Temperature     = "ix_temperature"
	FanSpeed        = "ix_fan_speed"
	SmClock         = "ix_sm_clock"
	MemClock        = "ix_mem_clock"
	MemTotal        = "ix_mem_total"
	MemUsed         = "ix_mem_used"
	MemFree         = "ix_mem_free"
	MemUtilization  = "ix_mem_utilization"
	GpuUtilization  = "ix_gpu_utilization"
	PowerUsage      = "ix_power_usage"
	ProcessInfo     = "ix_process_info"
	XidErrors       = "ix_xid_errors"
	EccSbeVolStatus = "ix_ecc_sbe_vol_status"
	EccDbeVolStatus = "ix_ecc_dbe_vol_status"
	SmUtilization   = "ix_sm_utilization"
)

const (
	LabelGPU         = "gpu"
	LabelName        = "name"
	LabelUuid        = "uuid"
	LabelNamespace   = "namespace"
	LabelPod         = "pod"
	LabelContainer   = "container"
	LabelNodeName    = "node_name"
	LabelProcessPid  = "process_pid"
	LabelProcessName = "process_name"
)

var LabelList = []string{
	LabelGPU,
	LabelName,
	LabelUuid,
}

var LabelAllList = []string{
	LabelGPU,
	LabelName,
	LabelUuid,
	LabelNamespace,
	LabelPod,
	LabelContainer,
	LabelNodeName,
}
