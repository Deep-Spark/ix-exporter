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
	"context"
	"sync"
)

type ixContext struct {
	wg          *sync.WaitGroup
	ctx         context.Context
	cancelFunc  context.CancelFunc
	metrics     map[string][]metric
	labelValues map[string]labelType
	labels      []string
	collectors  []subCollector
	signalCh    chan struct{}
	mutex       sync.Mutex
}

func newContext() *ixContext {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())

	return &ixContext{
		wg:         wg,
		ctx:        ctx,
		cancelFunc: cancel,
		metrics:    make(map[string][]metric),
	}
}

func (ctx *ixContext) registerCollector(collector subCollector) {
	ctx.collectors = append(ctx.collectors, collector)
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

func (ctx *ixContext) updateMetrics(metrics interface{}) {
	ctx.mutex.Lock()

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

	ctx.mutex.Unlock()

	ctx.wg.Done()
}

func (ctx *ixContext) getMetrics() map[string][]metric {
	for range ctx.collectors {
		ctx.wg.Add(1)
	}

	// Notify all collectors to update metrics.
	ctx.mutex.Lock()
	if ctx.signalCh != nil {
		close(ctx.signalCh)
		ctx.signalCh = nil
	}
	ctx.mutex.Unlock()

	// Wait until metrics updated.
	ctx.wg.Wait()

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

func (ctx *ixContext) updateLabels(labels []string) {
	ctx.mutex.Lock()

	ctx.labels = append(ctx.labels, labels...)

	ctx.mutex.Unlock()

	ctx.wg.Done()
}

func (ctx *ixContext) getLabels() []string {
	if len(ctx.labels) == 0 {
		for range ctx.collectors {
			ctx.wg.Add(1)
		}

		// Wait until all labels updated.
		ctx.wg.Wait()
	}

	return ctx.labels
}
