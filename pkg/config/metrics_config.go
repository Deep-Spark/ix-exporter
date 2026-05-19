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

package config

import (
	"errors"
	"os"

	"strconv"

	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/utils"
	yaml "gopkg.in/yaml.v2"
)

var (
	LabelDriver    string = "driver"
	LabelIxml             = "ixml"
	LabelGpuIdx           = "gpu"
	LabelGpuName          = "name"
	LabelGpuUuid          = "uuid"
	LabelGpuSerial        = "serial"
	LabelNamespace        = "namespace"
	LabelPod              = "pod"
	LabelContainer        = "container"
	LabelNodeName         = "node_name"
)

type LabelMap struct {
	Name          string `yaml:"name,omitempty"`
	GpuIndex      string `yaml:"gpu,omitempty"`
	Uuid          string `yaml:"uuid,omitempty"`
	Serial        string `yaml:"serial,omitempty"`
	DriverVersion string `yaml:"driver,omitempty"`
	IxmlVersion   string `yaml:"ixml,omitempty"`
	NodeName      string `yaml:"node_name,omitempty"`
	Namespace     string `yaml:"namespace,omitempty"`
	Pod           string `yaml:"pod,omitempty"`
	Container     string `yaml:"container,omitempty"`
}

type MetricItem struct {
	Name string `yaml:"name"`
	Help string `yaml:"help"`
}

type MetricsLabels struct {
	Metrics  []MetricItem `yaml:"metrics"`
	LabelMap *LabelMap    `yaml:"labels"`
}

type MetricConfig struct {
	ConfigFile    string
	MetricsLabels *MetricsLabels
}

func (c *MetricConfig) ParseMetricConfig() error {
	exists, err := utils.CheckFileExists(c.ConfigFile)
	if err != nil {
		return err
	}
	if !exists {
		logger.IXLog.Errorf("config file not found: %s", c.ConfigFile)
		return err
	}

	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		logger.IXLog.Errorf("fail to read data from config file: %s", c.ConfigFile)
		return err
	}

	if err = yaml.Unmarshal(data, c.MetricsLabels); err != nil {
		logger.IXLog.Errorln("fail to parse config data.")
		return err
	}

	if err = c.verifyConfig(); err != nil {
		logger.IXLog.Errorln("fail to verify config.")
		return err
	}

	return nil
}

func (c *MetricConfig) verifyConfig() error {
	if len(c.MetricsLabels.Metrics) == 0 {
		return errors.New("no metrics configured in config file")
	}

	for i, metric := range c.MetricsLabels.Metrics {
		if metric.Name == "" {
			return errors.New("miss field 'name' in 'metrics' configuration of metrics " + strconv.Itoa(i))
		}
		if metric.Help == "" {
			return errors.New("miss field 'help' in 'metrics' configuration of metrics " + strconv.Itoa(i))
		}
	}

	return nil
}

func (c *MetricConfig) ParseLabelConfig(k8s bool) ([]string, error) {
	labels := make([]string, 0)
	l := c.MetricsLabels.LabelMap
	if l.Name != "" {
		LabelGpuName = l.Name
		labels = append(labels, LabelGpuName)
	}
	if l.GpuIndex != "" {
		LabelGpuIdx = l.GpuIndex
		labels = append(labels, LabelGpuIdx)
	}
	if l.Uuid != "" {
		LabelGpuUuid = l.Uuid
		labels = append(labels, LabelGpuUuid)
	}
	if l.Serial != "" {
		LabelGpuSerial = l.Serial
		labels = append(labels, LabelGpuSerial)
	}
	if l.DriverVersion != "" {
		LabelDriver = l.DriverVersion
		labels = append(labels, LabelDriver)
	}
	if l.IxmlVersion != "" {
		LabelIxml = l.IxmlVersion
		labels = append(labels, LabelIxml)
	}
	if l.NodeName != "" {
		LabelNodeName = l.NodeName
		labels = append(labels, LabelNodeName)
	}
	if l.Namespace != "" && k8s {
		LabelNamespace = l.Namespace
		labels = append(labels, LabelNamespace)
	}
	if l.Pod != "" && k8s {
		LabelPod = l.Pod
		labels = append(labels, LabelPod)
	}
	if l.Container != "" && k8s {
		LabelContainer = l.Container
		labels = append(labels, LabelContainer)
	}
	return labels, nil
}
