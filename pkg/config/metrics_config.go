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

type MetricItem struct {
	Name string `yaml:"name"`
	Help string `yaml:"help"`
}

type MetricItemList struct {
	Metrics []MetricItem `yaml:"metrics"`
}

type MetricConfig struct {
	ConfigFile string
	MetricsMap map[string]MetricItemList
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

	if err = yaml.Unmarshal(data, c.MetricsMap); err != nil {
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
	for k, v := range c.MetricsMap {
		if k == "" || len(v.Metrics) == 0 {
			return errors.New("miss field 'name' or 'metrics' in config file")
		}
		for i, metric := range v.Metrics {
			if metric.Name == "" {
				return errors.New("miss field 'name' in 'metrics' configuration of metrics" + strconv.Itoa(i))
			}
			if metric.Help == "" {
				return errors.New("miss field 'help' in 'metrics' configuration of metrics" + strconv.Itoa(i))
			}
		}
	}

	return nil
}
