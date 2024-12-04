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
	"io"
	"regexp"

	"gopkg.in/yaml.v2"
)

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
	Version string  `json:"version"             yaml:"version"`
	Flags   Flags   `json:"flags,omitempty"     yaml:"flags,omitempty"`
	Sharing Sharing `json:"sharing,omitempty"   yaml:"sharing,omitempty"`
}

func ParseConfigFrom(reader io.Reader) (*ClusterConfig, error) {
	var err error
	var configYaml []byte

	configYaml, err = io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read error: %v", err)
	}

	var ccfg ClusterConfig
	err = yaml.Unmarshal(configYaml, &ccfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}

	return &ccfg, nil
}

func RemoveDeviceIduffix(s string) string {
	re := regexp.MustCompile(`(::\d+)$`)
	return re.ReplaceAllString(s, "")
}
