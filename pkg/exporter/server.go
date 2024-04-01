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
	"net/http"
	"strconv"

	"gitee.com/deep-spark/ixexporter/pkg/ixml"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer
type MetricsServer struct {
	opts      *Options
	config    Config
	registry  *prometheus.Registry
	collector prometheus.Collector
}

func NewMetricsServer(opts *Options) *MetricsServer {
	return &MetricsServer{
		opts:     opts,
		config:   getConfig(opts.MetricsConfig),
		registry: prometheus.NewRegistry(),
	}
}

// Run starts a http server.
func (ms *MetricsServer) Run() {
	// start collector
	go ms.startCollector()

	// start http server
	go ms.startHttpServer()
}

// Exit server.
func (ms *MetricsServer) Exit() {
	err := ixml.Shutdown()
	if err != nil {
		glog.Errorln(err)
	}

	ms.registry.Unregister(ms.collector)
}

func (ms *MetricsServer) startHttpServer() {
	http.Handle(ms.opts.MetricsRouter, promhttp.HandlerFor(ms.registry, promhttp.HandlerOpts{}))
	http.ListenAndServe(":"+strconv.FormatUint(uint64(ms.opts.Port), 10), nil)
}

func (ms *MetricsServer) startCollector() {
	info := ms.getDeviceInfo()
	ms.collector = newCollector(ms.config, info)

	ms.registry.MustRegister(ms.collector)
	glog.Infof("Register collector.\n")
}

func (ms *MetricsServer) getDeviceInfo() iluvatarGPU {
	var err error
	var info iluvatarGPU
	info.gpus = make(map[string]gpuInfo)

	err = ixml.Init()
	if err != nil {
		glog.Fatalf("%v", err)
	}

	info.count, err = ixml.GetDeviceCount()
	if err != nil {
		glog.Errorln(err)
	}

	info.driverVersion, err = ixml.GetDriverVersion()
	if err != nil {
		glog.Errorln(err)
		return iluvatarGPU{}
	}

	info.cudaVersion, err = ixml.GetCudaVersion()
	if err != nil {
		glog.Errorln(err)
	}

	for index := uint(0); index < info.count; index++ {
		var uuid string
		gpu := gpuInfo{index: index}

		device, err := ixml.NewDeviceByIndex(index)
		if err != nil {
			glog.Errorln(err)
			return iluvatarGPU{}
		}

		gpu.name, err = device.DeviceGetName()
		if err != nil {
			glog.Errorln(err)
		}

		uuid, err = device.DeviceGetUUID()
		if err != nil {
			glog.Errorln(err)
		}

		info.gpus[uuid] = gpu
	}

	return info
}
