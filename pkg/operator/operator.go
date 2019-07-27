/*
Copyright © 2018 inwinSTACK Inc

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
	"context"
	"fmt"
	"time"

	blended "github.com/inwinstack/blended/generated/clientset/versioned"
	blendedinformers "github.com/inwinstack/blended/generated/informers/externalversions"
	"github.com/inwinstack/ipam/pkg/config"
	"github.com/inwinstack/ipam/pkg/operator/ip"
	"github.com/inwinstack/ipam/pkg/operator/pool"
)

const defaultSyncTime = time.Second * 30

// Operator represents an operator context
type Operator struct {
	clientset blended.Interface
	informer  blendedinformers.SharedInformerFactory
	cfg       *config.Config
	pool      *pool.Controller
	ip        *ip.Controller
}

// New creates an instance of the operator
func New(cfg *config.Config, clientset blended.Interface) *Operator {
	t := defaultSyncTime
	if cfg.SyncSec > 30 {
		t = time.Second * time.Duration(cfg.SyncSec)
	}
	o := &Operator{cfg: cfg, clientset: clientset}
	o.informer = blendedinformers.NewSharedInformerFactory(clientset, t)
	o.pool = pool.NewController(clientset, o.informer.Inwinstack().V1().Pools())
	o.ip = ip.NewController(clientset, o.informer.Inwinstack().V1().IPs())
	return o
}

// Run serves an isntance of the operator
func (o *Operator) Run(ctx context.Context) error {
	go o.informer.Start(ctx.Done())
	if err := o.pool.Run(ctx, o.cfg.Threads); err != nil {
		return fmt.Errorf("failed to run the pool controller: %s", err.Error())
	}
	if err := o.ip.Run(ctx, o.cfg.Threads); err != nil {
		return fmt.Errorf("failed to run the ip controller: %s", err.Error())
	}
	return nil
}

// Stop stops the main controller
func (o *Operator) Stop() {
	o.pool.Stop()
	o.ip.Stop()
}
