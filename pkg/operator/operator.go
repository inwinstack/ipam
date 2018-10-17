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
	ipamclientset "github.com/inwinstack/ipam-operator/pkg/client/clientset/versioned/typed/ipam/v1alpha1"
	"github.com/inwinstack/ipam-operator/pkg/operator/namespace"
	"github.com/inwinstack/ipam-operator/pkg/operator/pool"
	"github.com/inwinstack/ipam-operator/pkg/util/k8sutil"
	opkit "github.com/inwinstack/operator-kit"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Flag struct {
	Kubeconfig string
}

type Operator struct {
	ctx       *opkit.Context
	namespace *namespace.NamespaceController
	pool      *pool.PoolController
	resources []opkit.CustomResource
	flag      *Flag
}

func NewMainOperator(flag *Flag) *Operator {
	return &Operator{
		resources: []opkit.CustomResource{pool.Resource},
		flag:      flag,
	}
}

func (o *Operator) Initialize() error {
	glog.V(2).Info("initialize the operator resources.")

	ctx, clientset, err := o.initContextAndClient()
	if err != nil {
		return err
	}
	o.namespace = namespace.NewController(ctx)
	o.pool = pool.NewController(ctx, clientset)
	o.ctx = ctx
	return nil
}

func (o *Operator) initContextAndClient() (*opkit.Context, ipamclientset.IpamV1alpha1Interface, error) {
	glog.V(2).Info("initialize the operator context and client.")

	config, err := k8sutil.GetRestConfig(o.flag.Kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Kubernetes config. %+v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Kubernetes client. %+v", err)
	}

	extensionsclient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes API extension clientset. %+v", err)
	}

	ipamclient, err := ipamclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pan-operator clientset. %+v", err)
	}

	ctx := &opkit.Context{
		Clientset:             client,
		APIExtensionClientset: extensionsclient,
		Interval:              Interval,
		Timeout:               Timeout,
	}
	return ctx, ipamclient, nil
}

func (o *Operator) initResources() error {
	glog.V(2).Info("initialize the CRD resources.")

	ctx := opkit.Context{
		Clientset:             o.ctx.Clientset,
		APIExtensionClientset: o.ctx.APIExtensionClientset,
		Interval:              Interval,
		Timeout:               Timeout,
	}
	err := opkit.CreateCustomResources(ctx, o.resources)
	if err != nil {
		return fmt.Errorf("failed to create custom resource. %+v", err)
	}
	return nil
}

func (o *Operator) Run() error {
	for {
		err := o.initResources()
		if err == nil {
			break
		}
		glog.Errorf("failed to init resources. %+v. retrying...", err)
		<-time.After(InitRetryDelay)
	}

	signalChan := make(chan os.Signal, 1)
	stopChan := make(chan struct{})
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// start watching the resources
	o.pool.StartWatch(v1.NamespaceAll, stopChan)
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
