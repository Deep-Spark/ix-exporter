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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2" // Logrotate
)

var IXLog *LoggerWrapper
var logLevels = [...]string{"PANIC", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

func init() {
	IXLog = &LoggerWrapper{
		Logger: logrus.New(),
	}
}

func InitIXLog(fileName string, logLevel int) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
			return err
		}
	}
	err := IXLog.UpdateConfig(fileName, logLevel)
	if err != nil {
		return err
	}
	return nil
}

type LoggerWrapper struct {
	*logrus.Logger
}

func (ixLog *LoggerWrapper) UpdateConfig(fileName string, logLevel int) error {

	logWriter := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	var entryLevels = []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
	ixLog.Logger.SetLevel(entryLevels[logLevel])

	// Log the messages to stdout as well as logWriter
	ixLog.SetOutput(io.MultiWriter(os.Stdout, logWriter))

	// Required to get line number
	ixLog.SetReportCaller(true)

	// Set custom formatter
	ixLog.SetFormatter(&MyFormatter{})

	return nil
}

type NullFormatter struct {
}

func (f *NullFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(""), nil
}

type MyFormatter struct {
}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	level := logLevels[int(entry.Level)]
	strList := strings.Split(entry.Caller.File, "/")
	fileName := strList[len(strList)-1]

	b.WriteString(fmt.Sprintf("%s %5s [%s:%d] - %s\n",
		entry.Time.Format("2006/01/02 15:04:05.000"),
		level, fileName, entry.Caller.Line, entry.Message))
	return b.Bytes(), nil
}
