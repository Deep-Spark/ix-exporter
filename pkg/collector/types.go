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
	"gitee.com/deep-spark/go-ixml/pkg/ixml"
)

type Options struct {
	Loglevel      int64
	Logfile       string
	IP            string
	Port          string
	MetricsConfig string
	EnableKube    bool
}

type iluvatarGPU struct {
	count         uint
	driverVersion string
	cudaVersion   string
	gpus          map[string]gpuInfo
	pairChips     map[string]string
}

type gpuInfo struct {
	index             uint
	name              string
	temperature       float64
	fanSpeed          float64
	smClock           float64
	memClock          float64
	memoryTotal       float64
	memoryUsed        float64
	memoryFree        float64
	memoryUtilization float64
	gpuUtilization    float64
	powerUsage        float64
	pcieTxThroughput  float64
	pcieRxThroughput  float64
	pcieReplayCount   float64
}

type collectorConfig struct {
	Name string
	Help string
}

type metric struct {
	name   string
	value  float64
	labels map[string]string
}

type labelType map[string]string

type chip struct {
	uuid      string
	operation ixml.Device
}
