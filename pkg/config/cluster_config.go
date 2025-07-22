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
	IXResourceName  = "iluvatar.com/gpu"
	IXConfigMap     = "ix-config"
	IXConfigDataKey = "ix-config"
)

var SplitBoard bool = false // default is false, set to true if splitboard is enabled

type Flags struct {
	SplitBoard bool `json:"splitboard"                yaml:"splitboard"`
}

type ReplicatedResources struct {
	Replicas int `json:"replicas"         yaml:"replicas"`
}

// Sharing encapsulates the set of sharing strategies that are supported.
type Sharing struct {
	// TimeSlicing defines the set of replicas to be made for timeSlicing available resources.
	TimeSlicing ReplicatedResources `json:"timeSlicing,omitempty" yaml:"timeSlicing,omitempty"`
	// MPS defines the set of replicas to be shared using MPS
	MPS *ReplicatedResources `json:"mps,omitempty"         yaml:"mps,omitempty"`
}

// Config is a versioned struct used to hold configuration information.
type ClusterConfig struct {
	Version      string  `json:"version,omitempty"   yaml:"version,omitempty"`
	ResourceName string  `json:"resourceName"        yaml:"resourceName"`
	Flags        Flags   `json:"flags,omitempty"     yaml:"flags,omitempty"`
	Sharing      Sharing `json:"sharing,omitempty"   yaml:"sharing,omitempty"`
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
