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
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

const (
	socket               = "/var/lib/kubelet/pod-resources/kubelet.sock"
	iluvatarResourceName = "iluvatar.ai/gpu"
)

type kubeCollector struct {
	gpus    iluvatarGPU
	once    sync.Once
	conn    *grpc.ClientConn
	timeout time.Duration
}

func (kc *kubeCollector) collect(ctx *ixContext) {
	var labels []string

	kc.once.Do(func() {
		var err error

		ret := validatePath(socket)
		if !ret {
			ctx.updateLabels(labels)

			glog.Errorf("Failed to find '%s'\n", socket)
			return
		}

		kc.conn, err = kc.connectToKubelet(socket)
		if err != nil {
			ctx.updateLabels(labels)

			glog.Errorln(err)
			return
		}
	})

	labels = []string{"container", "pod", "namespace"}
	ctx.updateLabels(labels)

	for {
		select {
		case <-ctx.done():
			// Close the gRPC connection.
			glog.Infoln("Disconnect to kubelet")
			kc.conn.Close()
			return
		case <-ctx.signal():
			kc.collectMetrics(ctx)
		}
	}
}

func (kc *kubeCollector) collectMetrics(ctx *ixContext) {
	labels := make(map[string]labelType)

	pods, err := kc.listPods()
	if err != nil {
		glog.Errorln(err)
	} else {
		gpuPods := kc.filterGpuPods(pods, kc.gpus.gpus)
		for uuid, pod := range gpuPods {
			labels[uuid] = labelType{
				"container": pod.container,
				"pod":       pod.name,
				"namespace": pod.namespace,
			}
		}
	}

	ctx.updateMetrics(labels)
}

func (kc *kubeCollector) filterGpuPods(pods *podresourcesapi.ListPodResourcesResponse, gpus map[string]gpuInfo) map[string]gpuPod {
	gpuPods := make(map[string]gpuPod)

	for _, pod := range pods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {
				resourceName := device.GetResourceName()
				if resourceName != iluvatarResourceName {
					continue
				}

				info := gpuPod{
					name:      pod.GetName(),
					namespace: pod.GetNamespace(),
					container: container.GetName(),
				}

				for _, uuid := range device.GetDeviceIds() {
					gpuPods[uuid] = info
				}
			}
		}
	}

	return gpuPods
}

func (kc *kubeCollector) connectToKubelet(socket string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), kc.timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to %s: %v", socket, err)
	}

	return conn, nil
}

func (kc *kubeCollector) listPods() (*podresourcesapi.ListPodResourcesResponse, error) {
	client := podresourcesapi.NewPodResourcesListerClient(kc.conn)

	ctx, cancel := context.WithTimeout(context.Background(), kc.timeout)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("Failed to pod resources %v", err)
	}

	return resp, nil
}

func registerKubeCollector(ctx *ixContext, gpuInfo iluvatarGPU) {
	var collector subCollector

	collector = &kubeCollector{
		gpus:    gpuInfo,
		conn:    nil,
		timeout: 10 * time.Second,
	}
	ctx.registerCollector(collector)

	go collector.collect(ctx)
}
