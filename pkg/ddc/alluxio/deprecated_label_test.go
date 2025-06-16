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

// TestAlluxioEngine_HasDeprecatedCommonLabelname tests the detection of deprecated labels in the Alluxio engine.
// This test verifies whether the HasDeprecatedCommonLabelname method can correctly identify if a DaemonSet
// contains deprecated label formats.
func TestAlluxioEngine_HasDeprecatedCommonLabelname(t *testing.T) {
	// Create a DaemonSet with a specific node selector.
	// This DaemonSet uses a deprecated label format: "data.fluid.io/storage-<runtime>-<dataset>"
	daemonSetWithSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",  // workload name
			Namespace: "fluid",         // namespace
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					// Node selector uses the deprecated label format
					NodeSelector: map[string]string{
						"data.fluid.io/storage-fluid-hbase": "selector",
					},
				},
			},
		},
	}

	// Create another DaemonSet (different name but same deprecated label format)
	daemonSetWithoutSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hadoop-worker",  // different workload
			Namespace: "fluid",          // same namespace
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					// Also uses the deprecated label format
					NodeSelector: map[string]string{
						"data.fluid.io/storage-fluid-hbase": "selector",
					},
				},
			},
		},
	}

	// Prepare Kubernetes API objects for testing
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, daemonSetWithSelector)
	runtimeObjs = append(runtimeObjs, daemonSetWithoutSelector)
	// Create Scheme and register API types
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	// Use a fake client to simulate Kubernetes API
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)
	// Define test cases
	testCases := []struct {
		name      string // dataset name
		namespace string // namespace
		out       bool   // expected result
		isErr     bool   // whether an error is expected
	}{
		{
			name:      "hbase",  // matches an existing DaemonSet
			namespace: "fluid",
			out:       true,     // should detect deprecated label
			isErr:     false,
		},
		{
			name:      "none",   // dataset does not exist
			namespace: "fluid",
			out:       false,    // should not detect deprecated label
			isErr:     false,
		},
		{
			name:      "hadoop", // DaemonSet exists but name does not match
			namespace: "fluid",
			out:       false,    // should not detect deprecated label
			isErr:     false,
		},
	}

	// Execute all test cases
	for _, test := range testCases {
		// Create an Alluxio engine instance for the current test case
		engine := getTestAlluxioEngine(fakeClient, test.name, test.namespace)
		// Call the method under test
		out, err := engine.HasDeprecatedCommonLabelname()
		// Validate the result
		if out != test.out {
			t.Errorf(
				"Dataset %s/%s test failed: expected %t, got %t",
				test.namespace, test.name, test.out, out,
			)
		}
		// Validate error expectation
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf(
				"Dataset %s/%s error check failed: expected error %t, got %t",
				test.namespace, test.name, test.isErr, isErr,
			)
		}
	}
}
