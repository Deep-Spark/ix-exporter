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

const (
	// DefaultMetricsConfig
	DefaultMetricsConfig = "/etc/ixexporter/exporter.yaml"

	// DefaultMetricsRouter
	DefaultMetricsRouter = "/metrics"

	// DefaultAddress
	DefaultAddress = "localhost"

	// DefaultPort
	DefaultPort = 32021
)

type resourceType string

type labelType map[string]string

type Config struct {
	Port             uint64   `yaml:"port"`
	EnableKubernetes bool     `yaml:"enableKubernetes"`
	Metrics          []Metric `yaml:"metrics"`
}

type Metric struct {
	Name resourceType `yaml:"name"`
	Help string       `yaml:"help"`
}

type Options struct {
	MetricsConfig string
	MetricsRouter string
	Address       string
	Port          uint
}

type iluvatarGPU struct {
	count         uint
	driverVersion string
	cudaVersion   string
	gpus          map[string]gpuInfo
}

type gpuInfo struct {
	name  string
	index uint
}

type metric struct {
	name   resourceType
	value  float64
	labels labelType
}

type gpuPod struct {
	name      string
	container string
	namespace string
}
