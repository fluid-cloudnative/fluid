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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBuildReferenceDatasetThinEngine(t *testing.T) {
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

	var multipleRefDataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid-mul",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "dataset://big-data/done",
				},
				{
					MountPoint: "http://big-test/done",
				},
			},
		},
	}

	var multipleRefRuntime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid-mul",
		},
	}

	testObjs = append(testObjs, &dataset, &refDataset, &multipleRefDataset)
	testObjs = append(testObjs, &runtime, &refRuntime, &multipleRefRuntime)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	testcases := []struct {
		name    string
		ctx     cruntime.ReconcileRequestContext
		wantErr bool
	}{
		{
			name: "success",
			ctx: cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Client:      client,
				Log:         fake.NullLogger(),
				RuntimeType: common.ThinRuntime,
				Runtime:     &refRuntime,
			},
			wantErr: false,
		},
		{
			name: "dataset-not-ref",
			ctx: cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "done",
					Namespace: "big-data",
				},
				Client:      client,
				Log:         fake.NullLogger(),
				RuntimeType: common.ThinRuntime,
				Runtime:     &runtime,
			},
			wantErr: true,
		},
		{
			name: "dataset-with-different-format",
			ctx: cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "hbase",
					Namespace: "fluid-mul",
				},
				Client:      client,
				Log:         fake.NullLogger(),
				RuntimeType: common.ThinRuntime,
				Runtime:     &multipleRefRuntime,
			},
			wantErr: true,
		},
	}

	for _, testcase := range testcases {
		_, err := BuildReferenceDatasetThinEngine(testcase.name, testcase.ctx)
		hasError := err != nil
		if testcase.wantErr != hasError {
			t.Errorf("expect error %t, get error %v", testcase.wantErr, err)
		}
	}
}

func TestReferenceDatasetEngine_Setup(t *testing.T) {
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
				"fluid/test",
			},
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
	var configCM = v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtime.Name + "-config",
			Namespace: runtime.Namespace,
		},
		Data: map[string]string{
			"check.sh": "/bin/sh check",
		},
	}

	var fuseDs = appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "done-fuse",
			Namespace: "big-data",
		},
		Spec: appsv1.DaemonSetSpec{},
	}

	testObjs = append(testObjs, &dataset, &refDataset, &configCM, &runtime, &refRuntime, &fuseDs)

	fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	type fields struct {
		Id                string
		Client            client.Client
		Log               logr.Logger
		name              string
		namespace         string
		syncRetryDuration time.Duration
		timeOfLastSync    time.Time
		runtimeType       string
	}
	tests := []struct {
		name      string
		fields    fields
		ctx       cruntime.ReconcileRequestContext
		wantReady bool
		wantErr   bool
		wantCMs   int
	}{
		{
			name: "setup",
			fields: fields{
				Client:    fakeClient,
				name:      refRuntime.Name,
				namespace: refRuntime.Namespace,
			},
			ctx: cruntime.ReconcileRequestContext{
				Dataset: &refDataset,
				Client:  fakeClient,
			},
			wantReady: true,
			wantErr:   false,
			wantCMs:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ReferenceDatasetEngine{
				Id:                tt.fields.Id,
				Client:            tt.fields.Client,
				Log:               tt.fields.Log,
				name:              tt.fields.name,
				namespace:         tt.fields.namespace,
				syncRetryDuration: tt.fields.syncRetryDuration,
				timeOfLastSync:    tt.fields.timeOfLastSync,
				runtimeType:       tt.fields.runtimeType,
			}
			gotReady, err := e.Setup(tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("Setup() gotReady = %v, want %v", gotReady, tt.wantReady)
			}
			if gotReady {
				updatedDataset := &datav1alpha1.Dataset{}
				err := fakeClient.Get(context.TODO(), types.NamespacedName{
					Namespace: runtime.Namespace, Name: runtime.Name,
				}, updatedDataset)
				if err != nil {
					t.Errorf("Get dataset error %v", err)
					return
				}
				if !utils.ContainsString(updatedDataset.Status.DatasetRef, base.GetDatasetRefName(e.name, e.namespace)) {
					t.Errorf("Setup() not add dataset field DatasetRef")
				}
				cmList := &v1.ConfigMapList{}
				err = fakeClient.List(context.TODO(), cmList, &client.ListOptions{Namespace: e.namespace})
				if err != nil {
					t.Errorf("Get dataset error %v", err)
					return
				}
				items := len(cmList.Items)
				if items != tt.wantCMs {
					t.Errorf("copy configmap wrong, expect %d, but got %d", tt.wantCMs, items)
				}
			}
		})
	}
}

func TestReferenceDatasetEngine_Shutdown(t *testing.T) {
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
		runtimeType       string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "shutdown",
			fields: fields{
				Client:    fakeClient,
				name:      refRuntime.Name,
				namespace: refRuntime.Namespace,
			},
			wantErr: false,
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
			runtimeType:       tt.fields.runtimeType,
		}
		if err := e.Shutdown(); (err != nil) != tt.wantErr {
			t.Errorf("Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		updatedDataset := &datav1alpha1.Dataset{}
		// physicalRuntimeInfo is calculated in Shutdown
		err := fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: e.physicalRuntimeInfo.GetNamespace(), Name: e.physicalRuntimeInfo.GetName(),
		}, updatedDataset)
		if err != nil {
			t.Errorf("Get dataset error %v", err)
			return
		}
		if utils.ContainsString(updatedDataset.Status.DatasetRef, base.GetDatasetRefName(e.name, e.namespace)) {
			t.Errorf("Shutdown() not remove dataset field DatasetRef")
		}
	}
}
