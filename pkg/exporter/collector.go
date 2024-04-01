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
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

type subCollector interface {
	collect(ctx *ixContext)
}

type collector struct {
	gpus             iluvatarGPU
	enableKubernetes bool
	metrics          []Metric
	resources        map[resourceType]*prometheus.Desc
	labels           []string
	ctx              *ixContext
}

func newCollector(config Config, gpus iluvatarGPU) prometheus.Collector {
	return &collector{
		enableKubernetes: config.EnableKubernetes,
		metrics:          config.Metrics,
		resources:        make(map[resourceType]*prometheus.Desc),
		gpus:             gpus,
		ctx:              nil,
	}
}

// Collect is the implementation of the interface of 'prometheus.Collector.Collect()', once
// there is a request from client, it will be triggered, then collect the metrics.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	// get all metrics
	metrics := c.ctx.getMetrics()

	for _, ms := range metrics {
		for _, m := range ms {
			labelValues := []string{}
			for _, label := range c.labels {
				labelValues = append(labelValues, m.labels[label])
			}
			if desc, ok := c.resources[m.name]; ok {
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, m.value, labelValues...)
			}
		}
	}
}

// Describe is the implementation of the interface of 'prometheus.Collecter.Describe()', once
// 'prometheus.MustRegtister()' or 'prometheus.Unregister()' was called, it will be triggered.
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	if c.ctx == nil {
		c.ctx = newContext()

		registerGpuCollector(c.ctx, c.metrics, c.gpus)
		if c.enableKubernetes {
			registerKubeCollector(c.ctx, c.gpus)
		}

		c.labels = c.ctx.getLabels()

		for _, m := range c.metrics {
			desc := prometheus.NewDesc(string(m.Name), m.Help, c.labels, nil)
			c.resources[m.Name] = desc

			ch <- desc

			glog.Infof("Register gpu resource '%s'\n", m.Name)
		}

	} else {
		c.ctx.cancel()

		for key, _ := range c.resources {
			delete(c.resources, key)
			c.ctx = nil

			glog.Infof("Unregister gpu resource '%s'\n", string(key))
		}
	}
}
