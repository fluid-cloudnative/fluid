/*
 Copyright 2022 The Fluid Authors.

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

package hcfsaddressesinjector

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestMutate(t *testing.T) {
	var (
		runtimeName = "runtimeName"
		namespace   = "default"
	)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name:  "key",
							Value: "value",
						},
					},
				},
			},
			InitContainers: []corev1.Container{
				{
					Env: []corev1.EnvVar{
						{
							Name:  "key",
							Value: "value",
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		runtimeName string
		expectURL   string
		hcfsStatus  *v1alpha1.HCFSStatus
	}{
		"test case 1": {
			runtimeName: runtimeName,
			hcfsStatus: &v1alpha1.HCFSStatus{
				Endpoint: "hcfsEndpoints",
			},
			expectURL: "hcfsEndpoints",
		},
		"test case 2": {
			runtimeName: "fake",
			hcfsStatus: &v1alpha1.HCFSStatus{
				Endpoint: "hcfsEndpoints",
			},
			expectURL: common.UnknownURL,
		},
		"test case 3": {
			runtimeName: runtimeName,
			hcfsStatus: &v1alpha1.HCFSStatus{
				Endpoint: "",
			},
			expectURL: common.UnknownURL,
		},
		"test case 4": {
			runtimeName: "runtimeName3",
			hcfsStatus:  nil,
			expectURL:   common.UnknownURL,
		},
	}

	for index, testCase := range testCases {
		dataset := v1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: namespace,
			},
			Status: v1alpha1.DatasetStatus{
				HCFSStatus: testCase.hcfsStatus,
			},
		}
		s := runtime.NewScheme()
		_ = v1alpha1.AddToScheme(s)
		fakeClient := fake.NewFakeClientWithScheme(s, &dataset)
		plugin := NewPlugin(fakeClient)
		if plugin.GetName() != NAME {
			t.Errorf("testcase %v fail, GetName expect %v, got %v", index, NAME, plugin.GetName())
		}
		runtimeInfo, err := base.BuildRuntimeInfo(testCase.runtimeName, namespace, "alluxio", datav1alpha1.TieredStore{})
		if err != nil {
			t.Errorf("testcase %v fail, fail to create the runtimeInfo with error %v", index, err)
		}
		runtimeInfo.UseAsHcfs()
		shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": runtimeInfo})
		if err != nil {
			t.Errorf("testcase %v fail, fail to mutate pod with error %v", index, err)
		}
		if shouldStop {
			t.Errorf("testcase %v fail, expect shouldStop as false, but got %v", index, shouldStop)
		}
		containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
		for _, container := range containers {
			find := false
			for _, env := range container.Env {
				if env.Name == testCase.runtimeName+common.URLPostfix {
					find = true
					if env.Value != testCase.expectURL {
						t.Errorf("testcase %v fail, expect to inject env %v, but get %v", index, testCase.expectURL, env.Value)
					}
				}
			}
			if find != true {
				t.Errorf("testcase %v fail, fail to inject env %v", index, testCase.expectURL)
			}
		}
	}

}
