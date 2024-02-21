/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newVineyardEngineRT(client client.Client, name string, namespace string, withRuntimeInfo bool, unittest bool) *VineyardEngine {
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "vineyard", v1alpha1.TieredStore{})
	engine := &VineyardEngine{
		runtime:     &v1alpha1.VineyardRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: nil,
		Log:         fake.NullLogger(),
	}

	if withRuntimeInfo {
		engine.runtimeInfo = runTimeInfo
	}
	return engine
}

func TestGetRuntimeInfo(t *testing.T) {
	anno := map[string]string{
		"the-key-should-not-be-changed": "the-value-should-not-be-changed",
	}
	addedMeta := []v1alpha1.Metadata{
		{
			PodMetadata: v1alpha1.PodMetadata{
				Labels: map[string]string{"fluid.io/node-puhlish-method": "symlink"},
			},
			Selector: metav1.GroupKind{Kind: "PersistentVolume"},
		},
	}

	runtimeInputs := []*v1alpha1.VineyardRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "hbase",
				Namespace:   "fluid",
				Annotations: anno,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "hadoop",
				Namespace:   "fluid",
				Annotations: anno,
			},
		},
	}
	daemonSetInputs := []*v1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-worker",
				Namespace: "fluid",
			},
			Spec: v1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase": "selector"}},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-worker",
				Namespace: "fluid",
			},
			Spec: v1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hadoop": "selector"}},
				},
			},
		},
	}
	dataSetInputs := []*v1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
		},
	}
	objs := []runtime.Object{}
	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, daemonSetInput := range daemonSetInputs {
		objs = append(objs, daemonSetInput.DeepCopy())
	}
	for _, dataSetInput := range dataSetInputs {
		objs = append(objs, dataSetInput.DeepCopy())
	}
	//scheme := runtime.NewScheme()
	//scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	//scheme.AddKnownTypes(v1alpha1.GroupVersion,runtimeInput)
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	testCases := []struct {
		name            string
		namespace       string
		withRuntimeInfo bool
		unittest        bool
		isErr           bool
		isNil           bool
	}{
		{
			name:            "hbase",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        false,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "hbase",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        true,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "hbase",
			namespace:       "fluid",
			withRuntimeInfo: true,
			isErr:           false,
			isNil:           false,
		},
		{
			name:            "hadoop",
			namespace:       "fluid",
			withRuntimeInfo: false,
			unittest:        false,
			isErr:           false,
			isNil:           false,
		},
	}
	for _, testCase := range testCases {
		engine := newVineyardEngineRT(fakeClient, testCase.name, testCase.namespace, testCase.withRuntimeInfo, testCase.unittest)
		runtimeInfo, err := engine.getRuntimeInfo()
		isNil := runtimeInfo == nil
		isErr := err != nil
		runtime, _ := engine.getRuntime()
		if !reflect.DeepEqual(runtime.Annotations, anno) {
			t.Errorf(" want %v, got %v", anno, runtime.Annotations)
		}
		if !testCase.withRuntimeInfo && !reflect.DeepEqual(runtimeInfo.GetMetadataList(), addedMeta) {
			t.Errorf(" want %v, got %v", addedMeta, runtimeInfo.GetMetadataList())
		}
		if isNil != testCase.isNil {
			t.Errorf(" want %t, got %t", testCase.isNil, isNil)
		}
		if isErr != testCase.isErr {
			t.Errorf(" want %t, got %t", testCase.isErr, isErr)
		}
	}
}
