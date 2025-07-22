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

package informer

import (
	"context"
	"fmt"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"gitee.com/deep-spark/ixexporter/pkg/collector"
	"gitee.com/deep-spark/ixexporter/pkg/config"
	"gitee.com/deep-spark/ixexporter/pkg/logger"
)

const RetryUpdateCount = 3
const PatchWaitTime = 300 * time.Millisecond

type GpuResetInfo struct {
	NodeName string          `json:"nodename"             yaml:"nodename"`
	Reset    bool            `json:"reset"                yaml:"reset"`
	Occupy   map[string]bool `json:"occupy"               yaml:"occupy"`
}

type IxInformer interface {
	Start()

	UpdateCm(obj interface{}, sig chan os.Signal)
}

type ixInformer struct {
	clientset kubernetes.Interface
	informer  cache.SharedIndexInformer
	factory   informers.SharedInformerFactory
	opts      *collector.Options
	ns        string // Namespace where the ConfigMap is located
}

var _ IxInformer = &ixInformer{}

func StartIxInformer(opts *collector.Options, sig chan os.Signal) error {
	kConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		return err
	}

	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		return fmt.Errorf("failed to get POD_NAMESPACE environment variable")
	}
	fieldSelector := fields.SelectorFromSet(fields.Set{"metadata.name": config.IXConfigMap})
	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		0,
		informers.WithNamespace(ns),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fieldSelector.String()
		}),
	)

	cmInformer := factory.Core().V1().ConfigMaps().Informer()
	resetInformer := &ixInformer{
		clientset: clientset,
		informer:  cmInformer,
		factory:   factory,
		opts:      opts,
		ns:        ns,
	}

	if err := resetInformer.initSplitBoard(); err != nil {
		return fmt.Errorf("failed to initialize splitboard config: %v", err)
	}

	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logger.IXLog.Infof("ConfigMap Added: %s/%s", ns, config.IXConfigMap)
			resetInformer.UpdateCm(obj, sig)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			logger.IXLog.Infof("ConfigMap Updated: %s/%s", ns, config.IXConfigMap)
			resetInformer.UpdateCm(newObj, sig)
		},
		DeleteFunc: func(obj interface{}) {
			logger.IXLog.Infof("ConfigMap Deleted: %s/%s", ns, config.IXConfigMap)
		},
	})

	go resetInformer.Start()

	return nil
}

func (f *ixInformer) Start() {
	// Wait for the metrics server is running
	time.Sleep(5 * time.Second)

	stopCh := make(chan struct{})
	defer close(stopCh)

	logger.IXLog.Info("Starting informer for ConfigMap")
	f.factory.Start(stopCh)

	// Wait for the informer to sync before processing events
	if !cache.WaitForCacheSync(stopCh, f.informer.HasSynced) {
		logger.IXLog.Fatal("Failed to sync informer cache.")
	}

	<-stopCh
}

func (f *ixInformer) initSplitBoard() error {
	logger.IXLog.Infoln("Start to init splitboard config")
	cm, err := f.clientset.CoreV1().ConfigMaps(f.ns).Get(context.TODO(), config.IXConfigMap, metav1.GetOptions{})
	if err != nil {
		logger.IXLog.Warningf("Failed to get %s configmap from %s namespace, err: %v", config.IXConfigMap, f.ns, err)
		config.SplitBoard = false // Default to false if configmap is not found
		logger.IXLog.Infof("SplitBoard config initialized to: %v", config.SplitBoard)
		return nil
	}

	ixConfig, ok := cm.Data[config.IXConfigDataKey]
	if !ok {
		return fmt.Errorf("key %q not found in ConfigMap %q", config.IXConfigDataKey, config.IXConfigMap)
	}

	clusterConfig, err := config.ParseClusterConfig(ixConfig)
	if err != nil {
		return err
	}

	config.SplitBoard = clusterConfig.Flags.SplitBoard
	logger.IXLog.Infof("SplitBoard config initialized to: %v", config.SplitBoard)
	return nil
}

func (f *ixInformer) UpdateCm(obj interface{}, sig chan os.Signal) {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		logger.IXLog.Errorf("the object is not a configmap")
		return
	}

	ixConfig, ok := cm.Data[config.IXConfigDataKey]
	if !ok {
		logger.IXLog.Errorf("can't find %s data in %s configmap", config.IXConfigDataKey, config.IXConfigMap)
		return
	}

	clusterConfig, err := config.ParseClusterConfig(ixConfig)
	if err != nil {
		logger.IXLog.Errorf("failed to parse cluster config: %v", err)
		return
	}

	if config.SplitBoard != clusterConfig.Flags.SplitBoard {
		config.SplitBoard = clusterConfig.Flags.SplitBoard
		logger.IXLog.Infof("SplitBoard config changed to: %v", config.SplitBoard)
	}
}
