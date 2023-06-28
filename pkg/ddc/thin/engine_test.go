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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuild(t *testing.T) {
	var namespace = v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, namespace.DeepCopy())

	var dataset = datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs = append(testObjs, dataset.DeepCopy())
	var runtimeProfile = datav1alpha1.ThinRuntimeProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-profile",
		},
		Spec: datav1alpha1.ThinRuntimeProfileSpec{FileSystemType: "test-fstype"},
	}
	testObjs = append(testObjs, runtimeProfile.DeepCopy())

	var runtime = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.ThinRuntimeSpec{
			ThinRuntimeProfileName: "test-profile",
			Fuse:                   datav1alpha1.ThinFuseSpec{},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
	testObjs = append(testObjs, runtime.DeepCopy())
	var runtime2 = datav1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.ThinRuntimeSpec{
			ThinRuntimeProfileName: "test-profile",
			Fuse:                   datav1alpha1.ThinFuseSpec{},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}

	var sts = appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",
			Namespace: "fluid",
		},
	}
	testObjs = append(testObjs, sts.DeepCopy())
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	var ctx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "thin",
		Runtime:     &runtime,
	}

	engine, err := Build("testId", ctx)
	if err != nil || engine == nil {
		t.Errorf("fail to exec the build function with the eror %v", err)
	}

	var errCtx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "thin",
		Runtime:     nil,
	}

	got, err := Build("testId", errCtx)
	if err == nil {
		t.Errorf("expect err, but no err got %v", got)
	}

	var errrCtx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "thin",
		Runtime:     &runtime2,
	}

	gott, err := Build("testId", errrCtx)
	if err == nil {
		t.Errorf("expect err, but no err got %v", gott)
	}
}

func TestBuildReferenceDatasetEngine(t *testing.T) {
	var namespace = v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, namespace.DeepCopy())

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

	testObjs = append(testObjs, &dataset, &refDataset)

	testObjs = append(testObjs, &runtime, &refRuntime)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	var ctx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "thin",
		Runtime:     &refRuntime,
	}

	engine, err := Build("testId", ctx)
	if err != nil || engine == nil {
		t.Errorf("fail to exec the build function with the eror %v", err)
	}

	var errCtx = cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: "thin",
		Runtime:     &runtime,
	}

	got, err := Build("testId", errCtx)
	if err == nil {
		t.Errorf("expect err, but no err got %v", got)
	}
}

func TestCheckReferenceDatasetRuntime(t *testing.T) {
	tests := []struct {
		name    string
		dataset *datav1alpha1.Dataset
		runtime *datav1alpha1.ThinRuntime
		ctx     cruntime.ReconcileRequestContext
		want    bool
		wantErr bool
	}{
		{
			name: "ref-dataset",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://test/test",
						},
					},
				},
			},
			runtime: &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.ThinRuntimeSpec{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not-ref-dataset",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			},
			runtime: &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					ThinRuntimeProfileName: "",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "dataset-not-exist",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-no-use",
					Namespace: "fluid",
				},
			},
			runtime: &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					ThinRuntimeProfileName: "1",
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "dataset-not-exist-but-get-physical-dataset-from-runtime",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-no-use",
					Namespace: "fluid",
				},
			},
			runtime: &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					ThinRuntimeProfileName: "1",
				},
				Status: datav1alpha1.RuntimeStatus{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "dataset://ns-a/n-a",
					}},
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewFakeClientWithScheme(testScheme, tt.dataset, tt.runtime)
			var ctx = cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Client:      fakeClient,
				Log:         fake.NullLogger(),
				RuntimeType: "thin",
				Runtime:     tt.runtime,
			}
			isRef, err := CheckReferenceDatasetRuntime(ctx, tt.runtime)

			if (err != nil) != tt.wantErr {
				t.Errorf("expect has error %t, but get error %v", tt.wantErr, err)
				return
			}

			if isRef != tt.want {
				t.Errorf(" expect is ref dataset %t, but get %t", tt.want, err)
			}
		})
	}
}
