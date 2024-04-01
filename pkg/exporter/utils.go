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

import (
	"io/ioutil"
	"os"
	"reflect"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

func load(c *Config, path string) error {
	f, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		glog.Errorf("Can not find config file %s", path)
	}
	if err != nil {
		return err
	}

	return yaml.Unmarshal(f, c)
}

func getConfig(path string) Config {
	var conf Config
	if err := load(&conf, path); err != nil {
		glog.Fatalln(err)
	}

	return conf
}

func isInterfaceNil(in interface{}) bool {
	defer func() {
		recover()
	}()

	vi := reflect.ValueOf(in)
	return vi.IsNil()
}

func validatePath(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}
