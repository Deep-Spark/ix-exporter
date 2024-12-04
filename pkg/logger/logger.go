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

package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tsaikd/KDGoLib/logrusutil"
	"gopkg.in/natefinch/lumberjack.v2" // Logrotate
)

var IluvatarLog *LoggerWrapper

func InitIluvatarLog(fileName string, logLevel int64) error {
	IluvatarLog = NewIluvatarLog()
	err := IluvatarLog.UpdateConfig(fileName, logLevel)
	if err != nil {
		return err
	}
	return nil
}

type LoggerWrapper struct {
	*logrus.Logger
	logFile *os.File
}

func NewIluvatarLog() *LoggerWrapper {
	return &LoggerWrapper{
		Logger: logrus.New(),
	}
}

func (ixLog *LoggerWrapper) UpdateConfig(fileName string, logLevel int64) error {

	logWriter := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    1024,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	ixLog.SetLogLevel(logLevel)

	formatter := &logrusutil.ConsoleLogFormatter{
		TimestampFormat: "2006/01/02 15:04:07",
		Flag:            logrusutil.Llevel | logrusutil.Ltime | logrusutil.Lshortfile,
	}
	ixLog.SetFormatter(formatter)

	ixLog.SetOutput(io.MultiWriter(os.Stdout, logWriter))

	return nil
}

func (ixLog *LoggerWrapper) SetLogLevel(logLevel int64) {
	var level logrus.Level
	switch logLevel {
	case 0:
		level = logrus.DebugLevel
	case 1:
		level = logrus.InfoLevel
	case 2:
		level = logrus.WarnLevel
	default:
		level = logrus.ErrorLevel
	}
	ixLog.Logger.SetLevel(level)
}
