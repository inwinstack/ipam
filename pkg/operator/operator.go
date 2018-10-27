/*
Copyright © 2018 inwinSTACK.inc

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

package operator

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	inwinclientset "github.com/inwinstack/blended/client/clientset/versioned/typed/inwinstack/v1"
	"github.com/inwinstack/ipam-operator/pkg/k8sutil"
	"github.com/inwinstack/ipam-operator/pkg/operator/ip"
	"github.com/inwinstack/ipam-operator/pkg/operator/namespace"
	"github.com/inwinstack/ipam-operator/pkg/operator/pool"
	opkit "github.com/inwinstack/operator-kit"
	apiextensionsclients "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Flag struct {
	Kubeconfig                string
	Address                   string
	IgnoreNamespaces          []string
	IgnoreNamespaceAnnotation bool
	AutoAssignToNamespace     bool
}

type Operator struct {
	ctx       *opkit.Context
	namespace *namespace.NamespaceController
	pool      *pool.PoolController
	ip        *ip.IPController
	resources []opkit.CustomResource
	flag      *Flag
}

const (
	initRetryDelay = 10 * time.Second
	interval       = 500 * time.Millisecond
	timeout        = 60 * time.Second
)

func NewMainOperator(flag *Flag) *Operator {
	return &Operator{
		resources: []opkit.CustomResource{pool.Resource, ip.Resource},
		flag:      flag,
	}
}

func (o *Operator) Initialize() error {
	glog.V(2).Info("Initialize the operator resources.")

	ctx, clientset, err := o.initContextAndClient()
	if err != nil {
		return err
	}

	o.namespace = namespace.NewController(ctx, clientset)
	o.pool = pool.NewController(ctx, clientset)
	o.ip = ip.NewController(ctx, clientset)
	o.ctx = ctx
	return nil
}

func (o *Operator) initContextAndClient() (*opkit.Context, inwinclientset.InwinstackV1Interface, error) {
	glog.V(2).Info("Initialize the operator context and client.")

	config, err := k8sutil.GetRestConfig(o.flag.Kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get Kubernetes config. %+v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get Kubernetes client. %+v", err)
	}

	extensionsclient, err := apiextensionsclients.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create Kubernetes API extension clientset. %+v", err)
	}

	inwinclient, err := inwinclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create pan-operator clientset. %+v", err)
	}

	ctx := &opkit.Context{
		Clientset:             client,
		APIExtensionClientset: extensionsclient,
		Interval:              interval,
		Timeout:               timeout,
	}
	return ctx, inwinclient, nil
}

func (o *Operator) initResources() error {
	glog.V(2).Info("Initialize the CRD resources.")

	ctx := opkit.Context{
		Clientset:             o.ctx.Clientset,
		APIExtensionClientset: o.ctx.APIExtensionClientset,
		Interval:              interval,
		Timeout:               timeout,
	}

	if err := opkit.CreateCustomResources(ctx, o.resources); err != nil {
		return fmt.Errorf("Failed to create custom resource. %+v", err)
	}
	return nil
}

func (o *Operator) Run() error {
	for {
		err := o.initResources()
		if err == nil {
			break
		}
		glog.Errorf("Failed to init resources. %+v. retrying...", err)
		<-time.After(initRetryDelay)
	}

	signalChan := make(chan os.Signal, 1)
	stopChan := make(chan struct{})
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// start watching the custom resources
	o.ip.StartWatch(v1.NamespaceAll, stopChan)
	o.pool.StartWatch(v1.NamespaceAll, stopChan)

	// init the custom resources
	err := o.pool.CreateDefaultPool(
		o.flag.Address,
		o.flag.IgnoreNamespaces,
		o.flag.AutoAssignToNamespace,
		o.flag.IgnoreNamespaceAnnotation)
	if err != nil {
		return fmt.Errorf("Failed to create default IP pool. %+v", err)
	}

	// start watching the resources
	o.namespace.StartWatch(v1.NamespaceAll, stopChan)

	for {
		select {
		case <-signalChan:
			glog.Infof("Shutdown signal received, exiting...")
			close(stopChan)
			return nil
		}
	}
}
