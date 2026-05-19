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

package reset

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/yaml"

	"gitee.com/deep-spark/ixexporter/pkg/collector"
	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
)

type gpuResetInformer struct {
	clientset kubernetes.Interface
	informer  cache.SharedIndexInformer
	factory   informers.SharedInformerFactory
	opts      *collector.Options

	gpuResetCM string
	resetState bool
}

var _ GpuResetInformer = &gpuResetInformer{}

func StartResetInformer(opts *collector.Options, sig chan os.Signal, ch chan ResetEvent) GpuResetInformer {
	cm := GpuResetCMPrefix + os.Getenv("NODE_NAME")
	fieldSelector := fields.SelectorFromSet(fields.Set{"metadata.name": cm})
	factory := informers.NewSharedInformerFactoryWithOptions(
		opts.ClientSet,
		0,
		informers.WithNamespace(opts.Namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fieldSelector.String()
		}),
	)

	cmInformer := factory.Core().V1().ConfigMaps().Informer()

	resetInformer := &gpuResetInformer{
		clientset: opts.ClientSet,
		informer:  cmInformer,
		factory:   factory,
		opts:      opts,

		gpuResetCM: cm,
		resetState: false, // default is false
	}

	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if config.ResetGpu {
				logger.IXLog.Infof("ConfigMap Added: %s/%s", opts.Namespace, cm)
				resetInformer.AddCm(obj, sig, ch)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if config.ResetGpu {
				logger.IXLog.Infof("ConfigMap Updated: %s/%s", opts.Namespace, cm)
				resetInformer.UpdateCm(newObj, sig, ch)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if config.ResetGpu {
				logger.IXLog.Infof("ConfigMap Deleted: %s/%s", opts.Namespace, cm)
				resetInformer.DeleteCm(sig, ch)
			}
		},
	})

	go resetInformer.Start()
	return resetInformer
}

func (g *gpuResetInformer) Start() {
	logger.IXLog.Debug("Starting gpuResetInformer...")

	// Wait for the metrics server is running
	time.Sleep(5 * time.Second)

	stopCh := make(chan struct{})
	defer close(stopCh)

	logger.IXLog.Debugf("Starting informer for the configmap %s/%s", g.opts.Namespace, g.gpuResetCM)
	g.factory.Start(stopCh)

	// Wait for the informer to sync before processing events
	if !cache.WaitForCacheSync(stopCh, g.informer.HasSynced) {
		logger.IXLog.Fatal("Failed to wait for the gpu-reset informer to sync.")
	}

	<-stopCh
}

func (g *gpuResetInformer) Stop() {
	logger.IXLog.Debug("Stopping gpuResetInformer...")
	if config.ResetGpu {
		g.PatchCm(PatchOpDeleteKey)
	}
}

// For cases: reset configmap is created or exporter is restarted.
func (g *gpuResetInformer) AddCm(obj interface{}, sig chan os.Signal, ch chan ResetEvent) {
	info, err := g.parseResetInfo(obj)
	if err != nil {
		logger.IXLog.Error(err)
		return
	}
	if info.Reset {
		logger.IXLog.Infof("The found reset state is true when exporter is restarted.")
		sig <- syscall.SIGHUP
		ch <- ResetToTrue
		logger.IXLog.Infof("Waiting for metrics server to stop...")
		<-ShutdownFlag
		logger.IXLog.Infof("Metrics server is stopped. Set occupy flag to false in ConfigMap.")
		g.PatchCm(PatchOpSetFalse)
	} else {
		g.PatchCm(PatchOpSetTrue)
	}
	g.resetState = info.Reset

}

