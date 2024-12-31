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

type MetricConfig struct {
	Name string `yaml:"name"`
	Help string `yaml:"help"`
}

type ExporterConfig struct {
	Metrics []MetricConfig `yaml:"metrics"`
}

type Config struct {
	ConfigFile string
	IxExporter map[string]ExporterConfig
}

func (c *Config) ParseConfig() error {
	exists, err := utils.CheckFileExists(c.ConfigFile)
	if err != nil {
		return err
	}
	if !exists {
		logger.IluvatarLog.Errorf("file not found: %s", c.ConfigFile)
		return err
	}

	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		logger.IluvatarLog.Errorf("fail to open file: %s", c.ConfigFile)
		return err
	}

	if err = yaml.Unmarshal(data, c.IxExporter); err != nil {
		logger.IluvatarLog.Errorf("fail to parse config file: %s", c.ConfigFile)
		return err
	}

	if err = c.verifyIxExporterConfig(); err != nil {
		logger.IluvatarLog.Errorf("verify config failed: %s", c.ConfigFile)
		return err
	}

	return nil
}

func (c *Config) verifyIxExporterConfig() error {
	for k, v := range c.IxExporter {
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
