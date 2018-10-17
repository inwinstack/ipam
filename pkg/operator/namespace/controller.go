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

package namespace

import (
	"reflect"

	"github.com/golang/glog"
	opkit "github.com/inwinstack/operator-kit"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

var Resource = opkit.CustomResource{
	Name:    "namespace",
	Plural:  "namespaces",
	Version: "v1",
	Kind:    reflect.TypeOf(v1.Service{}).Name(),
}

type NamespaceController struct {
	ctx *opkit.Context
}

func NewController(ctx *opkit.Context) *NamespaceController {
	return &NamespaceController{ctx: ctx}
}

func (c *NamespaceController) StartWatch(namespace string, stopCh chan struct{}) error {
	resourceHandlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	}

	glog.Infof("start watching namespace resources")
	watcher := opkit.NewWatcher(Resource, namespace, resourceHandlerFuncs, c.ctx.Clientset.CoreV1().RESTClient())
	go watcher.Watch(&v1.Service{}, stopCh)
	return nil
}

func (c *NamespaceController) onAdd(obj interface{}) {
	glog.Infof("Namespace resource onAdd: %s", obj.(*v1.Namespace).Name)
}

func (c *NamespaceController) onUpdate(oldObj, newObj interface{}) {
	glog.Infof("Namespace resource onUpdate: %s", newObj.(*v1.Namespace).Name)
}

func (c *NamespaceController) onDelete(obj interface{}) {
	glog.Infof("Namespace resource onDelete.")
}
