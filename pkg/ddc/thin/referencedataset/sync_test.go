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

package referencedataset

import (
	"context"
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestReferenceDatasetEngine_Sync(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	testObjs := []runtime.Object{}

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "big-data",
		},
		Status: datav1alpha1.DatasetStatus{
			Runtimes: []datav1alpha1.Runtime{
				{
					Name:      "done",
					Namespace: "big-data",
					Type:      common.AlluxioRuntime,
				},
			},
			DatasetRef: []string{
				"fluid/hbase",
				"fluid/test",
			},
			UfsTotal: "100Gi",
		},
	}
	var runtime = datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done",
			Namespace: "big-data",
		},
	}

	var refRuntime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	var refDataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
			},
		},
	}

	testObjs = append(testObjs, &dataset, &refDataset)

	testObjs = append(testObjs, &runtime, &refRuntime)
	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type fields struct {
		Id                string
		Client            client.Client
		Log               logr.Logger
		name              string
		namespace         string
		syncRetryDuration time.Duration
		timeOfLastSync    time.Time
	}
	type args struct {
		ctx cruntime.ReconcileRequestContext
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "sync",
			fields: fields{
				Id:                "reference-engine",
				Client:            fakeClient,
				Log:               fake.NullLogger(),
				name:              refRuntime.GetName(),
				namespace:         refRuntime.GetNamespace(),
				timeOfLastSync:    time.Now().Add(-defaultSyncRetryDuration),
				syncRetryDuration: defaultSyncRetryDuration,
			},
		},
	}
	for _, tt := range tests {
		e := &ReferenceDatasetEngine{
			Id:                tt.fields.Id,
			Client:            tt.fields.Client,
			Log:               tt.fields.Log,
			name:              tt.fields.name,
			namespace:         tt.fields.namespace,
			syncRetryDuration: tt.fields.syncRetryDuration,
			timeOfLastSync:    tt.fields.timeOfLastSync,
		}
		if err := e.Sync(tt.args.ctx); (err != nil) != tt.wantErr {
			t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		updatedRefDataset := &datav1alpha1.Dataset{}
		err := fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: refDataset.Namespace, Name: refDataset.Name,
		}, updatedRefDataset)
		if err != nil {
			t.Errorf("Get dataset error %v", err)
			return
		}
		// check updated status
		if updatedRefDataset.Status.UfsTotal != dataset.Status.UfsTotal || len(updatedRefDataset.Status.DatasetRef) != 0 {
			t.Errorf("Dataset status is not updated")
			return
		}

		boundRuntimes := updatedRefDataset.Status.Runtimes
		if len(boundRuntimes) != 1 {
			t.Errorf("Dataset is not bound runtime")
			return
		}
		boundRuntime := boundRuntimes[0]
		if boundRuntime.Type != common.ThinRuntime || boundRuntime.Name != refRuntime.Name ||
			boundRuntime.Namespace != refRuntime.Namespace {
			t.Errorf("Dataset bound runtime info wrong %v", boundRuntime)
		}

	}
}
