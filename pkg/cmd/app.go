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
	"errors"
	"net"
	"strconv"

	"github.com/urfave/cli/v2"

	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
)

const (
	CLILogLevel         = "log-level"
	CLILogFile          = "log-file"
	CLIEnableKubernetes = "enable-kubernetes"
	CLIMetricConfig     = "metrics-config"
	CLIRemoteHostEngine = "remote-ix-hostengine"
	CLIServiceIP        = "ip"
	CLIServicePort      = "port"
	CLIResourceName     = "resource-name"
)

var (
	LoglevelFlag      = 4
	LogfileFlag       string
	IPFlag            string
	PortFlag          string
	MetricsConfigFlag string
	EnableK8sFlag     bool
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
			Destination: &EnableK8sFlag,
			EnvVars:     []string{"IX_EXPORTER_ENABLE_KUBERNETES"},
		},
		&cli.StringFlag{
			Name:        CLIMetricConfig,
			Aliases:     []string{"c"},
			Value:       "/opt/ix-exporter/metrics.yaml",
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
		&cli.StringFlag{
			Name:        CLIResourceName,
			Value:       "iluvatar.com/gpu",
			Usage:       "Resource name of gpu in kubernetes.",
			Destination: &config.ResourceName,
			EnvVars:     []string{"IX_EXPORTER_RESOURCE_NAME"},
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
