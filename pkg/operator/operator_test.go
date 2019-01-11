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
	"testing"
	"time"

	opkit "github.com/inwinstack/operator-kit"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	corefake "k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createCRD(context *opkit.Context, resource opkit.CustomResource) error {
	crdName := fmt.Sprintf("%s.%s", resource.Plural, resource.Group)
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   resource.Group,
			Version: resource.Version,
			Scope:   resource.Scope,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Singular:   resource.Name,
				Plural:     resource.Plural,
				Kind:       resource.Kind,
				ShortNames: resource.ShortNames,
			},
		},
	}

	_, err := context.APIExtensionClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create %s CRD. %+v", resource.Name, err)
		}
	}
	return nil
}

func TestOperator(t *testing.T) {
	coreClient := corefake.NewSimpleClientset()
	extensionsClient := extensionsfake.NewSimpleClientset()

	operator := NewMainOperator("")
	operator.ctx = &opkit.Context{
		Clientset:             coreClient,
		APIExtensionClientset: extensionsClient,
		Interval:              500 * time.Millisecond,
		Timeout:               60 * time.Second,
	}

	assert.NotNil(t, operator)
	assert.Equal(t, coreClient, operator.ctx.Clientset)
	assert.Equal(t, extensionsClient, operator.ctx.APIExtensionClientset)

	for _, res := range operator.resources {
		assert.Nil(t, createCRD(operator.ctx, res))
	}

	crds, err := extensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	assert.Nil(t, err)
	assert.Equal(t, len(operator.resources), len(crds.Items))

	for index, crd := range crds.Items {
		assert.Equal(t, operator.resources[index].Group, crd.Spec.Group)
		assert.Equal(t, operator.resources[index].Scope, crd.Spec.Scope)
		assert.Equal(t, operator.resources[index].Name, crd.Spec.Names.Singular)
		assert.Equal(t, operator.resources[index].Kind, crd.Spec.Names.Kind)
		assert.Equal(t, operator.resources[index].Plural, crd.Spec.Names.Plural)
		assert.Equal(t, operator.resources[index].ShortNames, crd.Spec.Names.ShortNames)
	}
}
