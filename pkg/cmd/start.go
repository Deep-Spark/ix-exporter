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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"gitee.com/deep-spark/go-ixdcgm/pkg/ixdcgm"
	"gitee.com/deep-spark/go-ixml/pkg/ixml"
	"gitee.com/deep-spark/ixexporter/pkg/collector"
	"gitee.com/deep-spark/ixexporter/pkg/informer"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/reset"
	"gitee.com/deep-spark/ixexporter/pkg/server"
)

func start(c *cli.Context) (err error) {
	opts, err := configToOpts(c)
	if err != nil {
		return fmt.Errorf("failed to parse config to options: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer close(sig)

	opts.IxDevCh, err = collector.StartIxDevWatcher()
	if err != nil {
		return fmt.Errorf("failed to start IX dev watcher: %v", err)
	}

	resetCh := make(chan reset.ResetEvent, 1)
	var resetInformer reset.GpuResetInformer

	if opts.EnableK8s {
		if err := informer.StartIxInformer(opts, sig); err != nil {
			return fmt.Errorf("failed to start ix-config informer: %v", err)
		}
		logger.IXLog.Info("ix-config informer started successfully!")

		time.Sleep(500 * time.Millisecond)
		resetInformer = reset.StartResetInformer(opts, sig, resetCh)
		if resetInformer == nil {
			return fmt.Errorf("failed to start reset informer")
		}
		logger.IXLog.Info("reset informer started successfully!")
	}

	defer func() {
		if resetInformer != nil {
			resetInformer.Stop()
		}
	}()

	for {
		restart, err := startMetricsServer(opts, sig)
		if err != nil {
			return fmt.Errorf("failed to start metrics server: %v", err)
		}
		if !restart {
			break
		}
		logger.IXLog.Info("Preparing to restart ix-exporter.")

		if len(resetCh) > 0 {
			logger.IXLog.Warn("restart signal triggered by reset informer, handling reset event.")
			reset.ShutdownFlag <- struct{}{}
		} else {
			logger.IXLog.Warn("restart signal triggered, restarting metrics server.")
			continue
		}

	resetLoop:
		for {
			time.Sleep(200 * time.Millisecond)
			logger.IXLog.Debug("Waiting for GPU reset event ...")
			switch v := <-resetCh; v {
			case reset.ResetToFalse:
				logger.IXLog.Info("GPU reset state changed to false, restarting metrics server.")
				break resetLoop
			case reset.ResetToTrue:
				logger.IXLog.Info("GPU reset state changed to true, waiting for reset to complete.")
				continue resetLoop
			default:
				return fmt.Errorf("invalid reset event: %v", v)
			}
		}
	}

	return nil
}

func startMetricsServer(opts *collector.Options, sig chan os.Signal) (restart bool, err error) {
	ret := ixml.Init()
	if ret != ixml.SUCCESS {
		return false, fmt.Errorf("failed to init IXML: %v", ret)
	}
	logger.IXLog.Info("IXML successfully initialized!")

	cleanupIxDCGM, err := initIxDCGM(opts)
	if err != nil {
		return false, fmt.Errorf("failed to init IxDCGM: %v", err)
	}
	logger.IXLog.Info("IxDCGM successfully initialized!")

	// Create a new context for this run of the exporter
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reg := prometheus.NewRegistry()
	ixCollector, err := collector.NewIXCollector(opts)
	if err != nil {
		logger.IXLog.Errorf("Failed to create iluvatar collector: %v", err)
		return false, err
	}
	reg.MustRegister(ixCollector)

	defer func() {
		logger.IXLog.Infof("Shutting down the metrics server gracefully...")
		reg.Unregister(ixCollector)
		ixml.Shutdown()
		cleanupIxDCGM()
		time.Sleep(200 * time.Millisecond)
	}()

	mServer := server.NewMetricsServer(opts, reg)
	go func() {
		mServer.Run(ctx, cancel)
	}()

	for {
		select {
		case s := <-sig:
			logger.IXLog.Infof("Received signal: %s", s.String())
			if s != syscall.SIGHUP {
				return false, nil // Shutdown the ix-exporter
			} else {
				return true, nil // Restart the ix-exporter
			}

		case d := <-opts.IxDevCh:
			action := d.Action()
			logger.IXLog.Infof("Received udev action: %v, udev name: %v", action, d.Sysname())
			switch action {
			case "add", "remove":
				return true, nil // Restart the ix-exporter
			default:
				continue
			}
		}
	}

}

func configToOpts(c *cli.Context) (*collector.Options, error) {
	option := &collector.Options{
		Loglevel:      LoglevelFlag,
		Logfile:       LogfileFlag,
		IP:            IPFlag,
		Port:          PortFlag,
		MetricsConfig: MetricsConfigFlag,
		UseRemoteHE:   c.IsSet(CLIRemoteHostEngine),
		RemoteHEInfo:  c.String(CLIRemoteHostEngine),
		EnableK8s:     EnableK8sFlag,
	}

	if EnableK8sFlag {
		option.Namespace = os.Getenv("POD_NAMESPACE")
		if option.Namespace == "" {
			return nil, fmt.Errorf("failed to get POD_NAMESPACE environment variable")
		}
		option.NodeName = os.Getenv("NODE_NAME")
		if option.NodeName == "" {
			return nil, fmt.Errorf("failed to get NODE_NAME environment variable")
		}

		kConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(kConfig)
		if err != nil {
			return nil, err
		}
		option.ClientSet = clientset
	}

	return option, nil
}

func initIxDCGM(ops *collector.Options) (func(), error) {
	if ops.UseRemoteHE {
		logger.IXLog.Info("Attempting to connect to remote ix-hostengine at ", ops.RemoteHEInfo)
		cleanup, err := ixdcgm.Init(ixdcgm.Standalone, ops.RemoteHEInfo, "0")
		if err != nil {
			logger.IXLog.Errorf("Failed to connect to remote ix-hostengine: %v", err)
			cleanup()
			return nil, err
		}
		return cleanup, nil
	} else {
		logger.IXLog.Info("Initializing ixdcgm in embedded mode")
		cleanup, err := ixdcgm.Init(ixdcgm.Embedded)
		if err != nil {
			logger.IXLog.Errorf("Failed to initialize ixdcgm in embedded mode: %v", err)
			cleanup()
			return nil, err
		}
		return cleanup, nil
	}
}
