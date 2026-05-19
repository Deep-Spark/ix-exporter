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
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

const (
	IXConfigMap = "ix-config"
	IXConfigKey = "ix-config"
)

var (
	ResourceName = "iluvatar.com/gpu"
	SplitBoard   = false // default is false, set to true if splitboard is enabled
	ResetGpu     = false // default is false, set to true if support reset gpu
)

type Flags struct {
	SplitBoard bool `json:"splitboard"                yaml:"splitboard"`
	UseVolcano bool `json:"usevolcano,omitempty"      yaml:"usevolcano"`
	ResetGpu   bool `json:"reset_gpu,omitempty"       yaml:"reset_gpu"`
}

// Config is a versioned struct used to hold configuration information.
type ClusterConfig struct {
	ResourceName string `json:"resourceName"        yaml:"resourceName"`
	Flags        Flags  `json:"flags,omitempty"     yaml:"flags,omitempty"`
}

func ParseClusterConfig(yamlStr string) (*ClusterConfig, error) {
	var err error
	yamlBytes := []byte(yamlStr)

	var ccfg ClusterConfig
	err = yaml.Unmarshal(yamlBytes, &ccfg)
	if err != nil {
		return nil, fmt.Errorf("error to unmarshal ix config: %v", err)
	}

	return &ccfg, nil
}
