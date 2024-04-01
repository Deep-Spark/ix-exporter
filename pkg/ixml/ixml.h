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

#include <nvml.h>

#define IXML_LIBRARY "libixml.so"

#define IXML_INIT                         "nvmlInit"
#define IXML_SHUTDOWN                     "nvmlShutdown"
#define IXML_DEVICE_GET_COUNT             "nvmlDeviceGetCount"
#define IXML_GET_DRIVER_VERSION           "nvmlSystemGetDriverVersion"
#define IXML_GET_CUDA_DRIVER_VERSION      "nvmlSystemGetCudaDriverVersion"
#define IXML_DEVICE_GET_HANDLE_BY_INDEX   "nvmlDeviceGetHandleByIndex"
#define IXML_DEVICE_GET_HANDLE_BY_UUID    "nvmlDeviceGetHandleByUUID"
#define IXML_DEVICE_GET_NAME              "nvmlDeviceGetName"
#define IXML_DEVICE_GET_UUID              "nvmlDeviceGetUUID"
#define IXML_DEVICE_GET_INDEX             "nvmlDeviceGetIndex"
#define IXML_DEVICE_GET_FAN_SPEED         "nvmlDeviceGetFanSpeed"
#define IXML_DEVICE_GET_MEMORY_INFO       "nvmlDeviceGetMemoryInfo"
#define IXML_DEVICE_GET_TEMPERATURE       "nvmlDeviceGetTemperature"
#define IXML_DEVICE_GET_PCI_INFO          "nvmlDeviceGetPciInfo"
#define IXML_DEVICE_GET_BOARD_POSITION    "ixmlDeviceGetBoardPosition"
#define IXML_DEVICE_GET_POWER_USAGE       "nvmlDeviceGetPowerUsage"
#define IXML_DEVICE_GET_BOARD_POWER_USAGE "ixmlDeviceGetBoardPowerUsage"
#define IXML_DEVICE_GET_CLOCK_INFO        "nvmlDeviceGetClockInfo"
#define IXML_DEVICE_GET_UTILIZATION_RATES "nvmlDeviceGetUtilizationRates"

nvmlReturn_t dl_init();
nvmlReturn_t dl_close();
nvmlReturn_t ixmlInit();
nvmlReturn_t ixmlShutdown();
nvmlReturn_t ixmlDeviceGetCount(unsigned int* deviceCount);
nvmlReturn_t ixmlSystemGetDriverVersion(char *version, unsigned int length);
nvmlReturn_t ixmlSystemGetCudaDriverVersion(int *version);
nvmlReturn_t ixmlDeviceGetHandleByIndex(unsigned int index, nvmlDevice_t* device);
nvmlReturn_t ixmlDeviceGetHandleByUUID(const char *uuid, nvmlDevice_t* device);
nvmlReturn_t ixmlDeviceGetName(nvmlDevice_t device, char* name, unsigned int length);
nvmlReturn_t ixmlDeviceGetIndex(nvmlDevice_t device, unsigned int *index);
nvmlReturn_t ixmlDeviceGetUUID(nvmlDevice_t device, char* uuid, unsigned int length);
nvmlReturn_t ixmlDeviceGetFanSpeed(nvmlDevice_t device, unsigned int* speed);
nvmlReturn_t ixmlDeviceGetMemoryInfo(nvmlDevice_t device, nvmlMemory_t* memory);
nvmlReturn_t ixmlDeviceGetTemperature(nvmlDevice_t device, nvmlTemperatureSensors_t sensorType, unsigned int* temp);
nvmlReturn_t ixmlDeviceGetPciInfo(nvmlDevice_t device, nvmlPciInfo_t* pci);
nvmlReturn_t ixmlDeviceGetBoardPosition(nvmlDevice_t device, unsigned int* position);
nvmlReturn_t ixmlDeviceGetPowerUsage(nvmlDevice_t device, unsigned int* power);
nvmlReturn_t ixmlDeviceGetBoardPowerUsage(nvmlDevice_t device, unsigned int* power);
nvmlReturn_t ixmlDeviceGetClockInfo(nvmlDevice_t device, nvmlClockType_t type, unsigned int* clock);
nvmlReturn_t ixmlDeviceGetUtilizationRates(nvmlDevice_t device, nvmlUtilization_t* utilization);
