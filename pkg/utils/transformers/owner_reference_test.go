/*
Copyright 2021 The Fluid Authors.

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

package transfromer

import (
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestGenerateOwnerReferenceFromCRD(t *testing.T) {
	var (
		name      string                = "test-dataset"
		namespace string                = "fluid"
		dataset   *datav1alpha1.Dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				UID:       "12345",
			},
		}
		expect *common.OwnerReference = &common.OwnerReference{
			Enabled:            true,
			Controller:         true,
			BlockOwnerDeletion: false,
			UID:                "12345",
			Kind:               "Dataset",
			APIVersion:         "data.fluid.io/v1alpha1",
			Name:               name,
		}
	)

	var testScheme *runtime.Scheme = runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(testScheme)
	testScheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
	testObjs := []runtime.Object{}

	testObjs = append(testObjs, dataset.DeepCopy())

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	obj := &datav1alpha1.Dataset{}

	err := fakeClient.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, obj)

	if err != nil {
		t.Errorf("Failed due to %v", err)
	}

	result := GenerateOwnerReferenceFromObject(obj)
	if !reflect.DeepEqual(result, expect) {
		t.Errorf("The expect %v, the result: %v, they are not equal", expect, result)
	}
}
