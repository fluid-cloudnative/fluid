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

package alluxio

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTestAlluxioEngine(client client.Client, name string, namespace string) *AlluxioEngine {
	runTime := &datav1alpha1.AlluxioRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "alluxio")
	engine := &AlluxioEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	return engine
}

// TestAlluxioEngine_GetDeprecatedCommonLabelname tests the GetCommonLabelName function,
// to ensure that it generates the correct common label name under different test cases.
//
// Parameters for the GetCommonLabelName function:
//   - useDeprecated (bool): Whether to use the deprecated label format.
//   - namespace (string): The namespace to which the resource belongs.
//   - name (string): The name of the resource.
//   - suffix (string): The suffix of the label, which is not used in this test case.
//
// This test checks various combinations of `namespace` and `name`,
// and validates whether the generated label name matches the expected output, ensuring the function logic is correct.
func TestAlluxioEngine_GetDeprecatedCommonLabelname(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-hbase",
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-hadoop",
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

func TestAlluxioEngine_HasDeprecatedCommonLabelname(t *testing.T) {

	// worker-name = e.name+"-worker"
	daemonSetWithSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase": "selector"}},
			},
		},
	}
	daemonSetWithoutSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hadoop-worker",
			Namespace: "fluid",
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: map[string]string{"data.fluid.io/storage-fluid-hbase": "selector"}},
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
			name:      "hbase",
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
			name:      "hadoop",
			namespace: "fluid",
			out:       false,
			isErr:     false,
		},
	}

	for _, test := range testCases {
		engine := getTestAlluxioEngine(fakeClient, test.name, test.namespace)
		out, err := engine.HasDeprecatedCommonLabelname()
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.out, out)
		}
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf("input parameter is %s-%s,expected %t, got %t", test.namespace, test.name, test.isErr, isErr)
		}
	}
}
