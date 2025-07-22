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
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"

	"gitee.com/deep-spark/go-ixdcgm/pkg/ixdcgm"
	"gitee.com/deep-spark/go-ixml/pkg/ixml"
	"gitee.com/deep-spark/ixexporter/pkg/collector"
	"gitee.com/deep-spark/ixexporter/pkg/informer"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/server"
)

const (
	CLILogLevel         = "log-level"
	CLILogFile          = "log-file"
	CLIEnableKubernetes = "enable-kubernetes"
	CLIMetricConfig     = "metrics-config"
	CLIRemoteHostEngine = "remote-ix-hostengine"
	CLIServiceIP        = "ip"
	CLIServicePort      = "port"
)

var (
	LoglevelFlag      = 4
	LogfileFlag       string
	IPFlag            string
	PortFlag          string
	MetricsConfigFlag string
	EnableKubeFlag    bool
)

const (
	portLeft       = 1024
	portRight      = 65535
	minLogLevel    = -1
	maxLogLevel    = 7
	defaultLogFile = "/var/log/iluvatarcorex/ix-exporter/ix-exporter.log"
)

func NewApp(buildVersion ...string) *cli.App {
	app := cli.NewApp()
	app.Name = "IX Exporter"
	app.Usage = "Generates Iluvatar coreX metrics in the prometheus format"
	app.Before = verifyFlags
	app.Action = func(c *cli.Context) error {
		return start(c)
	}

	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:        CLILogLevel,
			Aliases:     []string{"v"},
			Value:       4,
			Usage:       "Log level, 0-panic, 1-fatal, 2-error, 3-warn, 4-info, 5-debug, 6-trace.",
			Destination: &LoglevelFlag,
			EnvVars:     []string{"IX_EXPORTER_LOGLEVEL"},
		},
		&cli.StringFlag{
			Name:        CLILogFile,
			Aliases:     []string{"f"},
			Value:       defaultLogFile,
			Usage:       "Path of log file.",
			Destination: &LogfileFlag,
			EnvVars:     []string{"IX_EXPORTER_LOGFILE"},
		},
		&cli.BoolFlag{
			Name:        CLIEnableKubernetes,
			Aliases:     []string{"k"},
			Value:       false,
			Usage:       "Enable kubernetes.",
			Destination: &EnableKubeFlag,
			EnvVars:     []string{"IX_EXPORTER_ENABLE_KUBERNETES"},
		},
		&cli.StringFlag{
			Name:        CLIMetricConfig,
			Aliases:     []string{"c"},
			Value:       "/etc/ixexporter/metrics.yaml",
			Usage:       "Path of metrics config file which contains of all fields.",
			Destination: &MetricsConfigFlag,
			EnvVars:     []string{"IX_EXPORTER_METRICS_CONFIG"},
		},
		&cli.StringFlag{
			Name:    CLIRemoteHostEngine,
			Aliases: []string{"r"},
			Usage:   "Connect to remote ix-hostengine at <HOST>:<PORT>. (e.g. localhost:5777)",
			EnvVars: []string{"IX_REMOTE_HOSTENGINE_INFO"},
		},
		&cli.StringFlag{
			Name:        CLIServiceIP,
			Value:       "0.0.0.0",
			Usage:       "Service IP.",
			Destination: &IPFlag,
			EnvVars:     []string{"IX_EXPORTER_SERVICE_IP"},
		},
		&cli.StringFlag{
			Name:        CLIServicePort,
			Aliases:     []string{"p"},
			Value:       "32021",
			Usage:       "Service port.",
			Destination: &PortFlag,
			EnvVars:     []string{"IX_EXPORTER_SERVICE_PORT"},
		},
	}

	return app
}

func verifyFlags(c *cli.Context) error {
	if minLogLevel > LoglevelFlag || LoglevelFlag > maxLogLevel {
		return errors.New("the log level is invalid")
	}
	err := logger.InitIXLog(LogfileFlag, LoglevelFlag)
	if err != nil {
		return err
	}
	logger.IXLog.Infof("cli flag '%s' is set to: %v", CLILogLevel, c.Int(CLILogLevel))
	logger.IXLog.Infof("cli flag '%s' is set to: %v", CLILogFile, c.String(CLILogFile))
	logger.IXLog.Infof("cli flag '%s' is set to: %v", CLIEnableKubernetes, c.Bool(CLIEnableKubernetes))
	logger.IXLog.Infof("cli flag '%s' is set to: %v", CLIMetricConfig, c.String(CLIMetricConfig))

	// Check if remote hostengine is set.
	if c.IsSet(CLIRemoteHostEngine) {
		logger.IXLog.Infof("cli flag '%s' is set to: %v", CLIRemoteHostEngine, c.String(CLIRemoteHostEngine))
	} else {
		logger.IXLog.Infof("cli flag '%s' is unset.", CLIRemoteHostEngine)
	}

	logger.IXLog.Infof("cli flag '%s' is set to: %v", CLIServiceIP, c.String(CLIServiceIP))
	if parseIP := net.ParseIP(IPFlag); parseIP == nil {
		return errors.New("the address is invalid")
	}

	logger.IXLog.Infof("cli flag '%s' is set to: %v", CLIServicePort, c.String(CLIServicePort))
	if portInt, err := strconv.Atoi(PortFlag); err != nil {
		return err
	} else if portInt < portLeft || portInt > portRight {
		return errors.New("the port is invalid")
	}

	logger.IXLog.Infof("All cli flags are verified successfully!")
	return nil
}

func start(c *cli.Context) error {
	opts := configToOpts(c)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer close(sig)

	if opts.EnableKube {
		if err := informer.StartIxInformer(opts, sig); err != nil {
			return fmt.Errorf("failed to start ix informer: %v", err)
		}
		logger.IXLog.Info("ix informer started successfully!")
	}

	for {
		loop, err := startMetricsServer(opts, sig)
		if err != nil {
			return fmt.Errorf("failed to start metrics server: %v", err)
		}
		if loop {
			logger.IXLog.Info("Restarting ix-exporter due to SIGHUP signal.")
			time.Sleep(1 * time.Second) // Sleep before restarting
		} else {
			logger.IXLog.Info("Exiting ix-exporter gracefully.")
			break
		}
	}

	return nil
}

func startMetricsServer(opts *collector.Options, sig chan os.Signal) (loop bool, err error) {
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

	mServer := server.NewMetricsServer(opts, reg)
	go func() {
		mServer.Run(ctx, cancel)
	}()

	s := <-sig
	logger.IXLog.Infof("Received signal: %s", s.String())
	cancel()

	logger.IXLog.Infof("Shutting down ix-exporter gracefully...")
	reg.Unregister(ixCollector)
	ixml.Shutdown()
	cleanupIxDCGM()

	if s != syscall.SIGHUP {
		return false, nil
	}

	return true, nil
}

func configToOpts(c *cli.Context) *collector.Options {
	return &collector.Options{
		Loglevel:      LoglevelFlag,
		Logfile:       LogfileFlag,
		IP:            IPFlag,
		Port:          PortFlag,
		EnableKube:    EnableKubeFlag,
		MetricsConfig: MetricsConfigFlag,
		UseRemoteHE:   c.IsSet(CLIRemoteHostEngine),
		RemoteHEInfo:  c.String(CLIRemoteHostEngine),
	}
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
