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

package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"gitee.com/deep-spark/ixexporter/pkg/collector"
	"gitee.com/deep-spark/ixexporter/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsServer
type MetricsServer struct {
	sync.Mutex
	server *http.Server
}

func NewMetricsServer(opts *collector.Options, reg *prometheus.Registry) *MetricsServer {

	mServer := &MetricsServer{
		server: &http.Server{
			Addr:           opts.IP + ":" + opts.Port,
			Handler:        http.DefaultServeMux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		},
	}

	http.Handle("/", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`<html>
				<head><title>Iluvatar-Exporter</title></head>
				<body>
				<h1>Iluvatar-Exporter</h1>
				<p><a href="./metrics">Metrics</a></p>
				</body>
				</html>
			`))
			if err != nil {
				logger.IluvatarLog.Errorf("Write response error: %v", err)
			}
		},
	))

	http.Handle("/metrics", promhttp.HandlerFor(reg,
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		}),
	)

	return mServer
}

func (ms *MetricsServer) Run(ctx context.Context, cancel context.CancelFunc) {
	logger.IluvatarLog.Infof("Metrics server is running on %s", ms.server.Addr)
	go func() {
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.IluvatarLog.Errorf("Metrics server error: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	// Disable keep-alives to ensure all connections are closed promptly
	ms.server.SetKeepAlivesEnabled(false)
	ms.serverShutdown()
}

func (ms *MetricsServer) serverShutdown() {
	logger.IluvatarLog.Infof("Metrics server is shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := ms.server.Shutdown(ctx); err != nil {
		logger.IluvatarLog.Errorf("Metrics server shutdown error: %v", err)
	} else {
		logger.IluvatarLog.Infof("Metrics server stopped")
	}
}
