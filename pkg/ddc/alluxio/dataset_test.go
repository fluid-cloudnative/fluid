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
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestUpdateCacheOfDataset(t *testing.T) {
	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	testRuntimeInputs := []*datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Master: datav1alpha1.AlluxioCompTemplateSpec{
					Replicas: 1,
				},
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

	engine := &AlluxioEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		name:      "hbase",
		namespace: "fluid",
		runtime:   testRuntimeInputs[0],
	}

	err := engine.UpdateCacheOfDataset()
	if err != nil {
		t.Errorf("fail to exec UpdateCacheOfDataset with error %v", err)
		return
	}

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

	var datasets datav1alpha1.DatasetList
	err = client.List(context.TODO(), &datasets)
	if err != nil {
		t.Errorf("fail to list the datasets with error %v", err)
		return
	}
	if !reflect.DeepEqual(datasets.Items[0].Status, expectedDataset.Status) {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
}

// TestUpdateDatasetStatus is a unit test function to verify the correctness of the UpdateDatasetStatus method 
// in the AlluxioEngine struct. This method is responsible for updating the status of a Dataset.
//
// Test procedure:
// 1. Create test datasets and runtime objects to simulate Kubernetes resources.
// 2. Use a fake client to manage these objects.
// 3. Initialize an instance of AlluxioEngine with test data.
// 4. Define test cases with different DatasetPhase values and expected results.
// 5. Execute UpdateDatasetStatus for each test case.
// 6. Validate whether the updated Dataset status matches the expected results, including Phase, CacheStates, and HCFSStatus.
// 7. If any test fails, output an error message.
//
// This test ensures that UpdateDatasetStatus correctly updates the Dataset's status and handles different DatasetPhase transitions properly.
func TestUpdateDatasetStatus(t *testing.T) {
	// Create test dataset inputs
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

	// Convert dataset inputs into runtime objects for the fake client
	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	// Create test runtime inputs
	testRuntimeInputs := []*datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Master: datav1alpha1.AlluxioCompTemplateSpec{
					Replicas: 1,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
	}

	// Convert runtime inputs into runtime objects for the fake client
	for _, runtimeInput := range testRuntimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}

	// Create a fake client with test objects
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	// Initialize AlluxioEngine instance
	engine := &AlluxioEngine{
		Client:    client,
		Log:       fake.NullLogger(),
		name:      "hbase",
		namespace: "fluid",
		runtime:   testRuntimeInputs[0],
	}

	// Define test cases with different dataset phases and expected results
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
							Type:           common.AlluxioRuntime,
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

	// Execute test cases
	for _, test := range testCase {
		// Update dataset status using the engine method
		err := engine.UpdateDatasetStatus(test.phase)
		if err != nil {
			t.Errorf("failed to execute UpdateDatasetStatus with error: %v", err)
			return
		}

		// Retrieve the updated dataset list
		var datasets datav1alpha1.DatasetList
		err = client.List(context.TODO(), &datasets)
		if err != nil {
			t.Errorf("failed to list datasets with error: %v", err)
			return
		}

		// Validate the dataset status matches the expected result
		if !reflect.DeepEqual(datasets.Items[0].Status.Phase, test.expectedResult.Status.Phase) ||
			!reflect.DeepEqual(datasets.Items[0].Status.CacheStates, test.expectedResult.Status.CacheStates) ||
			!reflect.DeepEqual(datasets.Items[0].Status.HCFSStatus, test.expectedResult.Status.HCFSStatus) {
			t.Errorf("unexpected dataset status, expected: %+v, got: %+v", test.expectedResult.Status, datasets.Items[0].Status)
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

	testRuntimeInputs := []*datav1alpha1.AlluxioRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
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

	engine := &AlluxioEngine{
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
