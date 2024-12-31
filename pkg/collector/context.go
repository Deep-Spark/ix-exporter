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
	"sync"

	"gitee.com/deep-spark/ixexporter/pkg/logger"
)

type ixContext struct {
	ctx         context.Context
	labels      []string
	cancelFunc  context.CancelFunc
	signalCh    chan struct{}
	collectors  []subCollector
	metrics     map[string][]metric
	labelValues map[string]labelType
	mutex       sync.Mutex
}

func newContext() *ixContext {
	ctx, cancel := context.WithCancel(context.Background())

	return &ixContext{
		ctx:        ctx,
		cancelFunc: cancel,
		metrics:    make(map[string][]metric),
	}
}

func (ctx *ixContext) cancel() {
	ctx.cancelFunc()
}

func (ctx *ixContext) done() <-chan struct{} {
	return ctx.ctx.Done()
}

func (ctx *ixContext) signal() <-chan struct{} {
	ctx.mutex.Lock()
	if ctx.signalCh == nil {
		ctx.signalCh = make(chan struct{})
	}
	ctx.mutex.Unlock()

	return ctx.signalCh
}

func (ctx *ixContext) registerCollector(collector subCollector) {
	ctx.collectors = append(ctx.collectors, collector)
}

func (ctx *ixContext) getMetrics() map[string][]metric {
	// Notify all collectors to update metrics.
	if ctx.signalCh != nil {
		close(ctx.signalCh)
		ctx.signalCh = nil
	}

	for uuid, ms := range ctx.metrics {
		updateMetrics := []metric{}

		for _, metric := range ms {
			if labels, ok := ctx.labelValues[uuid]; ok {
				for key, value := range labels {
					metric.labels[key] = value
				}
			}
			updateMetrics = append(updateMetrics, metric)
		}
		ctx.metrics[uuid] = updateMetrics
	}

	return ctx.metrics
}

func (ctx *ixContext) updateMetrics(metrics interface{}) {
	// Wait until metrics updated.
	logger.IluvatarLog.Logger.Infof("Start Update metrics...")

	// Store metrics.
	switch metrics := metrics.(type) {
	case map[string][]metric:
		updateMetrics := make(map[string][]metric)
		for uuid, ms := range metrics {
			var ms_ []metric
			for _, m := range ms {
				for _, key := range ctx.labels {
					if _, ok := m.labels[key]; !ok {
						m.labels[key] = ""
					}
				}
				ms_ = append(ms_, m)
			}
			updateMetrics[uuid] = ms_
		}
		ctx.metrics = updateMetrics
	case map[string]labelType:
		ctx.labelValues = metrics
	}
}
