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
	"fmt"
	"net"
	"os"
	"time"

	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
	"gitee.com/deep-spark/ixexporter/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

const (
	socketFile           = "/var/lib/kubelet/pod-resources/kubelet.sock"
	IXResourceName       = "iluvatar.com/gpu"
	DefaultPodResMaxSize = 1024 * 1024 * 16 // 16 Mb
	IXConfigMap          = "ix-config"
	IXConfigDataKey      = "ix-config"
)

type gpuPod struct {
	name      string
	container string
	namespace string
}

type kubeCollector struct {
	clientset  kubernetes.Interface
	sysinfo    SysInfo
	timeout    time.Duration
	SplitBoard bool
}

func initClientSet() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.IXLog.Printf("Failed to get in cluster config, err: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.IXLog.Printf("Failed to create clientset, err: %v", err)
	}
	return clientset
}

func InitKubeCollector(ic *IXCollector) {
	if !ic.opts.EnableKube {
		return
	}
	logger.IXLog.Infof("Init kube collector.")

	kc := &kubeCollector{
		sysinfo: ic.sysinfo,
		timeout: KubeletConnTimeout,
	}

	ret := utils.ValidatePath(socketFile)
	if !ret {
		logger.IXLog.Errorf("Failed to find socket file: '%s'\n", socketFile)
		return
	}
	kc.clientset = initClientSet()
	if err := kc.confirmSplitBoard(); err != nil {
		logger.IXLog.Errorf("Failed to confirm splitboard config: %v", err)
		return
	}
	logger.IXLog.Infof("Splitboard config: %v", kc.SplitBoard)

	ic.ctx.registerCollector(kc)
	go kc.collect(ic.ctx)
}

func (kc *kubeCollector) collect(ctx *ixContext) {
	for {
		select {
		case <-ctx.done(): // Triggered by prometheus.Unregister()
			logger.IXLog.Infoln("Disconnect to kubelet")
			return
		case <-ctx.signalCh:
			logger.IXLog.Infoln("Start to collect kubernetes metrics")
			kc.collectMetrics(ctx)
			ctx.updateStat()
			logger.IXLog.Infoln("Kube collector has updated all metrics.")
		}
	}
}

func (kc *kubeCollector) collectMetrics(ctx *ixContext) {
	labels := make(map[string]LabelsMap)

	pods, err := kc.listPods()
	if err != nil {
		logger.IXLog.Errorln(err)
	} else {
		gpuPods := kc.filterGpuPods(pods)
		for uuid, pod := range gpuPods {
			logger.IXLog.Infof("Pod %s in namespace %s is running on node: %s", pod.name, pod.namespace, ctx.nodeName)

			labels[uuid] = LabelsMap{
				"container": pod.container,
				"pod":       pod.name,
				"namespace": pod.namespace,
			}
		}
	}

	logger.IXLog.Infof("Store kubernetes label values.")
	ctx.labelValues = labels
}

func (kc *kubeCollector) confirmSplitBoard() error {
	logger.IXLog.Infoln("Start to confirm splitboard config")
	kc.SplitBoard = false
	if os.Getenv("POD_NAMESPACE") == "" {
		logger.IXLog.Warningf("Haven't set POD_NAMESPACE environment variable")
		return nil
	}

	ns := os.Getenv("POD_NAMESPACE")
	cm, err := kc.clientset.CoreV1().ConfigMaps(ns).Get(context.TODO(), IXConfigMap, metav1.GetOptions{})
	if err != nil {
		logger.IXLog.Warningf("Can't get %s configmap from %s namespace: %v", IXConfigMap, ns, err)
		return nil
	}

	ixConfig, ok := cm.Data[IXConfigDataKey]
	if !ok {
		return fmt.Errorf("can't find %s data in %s configmap", IXConfigDataKey, IXConfigMap)
	}

	clusterConfig, err := config.ParseClusterConfig(ixConfig)
	if err != nil {
		return fmt.Errorf("error to parse cluster config: %v", err)
	}

	kc.SplitBoard = clusterConfig.Flags.SplitBoard
	return nil
}

func (kc *kubeCollector) filterGpuPods(pods *podresourcesapi.ListPodResourcesResponse) map[string]gpuPod {
	gpuPods := make(map[string]gpuPod)

	for _, pod := range pods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, containerDevices := range container.GetDevices() {
				resourceName := containerDevices.GetResourceName()
				if resourceName != IXResourceName {
					continue
				}
				var gpuUuids []string

				for _, uuid := range containerDevices.GetDeviceIds() {
					// gpuUuids = append(gpuUuids, utils.RemoveDeviceIduffix(uuid))
					uuidTmp := utils.RemoveDeviceIduffix(uuid)
					if !kc.SplitBoard {
						if uuid_slary, ok := kc.sysinfo.pairChips[uuidTmp]; ok {
							if uuid_slary != uuidTmp {
								gpuUuids = append(gpuUuids, uuidTmp)
								gpuUuids = append(gpuUuids, uuid_slary)
							} else {
								gpuUuids = append(gpuUuids, uuidTmp)
							}
						}
					} else {
						gpuUuids = append(gpuUuids, uuidTmp)
					}
				}

				logger.IXLog.Infoln("Get gpuUuids", gpuUuids)

				for _, uuid := range gpuUuids {
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

	dial := func(ctx context.Context, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", addr)
	}

	conn, err := grpc.DialContext(ctx, socket, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dial), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(DefaultPodResMaxSize)))

	if err != nil {
		logger.IXLog.Errorf("Failed to connect to %s: %v", socket, err)
		return nil, err
	}

	return conn, nil
}

// listPods list pods of local node.
func (kc *kubeCollector) listPods() (*podresourcesapi.ListPodResourcesResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), kc.timeout)
	defer cancel()

	logger.IXLog.Infof("Connecting to kubelet at %s", socketFile)
	conn, err := kc.connectToKubelet(socketFile)
	if err != nil {
		logger.IXLog.Errorf("Failed to connect to kubelet: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := podresourcesapi.NewPodResourcesListerClient(conn)
	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		logger.IXLog.Errorf("Failed to list pod resources from kubelet: %v", err)
		return nil, err
	}

	return resp, nil
}
