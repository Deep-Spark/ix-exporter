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
	"os"
	"sync"
	"time"

	"gitee.com/deep-spark/ixexporter/pkg/logger"
)

type ixContext struct {
	ctx         context.Context
	labels      []string
	cancelFunc  context.CancelFunc
	signalCh    chan struct{}
	collectors  []subCollector
	metricss    MetricsMap
	labelValues map[string]LabelsMap
	nodeName    string

	mutex         sync.Mutex
	procNum       int
	procStartFlag bool
	procEndFlag   bool
	procEndTime   time.Time
	metricColCh   chan struct{} // singal of collect metrics
	metricUptCh   chan struct{} // singal of update metrics
}

func newContext() *ixContext {
	ctx, cancel := context.WithCancel(context.Background())

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		logger.IXLog.Warningf("NODE_NAME was not set in env, get nodeName from hostname.")
		nn, err := os.Hostname()
		if err != nil {
			logger.IXLog.Panicf("Error to get hostname from os.")
		}
		nodeName = nn
	}

	return &ixContext{
		ctx:        ctx,
		cancelFunc: cancel,
		signalCh:   make(chan struct{}),
		metricss:   make(MetricsMap),
		nodeName:   nodeName,

		procNum:       0,
		procStartFlag: false,
		procEndFlag:   true,
		procEndTime:   time.Now().Add(-CacheExpiryTime),
		metricColCh:   make(chan struct{}),
		metricUptCh:   make(chan struct{}),
	}
}

func (ctx *ixContext) cancel() {
	ctx.cancelFunc()
}

func (ctx *ixContext) done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *ixContext) registerCollector(collector subCollector) {
	ctx.collectors = append(ctx.collectors, collector)
}

func (ctx *ixContext) initStat() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	// Notify all collectors to update metrics.
	close(ctx.signalCh)
	ctx.signalCh = make(chan struct{})

	ctx.procNum = 0
	ctx.metricColCh = make(chan struct{})
	ctx.procStartFlag = true
	ctx.procEndFlag = false
}

func (ctx *ixContext) updateStat() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.procNum++
	if ctx.procNum == len(ctx.collectors) {
		close(ctx.metricColCh)
	}
}

func (ctx *ixContext) endStat() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.procEndTime = time.Now()
	ctx.procStartFlag = false
	ctx.procEndFlag = true
}

func (ctx *ixContext) getCachedMetrics() bool {
	if ctx.procEndFlag {
		now := time.Now()
		if now.Before(ctx.procEndTime.Add(CacheExpiryTime)) {
			logger.IXLog.Infoln("Last collect process is finished and cached is fresh, return the cached metrics.")
			return true
		} else {
			logger.IXLog.Infoln("Last collect process is finished but cached is stale, start to collect new metrics.")
			return false
		}
	} else if ctx.procStartFlag {
		logger.IXLog.Infoln("Collect process is running, wait and return the collected metrics.")
		select {
		case <-time.After(CollectWaitTime):
			logger.IXLog.Errorln("Collect process is timeout, start to collect new metrics.")
			return false
		case <-ctx.metricUptCh: // Wait until metrics updated.
			return true
		}
	}

	return false
}

func (ctx *ixContext) getMetrics() MetricsMap {
	if ok := ctx.getCachedMetrics(); ok {
		return ctx.metricss
	}

	ctx.initStat()
	ctx.metricUptCh = make(chan struct{})
	defer close(ctx.metricUptCh)

	select {
	case <-time.After(CollectWaitTime):
		logger.IXLog.Errorln("Not all of metrics are updated within timeout peroid.")
	case <-ctx.metricColCh: // Wait all metrics are updated.
		for uuid, ms := range ctx.metricss {
			updateMetrics := []*Metric{}
			for _, metric := range ms {
				if labels, ok := ctx.labelValues[uuid]; ok {
					for key, value := range labels {
						metric.labels[key] = value
					}
				}
				updateMetrics = append(updateMetrics, metric)
			}
			ctx.metricss[uuid] = updateMetrics
		}
		logger.IXLog.Infoln("All of metrics are updated.")
		ctx.endStat()
	}

	return ctx.metricss
}
