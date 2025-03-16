This is an symbol text./*
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
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)


// TestUpdateCacheOfDataset tests the UpdateCacheOfDataset function of JuiceFSEngine.
// It verifies if the Dataset's cache states are correctly updated based on the associated JuiceFSRuntime's status.
// The test sets up a mock Dataset and JuiceFSRuntime, then triggers the cache update and checks if the Dataset's
// CacheStates field matches the expected status from the Runtime. It validates the synchronization between
// Runtime status and Dataset cache states in the Kubernetes cluster.

func TestUpdateCacheOfDataset(t *testing.T) {
// 1. Define mock Dataset objects for testing
	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

// 2. Convert Dataset objects to Kubernetes runtime objects and add to test environment
	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	
// 3. Define mock JuiceFSRuntime objects with cache status configuration
	testRuntimeInputs := []*datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Replicas: 1,
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
	}
	
// 4. Add Runtime objects to test environment
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	
// 5. Initialize fake Kubernetes client with test objects and scheme
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

// 6. Create JuiceFSEngine instance with mock dependencies
	engine := &JuiceFSEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		name:      "hbase",
		namespace: "fluid",
		runtime:   testRuntimeInputs[0],
	}
	
// 7. Execute the core function being tested: UpdateCacheOfDataset
	err := engine.UpdateCacheOfDataset()
	if err != nil {
		t.Errorf("fail to exec UpdateCacheOfDataset with error %v", err)
		return
	}

// 8. Define expected Dataset status for validation
	expectedDataset := datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Status: datav1alpha1.DatasetStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
	
// 9. Retrieve actual Dataset status from cluster
	var datasets datav1alpha1.DatasetList
	err = client.List(context.TODO(), &datasets)
	if err != nil {
		t.Errorf("fail to list the datasets with error %v", err)
		return
	}
	
// 10. Validate if actual status matches expected status
	if !reflect.DeepEqual(datasets.Items[0].Status, expectedDataset.Status) {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
}

func TestUpdateDatasetStatus(t *testing.T) {
	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
			Status: datav1alpha1.DatasetStatus{
				HCFSStatus: &datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	testRuntimeInputs := []*datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Replicas: 1,
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
	}
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		name:      "hbase",
		namespace: "fluid",
		runtime:   testRuntimeInputs[0],
	}

	var testCase = []struct {
		phase          datav1alpha1.DatasetPhase
		expectedResult datav1alpha1.Dataset
	}{
		{
			phase: datav1alpha1.BoundDatasetPhase,
			expectedResult: datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
					Runtimes: []datav1alpha1.Runtime{
						{
							Name:           "hbase",
							Namespace:      "fluid",
							Category:       common.AccelerateCategory,
							Type:           common.JuiceFSRuntime,
							MasterReplicas: 1,
						},
					},
				},
			},
		},
		{
			phase: datav1alpha1.FailedDatasetPhase,
			expectedResult: datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.FailedDatasetPhase,
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
				},
			},
		},
		{
			phase: datav1alpha1.NoneDatasetPhase,
			expectedResult: datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.NoneDatasetPhase,
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
				},
			},
		},
	}

	for _, test := range testCase {
		err := engine.UpdateDatasetStatus(test.phase)
		if err != nil {
			t.Errorf("fail to exec UpdateCacheOfDataset with error %v", err)
			return
		}

		var datasets datav1alpha1.DatasetList
		err = client.List(context.TODO(), &datasets)
		if err != nil {
			t.Errorf("fail to list the datasets with error %v", err)
			return
		}
		if !reflect.DeepEqual(datasets.Items[0].Status.Phase, test.expectedResult.Status.Phase) ||
			!reflect.DeepEqual(datasets.Items[0].Status.CacheStates, test.expectedResult.Status.CacheStates) ||
			!reflect.DeepEqual(datasets.Items[0].Status.HCFSStatus, test.expectedResult.Status.HCFSStatus) {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
	}
}

func TestBindToDataset(t *testing.T) {
	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
			Status: datav1alpha1.DatasetStatus{
				HCFSStatus: &datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	testRuntimeInputs := []*datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
	}
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		name:      "hbase",
		namespace: "fluid",
		runtime:   testRuntimeInputs[0],
	}

	var expectedResult = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Status: datav1alpha1.DatasetStatus{
			Phase: datav1alpha1.BoundDatasetPhase,
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
			HCFSStatus: &datav1alpha1.HCFSStatus{
				Endpoint:                    "test Endpoint",
				UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
			},
		},
	}
	err := engine.BindToDataset()
	if err != nil {
		t.Errorf("fail to exec UpdateCacheOfDataset with error %v", err)
		return
	}

	var datasets datav1alpha1.DatasetList
	err = client.List(context.TODO(), &datasets)
	if err != nil {
		t.Errorf("fail to list the datasets with error %v", err)
		return
	}
	if !reflect.DeepEqual(datasets.Items[0].Status.Phase, expectedResult.Status.Phase) ||
		!reflect.DeepEqual(datasets.Items[0].Status.CacheStates, expectedResult.Status.CacheStates) ||
		!reflect.DeepEqual(datasets.Items[0].Status.HCFSStatus, expectedResult.Status.HCFSStatus) {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
}
