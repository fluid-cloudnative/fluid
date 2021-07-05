package alluxio

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
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
		Log:       log.NullLogger{},
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
			Runtimes: []datav1alpha1.Runtime{
				{
					Name:           "hbase",
					Namespace:      "fluid",
					Category:       common.AccelerateCategory,
					Type:           common.ALLUXIO_RUNTIME,
					MasterReplicas: 1,
				},
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
		Log:       log.NullLogger{},
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
		Log:       log.NullLogger{},
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
