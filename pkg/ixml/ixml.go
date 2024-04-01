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

package ixml

// MemoryInfo contains information of a gpu device.
type MemoryInfo struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// PciInfo contains of information of a gpu device.
type PciInfo struct {
	Bus            uint
	BusId          string
	BusIdLegacy    string
	Device         uint
	Domain         uint
	PciDeviceId    uint
	PciSubSystemId uint
}

// PowerLimitConstraints contains of the limitation of a gpu device.
type PowerLimitConstraints struct {
	MaxLimit uint
	MinLimit uint
}

// ClockInfo contains of all clock information of a gpu device.
type ClockInfo struct {
	Mem uint
	Sm  uint
}

// Utilization contains of the percent of gpu and memory used.
type Utilization struct {
	GPU uint
	Mem uint
}

// Device defines the implementation of specified device.
type Device interface {
	// DeviceGetName returns the name of the gpu.
	DeviceGetName() (string, error)

	// DeviceGetUUID returns the uuid of the gpu.
	DeviceGetUUID() (string, error)

	// DeviceGetUUID returns the uuid of the gpu.
	DeviceGetIndex() (uint, error)

	// DeviceGetFanSpeed returns the value of the gpu fan speed.
	DeviceGetFanSpeed() (uint, error)

	// DeviceGetMemoryInfo returns the memory status of the gpu.
	DeviceGetMemoryInfo() (MemoryInfo, error)

	// DeviceGetTemperature returns the current temperature of the gpu.
	DeviceGetTemperature() (uint, error)

	// DeviceGetPciInfo returns the pci information of the gpu.
	DeviceGetPciInfo() (PciInfo, error)

	// DeviceGetPowerUsage returns the current power usage of the gpu.
	DeviceGetPowerUsage() (uint, error)

	// DeviceGetClockInfo returns the sm clock and memory clock of the gpu.
	DeviceGetClockInfo() (ClockInfo, error)

	// DeviceGetUtilization returns gpu and memory used percent.
	DeviceGetUtilization() (Utilization, error)
}

// Init
func Init() error {
	return deviceInit()
}

// GetDeviceCount get the number of gpu.
func GetDeviceCount() (uint, error) {
	return getDeviceCount()
}

// GetDriverVersion get the current driver version.
func GetDriverVersion() (string, error) {
	return getDriverVersion()
}

// GetCudaVersion get which CUDA version is used.
func GetCudaVersion() (string, error) {
	return getCudaVersion()
}

// Shutdown
func Shutdown() error {
	return deviceShutdown()
}

// NewDeviceByIndex creates a device instance by index.
func NewDeviceByIndex(index uint) (Device, error) {
	dev, err := getDeviceByIndex(index)
	if err != nil {
		return nil, err
	}

	return &device{
		handle: dev,
	}, nil
}

func NewDeviceByUUID(uuid string) (Device, error) {
	dev, err := getDeviceByUUID(uuid)
	if err != nil {
		return nil, err
	}

	return &device{
		handle: dev,
	}, nil
}
