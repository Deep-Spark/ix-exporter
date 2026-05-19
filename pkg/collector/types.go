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
	udev "github.com/jochenvg/go-udev"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	Loglevel      int
	Logfile       string
	IP            string
	Port          string
	MetricsConfig string
	UseRemoteHE   bool
	RemoteHEInfo  string
	IxDevCh       <-chan *udev.Device

	EnableK8s bool
	Namespace string
	NodeName  string
	ClientSet kubernetes.Interface
}

type SysInfo struct {
	driverVersion string
	ixmlVersion   string
	cudaVersion   string
	GPUCount      uint
	GPUs          map[string]GpuInfo // key is gpu uuid
	pairChips     map[string]string  // key is gpu uuid
}

type GpuInfo struct {
	index  uint
	uuid   string
	name   string
	serial string
}

type Metric struct {
	name   string
	value  float64
	labels LabelsMap
}

type LabelsMap map[string]string // key is label name, vaule is label value
type MetricsMap map[string][]*Metric

type chip struct {
	uuid      string
	operation ixml.Device
}
