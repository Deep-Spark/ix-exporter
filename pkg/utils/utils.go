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

package utils

import (
	"bufio"
	"os"
	"strings"
)

func CheckFileExists(path string) (bool, error) {
	if path == "" {
		return false, nil
	}

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ValidatePath(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func IsDocker() bool {
	file, err := os.Open("proc/1/cgroup")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for strings.Contains(line, "docker") || strings.Contains(line, "kubepods") {
			return true
		}
	}
	return false
}
