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
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestUpdateCacheOfDataset: Tests the UpdateCacheOfDataset method of JuiceFSEngine
// This test creates mock datasets and runtime objects, invokes UpdateCacheOfDataset,
// and verifies the cache state updates in the dataset status.
//
// TestUpdateDatasetStatus: Tests the UpdateDatasetStatus method of JuiceFSEngine
// This test simulates different dataset phases (Bound/Failed/None) and verifies
// the status field updates in the dataset.
//
// TestBindToDataset: Tests the BindToDataset method of JuiceFSEngine
// This test initializes datasets and runtime objects, invokes BindToDataset,
// and verifies the dataset status transitions to Bound phase with correct cache states.

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

// TestUpdateDatasetStatus is a test function that verifies the behavior of the UpdateDatasetStatus method in the JuiceFSEngine.
// It tests the function with different dataset phases (Bound, Failed, None) and ensures that the dataset status is updated correctly.
// The function creates a fake client with test datasets and runtime objects, then calls UpdateDatasetStatus with each phase.
// It checks if the dataset status matches the expected result after each update.
// If any of the checks fail, the test will report an error.
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