func (g *gpuResetInformer) UpdateCm(obj interface{}, sig chan os.Signal, ch chan ResetEvent) {
	info, err := g.parseResetInfo(obj)
	if err != nil {
		logger.IXLog.Error(err)
		return
	}

	newResetVal := info.Reset
	if g.resetState != newResetVal && newResetVal {
		logger.IXLog.Infof("Reset state changed to true, pause metrics server and release GPUs")
		sig <- syscall.SIGHUP
		ch <- ResetToTrue

		logger.IXLog.Infof("Waiting for metrics server to stop...")
		<-ShutdownFlag
		logger.IXLog.Infof("Metrics server is stopped. Setting occupy flag to false in ConfigMap.")
		if err := g.PatchCm(PatchOpSetFalse); err != nil {
			return
		}

	} else if g.resetState != newResetVal && !newResetVal {
		logger.IXLog.Infof("Reset state changed to false, restart the metrics server and occupy GPUs")
		ch <- ResetToFalse
		logger.IXLog.Infof("Restart metrics server, and set occupy flag to true in ConfigMap.")
		if err := g.PatchCm(PatchOpSetTrue); err != nil {
			return
		}
	} else {
		occupyExist := false
		for k := range info.Occupy {
			if k == ExporterOccupyKey {
				occupyExist = true
				break
			}
		}
		// When the gpu-reset configmap is created, the occupy key may not exist.
		if !occupyExist {
			logger.IXLog.Infof("The occupy key %s does not exist, set it to true.", ExporterOccupyKey)
			if err := g.PatchCm(PatchOpSetTrue); err != nil {
				return
			}
		}
	}

	g.resetState = newResetVal
}

func (g *gpuResetInformer) PatchCm(op PatchOperation) error {
	ctx := context.Background()

	for cnt := 0; cnt < RetryCount; cnt++ {
		cm, err := g.clientset.CoreV1().ConfigMaps(g.opts.Namespace).Get(ctx, g.gpuResetCM, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return fmt.Errorf("configmap %s/%s not found", g.opts.Namespace, g.gpuResetCM)
			}
			logger.IXLog.Warningf("Failed to get configmap %s/%s, err: %v", g.opts.Namespace, g.gpuResetCM, err)
			time.Sleep(RetryWaitTime)
			continue
		}

		gpuResetStr, exists := cm.Data[GpuResetKey]
		if !exists {
			return fmt.Errorf("configmap %s does not have key %s", g.gpuResetCM, GpuResetKey)
		}

		info := GpuResetInfo{}
		if err := yaml.Unmarshal([]byte(gpuResetStr), &info); err != nil {
			return fmt.Errorf("failed to unmarshal gpuRest: %v", err)
		}
		switch op {
		case PatchOpSetTrue:
			info.Occupy[ExporterOccupyKey] = true
		case PatchOpSetFalse:
			info.Occupy[ExporterOccupyKey] = false
		case PatchOpDeleteKey:
			delete(info.Occupy, ExporterOccupyKey)
		default:
			return fmt.Errorf("unknown patch operation: %v", op)
		}

		updatedGpuRest, err := yaml.Marshal(info)
		if err != nil {
			return fmt.Errorf("failed to marshal updated gpuRest: %v", err)
		}
		cm.Data[GpuResetKey] = string(updatedGpuRest)
		_, err = g.clientset.CoreV1().ConfigMaps(g.opts.Namespace).Update(ctx, cm, metav1.UpdateOptions{})
		if err == nil {
			logger.IXLog.Infof("Successfully updated configmap %s/%s with data: \n%v", g.opts.Namespace, g.gpuResetCM, cm.Data)
			return nil
		}
		logger.IXLog.Warningf("Failed to update configmap %s, err: %v, retrying...", g.gpuResetCM, err)
		time.Sleep(RetryWaitTime)
	}

	logger.IXLog.Errorf("Failed to update configmap %s/%s after %d retries", g.opts.Namespace, g.gpuResetCM, RetryCount)
	return fmt.Errorf("failed to update configmap %s/%s", g.opts.Namespace, g.gpuResetCM)
}

func (g *gpuResetInformer) DeleteCm(sig chan os.Signal, ch chan ResetEvent) {
	if g.resetState {
		logger.IXLog.Infof("Reserverd reset value is true, change it to false and send ResetToFalse to channel")
		ch <- ResetToFalse
		g.resetState = false
	}
}

func (g *gpuResetInformer) parseResetInfo(obj interface{}) (*GpuResetInfo, error) {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("the object is not a configmap")
	}
	gpuResetStr, exists := cm.Data[GpuResetKey]
	if !exists {
		return nil, fmt.Errorf("configmap %s does not have key %s", g.gpuResetCM, GpuResetKey)
	}

	info := &GpuResetInfo{}
	if err := yaml.Unmarshal([]byte(gpuResetStr), info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal gpu reset info: %v", err)
	}
	return info, nil
}
