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

	testCases := map[string]struct {
		runtimeName string
		expectURL   string
		hcfsStatus  *v1alpha1.HCFSStatus
		expectErr   bool
	}{
		"test case 1": {
			runtimeName: runtimeName,
			hcfsStatus: &v1alpha1.HCFSStatus{
				Endpoint: "hcfsEndpoints",
			},
			expectURL: "hcfsEndpoints",
			expectErr: false,
		},
		"test case 2": {
			runtimeName: "fake",
			hcfsStatus: &v1alpha1.HCFSStatus{
				Endpoint: "hcfsEndpoints",
			},
			expectURL: "",
			expectErr: false,
		},
		"test case 3": {
			runtimeName: runtimeName,
			hcfsStatus: &v1alpha1.HCFSStatus{
				Endpoint: "",
			},
			expectURL: "",
			expectErr: true,
		},
		"test case 4": {
			runtimeName: "runtimeName3",
			hcfsStatus:  nil,
			expectURL:   "",
			expectErr:   false,
		},
		"test case 5": {
			runtimeName: "",
			hcfsStatus:  nil,
			expectURL:   "",
			expectErr:   true,
		},
	}

	for index, testCase := range testCases {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Annotations: map[string]string{common.DatasetUseAsHCFS: runtimeName},
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
		dataset := v1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testCase.runtimeName,
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
		if testCase.runtimeName == "" {
			_, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
			if err == nil {
				t.Errorf("expect error is not nil")
			}
			continue
		}
		runtimeInfo, err := base.BuildRuntimeInfo(testCase.runtimeName, namespace, "alluxio", v1alpha1.TieredStore{})
		if err != nil {
			t.Errorf("testcase %v fail, fail to create the runtimeInfo with error %v", index, err)
		}
		shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": runtimeInfo})
		if err != nil && !testCase.expectErr {
			t.Errorf("testcase %v fail, fail to mutate pod with error %v", index, err)
		}
		if err == nil && testCase.expectErr {
			t.Errorf("testcase %v fail, should return err", index)
		}
		if shouldStop {
			t.Errorf("testcase %v fail, expect shouldStop as false, but got %v", index, shouldStop)
		}
		containers := append(pod.Spec.InitContainers, pod.Spec.Containers...)
		for _, container := range containers {
			find := false
			for _, env := range container.Env {
				if env.Name == testCase.runtimeName+common.URLPostfix {
					if testCase.expectURL == "" {
						t.Errorf("testcase %v fail, expect not to inject env", index)
					} else {
						find = true
						if env.Value != testCase.expectURL {
							t.Errorf("testcase %v fail, expect to inject env %v, but get %v", index, testCase.expectURL, env.Value)
						}
					}
				}
			}
			if find != true && testCase.expectURL != "" {
				t.Errorf("testcase %v fail, fail to inject env %v", index, testCase.expectURL)
			}
		}
	}

}
