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

package juicefs

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func getTestJuiceFSEngine(client client.Client, name string, namespace string) *JuiceFSEngine {
	runTime := &datav1alpha1.JuiceFSRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.JuiceFSRuntime)
	engine := &JuiceFSEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	return engine
}

func TestJuiceFSEngine_HasDeprecatedCommonLabelName(t *testing.T) {
	daemonSetWithSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fuse1-fuse",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-fuse1": "selector"}},
			},
		},
	}
	daemonSetWithoutSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fuse2-fuse",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-fuse1": "selector"}},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, daemonSetWithSelector)
	runtimeObjs = append(runtimeObjs, daemonSetWithoutSelector)
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)

	testCases := []struct {
		name      string
		namespace string
		out       bool
		isErr     bool
	}{
		{
			name:      "fuse1",
			namespace: "fluid",
			out:       true,
			isErr:     false,
		},
		{
			name:      "none",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
		{
			name:      "fuse2",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
	}

	for _, test := range testCases {
		engine := getTestJuiceFSEngine(fakeClient, test.name, test.namespace)
		out, err := engine.HasDeprecatedCommonLabelName()
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.out, out)
		}
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.isErr, isErr)
		}
	}
}

func TestJuiceFSEngine_getDeprecatedCommonLabelName(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "fuse1",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-fuse1",
		},
		{
			name:      "fuse2",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-fuse2",
		},
		{
			name:      "fluid",
			namespace: "test",
			out:       "data.fluid.io/storage-test-fluid",
		},
	}
	for _, test := range testCases {
		out := utils.GetCommonLabelName(true, test.namespace, test.name, "")
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %s, got %s", test.namespace, test.name, test.out, out)
		}
	}
}
