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
	"context"
	"fmt"

	"gitee.com/deep-spark/ixexporter/pkg/logger"
	udev "github.com/jochenvg/go-udev"
)

const IxDevSubsystem = "iluvatar-sys"

func StartIxDevWatcher() (<-chan *udev.Device, error) {
	logger.IXLog.Info("Starting IX dev watcher.")
	ctx := context.Background()
	u := udev.Udev{}
	udevMonitor := u.NewMonitorFromNetlink("kernel")
	if udevMonitor == nil {
		return nil, fmt.Errorf("failed to create udev monitor")
	}

	err := udevMonitor.FilterAddMatchSubsystem(IxDevSubsystem)
	if err != nil {
		return nil, fmt.Errorf("failed to add subsystem %s: %v", IxDevSubsystem, err)
	}

	ixdevCh, _, err := udevMonitor.DeviceChan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open ix dev channel: %v", err)
	}
	return ixdevCh, nil
}
