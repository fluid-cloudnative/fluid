package jindo

import (
	"context"
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSyncMetadata(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "2Gi",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
					Path:     "local:///host1/erf",
					NodeName: "test-node",
				},
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	runtime := &datav1alpha1.JindoRuntime{
		Spec: datav1alpha1.JindoRuntimeSpec{
			Secret: "1",
		},
	}

	engines := []JindoEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
			runtime:   runtime,
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
			runtime:   runtime,
		},
	}

	for _, engine := range engines {
		err := engine.SyncMetadata()
		if err != nil {
			t.Errorf("fail to exec the function")
		}
	}

	engine := JindoEngine{
		name:      "hadoop",
		namespace: "fluid",
		Client:    client,
		Log:       log.NullLogger{},
		runtime:   runtime,
	}

	err := engine.SyncMetadata()
	if err != nil {
		t.Errorf("fail to exec function RestoreMetadataInternal")
	}
}

func TestShouldSyncMetadata(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "2Gi",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Status: datav1alpha1.DatasetStatus{
				UfsTotal: "",
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JindoEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       log.NullLogger{},
		},
	}

	var testCases = []struct {
		engine         JindoEngine
		expectedShould bool
	}{
		{
			engine:         engines[0],
			expectedShould: false,
		},
		{
			engine:         engines[1],
			expectedShould: true,
		},
	}

	for _, test := range testCases {
		should, err := test.engine.shouldSyncMetadata()
		if err != nil || should != test.expectedShould {
			t.Errorf("fail to exec the function")
		}
	}
}

func TestSyncMetadataInternal(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	runtime := &datav1alpha1.JindoRuntime{
		Spec: datav1alpha1.JindoRuntimeSpec{
			Secret: "1",
		},
	}

	engines := []JindoEngine{
		{
			name:               "hbase",
			namespace:          "fluid",
			Client:             client,
			Log:                log.NullLogger{},
			MetadataSyncDoneCh: make(chan MetadataSyncResult),
			runtime:            runtime,
		},
		{
			name:               "spark",
			namespace:          "fluid",
			Client:             client,
			Log:                log.NullLogger{},
			MetadataSyncDoneCh: nil,
			runtime:            runtime,
		},
	}

	result := MetadataSyncResult{
		StartTime: time.Now(),
		UfsTotal:  "2GB",
		Done:      true,
	}

	var testCase = []struct {
		engine           JindoEngine
		expectedResult   bool
		expectedUfsTotal string
	}{
		{
			engine:           engines[0],
			expectedUfsTotal: "2GB",
		},
	}

	for index, test := range testCase {
		if index == 0 {
			go func() {
				test.engine.MetadataSyncDoneCh <- result
			}()
		}

		err := test.engine.syncMetadataInternal()
		//	fmt.Println(index)
		if err != nil {
			t.Errorf("fail to exec the function with error %v", err)
		}

		key := types.NamespacedName{
			Namespace: test.engine.namespace,
			Name:      test.engine.name,
		}

		dataset := &datav1alpha1.Dataset{}
		err = client.Get(context.TODO(), key, dataset)
		if err != nil {
			t.Errorf("failt to get the dataset with error %v", err)
		}

		if dataset.Status.UfsTotal != test.expectedUfsTotal {
			t.Errorf("expected UfsTotal %s, get UfsTotal %s", test.expectedUfsTotal, dataset.Status.UfsTotal)
		}
	}
}
