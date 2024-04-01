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

package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"gitee.com/deep-spark/ixexporter/pkg/exporter"
	"github.com/golang/glog"
)

func ParseOptions(opts *exporter.Options) {
	flag.StringVar(&opts.MetricsConfig, "c", exporter.DefaultMetricsConfig, "Metrics config file which contains of all fields.")
	flag.StringVar(&opts.MetricsRouter, "r", exporter.DefaultMetricsRouter, "Metrics router.")
	flag.StringVar(&opts.Address, "a", exporter.DefaultAddress, "Metrics config file which contains of all fields.")
	flag.UintVar(&opts.Port, "p", exporter.DefaultPort, "Service port.")
}

func main() {
	opts := exporter.Options{}

	ParseOptions(&opts)
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	flag.Lookup("stderrthreshold").Value.Set("INFO")
	defer glog.Flush()

	ms := exporter.NewMetricsServer(&opts)
	ms.Run()

	sigChn := make(chan os.Signal, 1)
	defer close(sigChn)

	signal.Notify(sigChn, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case s := <-sigChn:
		glog.Infof("Get signal: %v", s)

		ms.Exit()
	}
}
