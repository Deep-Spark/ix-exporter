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
	"net"
	"os"
	"sync"
	"time"

	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/utils"
	"google.golang.org/grpc"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

const (
	socket               = "/var/lib/kubelet/pod-resources/kubelet.sock"
	iluvatarResourceName = "iluvatar.com/gpu"
	ConfigFile           = "/iluvatar-config/ix-config"
)

type gpuPod struct {
	name      string
	container string
	namespace string
}

type kubeCollector struct {
	clientset  kubernetes.Interface
	gpus       iluvatarGPU
	once       sync.Once
	conn       *grpc.ClientConn
	timeout    time.Duration
	SplitBoard bool
}

func initClientSet() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.IluvatarLog.Printf("Failed to get in cluser config, err: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.IluvatarLog.Printf("Failed to create clientset, err: %v", err)
	}
	return clientset
}

func (kc *kubeCollector) collect(ctx *ixContext) {
	kc.once.Do(func() {
		var err error

		ret := utils.ValidatePath(socket)
		if !ret {
			logger.IluvatarLog.Errorf("Failed to find '%s'\n", socket)
			return
		}
		kc.clientset = initClientSet()

		kc.conn, err = kc.connectToKubelet(socket)
		if err != nil {
			logger.IluvatarLog.Errorln(err)
			return
		}
	})

	for {
		select {
		case <-ctx.done():
			// Close the gRPC connection.
			logger.IluvatarLog.Infoln("Disconnect to kubelet")
			kc.conn.Close()
			return
		case <-ctx.signal():
			logger.IluvatarLog.Infoln("Start to collect kubernetes metrics")
			kc.collectMetrics(ctx)
		}
	}
}

func (kc *kubeCollector) collectMetrics(ctx *ixContext) {
	labels := make(map[string]labelType)

	pods, err := kc.listPods()
	if err != nil {
		logger.IluvatarLog.Errorln(err)
	} else {
		gpuPods := kc.filterGpuPods(pods, kc.gpus.gpus)
		for uuid, pod := range gpuPods {
			podInfo, err := kc.clientset.CoreV1().Pods(pod.namespace).Get(context.TODO(), pod.name, v1.GetOptions{})
			if err != nil {
				logger.IluvatarLog.Errorf("Failed to get pod %v", err)
			}

			nodeName := podInfo.Spec.NodeName
			logger.IluvatarLog.Infof("Pod %s in namespace %s is running on node: %s\n", pod.name, pod.namespace, nodeName)

			labels[uuid] = labelType{
				"container": pod.container,
				"pod":       pod.name,
				"namespace": pod.namespace,
				"node_name": nodeName,
			}
		}
	}

	ctx.updateMetrics(labels)
}

func (kc *kubeCollector) specificSplitBoard() error {
	reader, err := os.Open(ConfigFile)
	if err != nil {
		return err
	}

	defer reader.Close()

	clusterConfig, err := config.ParseConfigFrom(reader)
	if err != nil {
		logger.IluvatarLog.Errorf("error parsing config file: %v", err)
		return err
	}

	kc.SplitBoard = clusterConfig.Flags.SplitBoard
	return nil
}

func (kc *kubeCollector) filterGpuPods(pods *podresourcesapi.ListPodResourcesResponse, gpus map[string]gpuInfo) map[string]gpuPod {
	gpuPods := make(map[string]gpuPod)

	if err := kc.specificSplitBoard(); err != nil {
		logger.IluvatarLog.Errorf("Failed to get split board %v", err)
	}

	logger.IluvatarLog.Infoln("get split board", kc.SplitBoard)

	for _, pod := range pods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {
				resourceName := device.GetResourceName()
				if resourceName != iluvatarResourceName {
					continue
				}
				var gpusUuid []string

				for _, uuid := range device.GetDeviceIds() {
					uuidTmp := config.RemoveDeviceIduffix(uuid)
					if !kc.SplitBoard {
						if uuid_slary, ok := kc.gpus.pairChips[uuidTmp]; ok {
							if uuid_slary != uuidTmp {
								gpusUuid = append(gpusUuid, uuidTmp)
								gpusUuid = append(gpusUuid, uuid_slary)
							} else {
								gpusUuid = append(gpusUuid, uuidTmp)
							}
						}
					} else {
						gpusUuid = append(gpusUuid, uuidTmp)
					}
				}

				logger.IluvatarLog.Infoln("get gpusUuid", gpusUuid)

				for _, uuid := range gpusUuid {
					if _, ok := gpuPods[uuid]; !ok {
						gpuPods[uuid] = gpuPod{
							name:      pod.GetName(),
							namespace: pod.GetNamespace(),
							container: container.GetName(),
						}
					} else {
						infoTmp := gpuPods[uuid]
						infoTmp.name = infoTmp.name + ";" + pod.GetName()
						infoTmp.namespace = infoTmp.namespace + ";" + pod.GetNamespace()
						infoTmp.container = infoTmp.container + ";" + container.GetName()
						gpuPods[uuid] = infoTmp
					}
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
		logger.IluvatarLog.Errorf("Failed to connect to %s: %v", socket, err)
		return nil, err
	}

	return conn, nil
}

func (kc *kubeCollector) listPods() (*podresourcesapi.ListPodResourcesResponse, error) {
	client := podresourcesapi.NewPodResourcesListerClient(kc.conn)

	ctx, cancel := context.WithTimeout(context.Background(), kc.timeout)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		logger.IluvatarLog.Errorf("Failed to pod resources %v", err)
		return nil, err
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
