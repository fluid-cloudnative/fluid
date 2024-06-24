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

package thin

import (
	"context"
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
)

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
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "noautosync",
				Namespace: "fluid",
			},
		},
	}
	runtimeInputs := []datav1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "noautosync",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.ThinRuntimeSpec{
				RuntimeManagement: datav1alpha1.RuntimeManagement{
					MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
						AutoSync: pointer.Bool(false),
					},
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, input := range datasetInputs {
		testObjs = append(testObjs, input.DeepCopy())
	}
	for _, input := range runtimeInputs {
		testObjs = append(testObjs, input.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []ThinEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "noautosync",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine         ThinEngine
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
		{
			engine:         engines[2],
			expectedShould: false,
		},
	}

	for _, test := range testCases {
		should, err := test.engine.shouldSyncMetadata()
		if err != nil || should != test.expectedShould {
			t.Errorf("fail to exec the function")
		}
	}
}

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

	runtimeInputs := []datav1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, runtimeInput := range runtimeInputs {
		testObjs = append(testObjs, runtimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []ThinEngine{
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	for _, engine := range engines {
		err := engine.SyncMetadata()
		if err != nil {
			t.Errorf("fail to exec the function")
		}
	}

	engine := ThinEngine{
		name:      "hadoop",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	err := engine.SyncMetadata()
	if err != nil {
		t.Errorf("fail to exec function RestoreMetadataInternal: %v", err)
	}
}

func TestThinEngine_syncMetadataInternal(t *testing.T) {
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
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

	engines := []ThinEngine{
		{
			name:               "test1",
			namespace:          "fluid",
			Client:             client,
			Log:                fake.NullLogger(),
			MetadataSyncDoneCh: make(chan base.MetadataSyncResult),
		},
		{
			name:               "test2",
			namespace:          "fluid",
			Client:             client,
			Log:                fake.NullLogger(),
			MetadataSyncDoneCh: nil,
		},
	}

	result := base.MetadataSyncResult{
		StartTime: time.Now(),
		UfsTotal:  "2GB",
		Done:      true,
		FileNum:   "5",
	}

	var testCase = []struct {
		engine           ThinEngine
		expectedResult   bool
		expectedUfsTotal string
		expectedFileNum  string
	}{
		{
			engine:           engines[0],
			expectedUfsTotal: "2GB",
			expectedFileNum:  "5",
		},
	}

	for index, test := range testCase {
		if index == 0 {
			go func() {
				test.engine.MetadataSyncDoneCh <- result
			}()
		}

		err := test.engine.syncMetadataInternal()
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

		if dataset.Status.UfsTotal != test.expectedUfsTotal || dataset.Status.FileNum != test.expectedFileNum {
			t.Errorf("expected UfsTotal %s, get UfsTotal %s, expected FileNum %s, get FileNum %s", test.expectedUfsTotal, dataset.Status.UfsTotal, test.expectedFileNum, dataset.Status.FileNum)
		}
	}
}
