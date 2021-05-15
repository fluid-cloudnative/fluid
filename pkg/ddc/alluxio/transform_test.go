/*

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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestTransformFuse(t *testing.T) {

	var x int64 = 1000
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		dataset *datav1alpha1.Dataset
		value   *Alluxio
		expect  []string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Fuse: datav1alpha1.AlluxioFuseSpec{},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{datav1alpha1.Mount{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
				Owner: &datav1alpha1.User{
					UID: &x,
					GID: &x,
				},
			},
		}, &Alluxio{}, []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty,uid=1000,gid=1000,allow_other"}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.Log = ctrl.Log
		err := engine.transformFuse(test.runtime, test.dataset, test.value)
		if err != nil {
			t.Errorf("error %v", err)
		}
		if test.value.Fuse.Args[1] != test.expect[1] {
			t.Errorf("expected %v, got %v", test.expect, test.value.Fuse.Args)
		}
	}
}

func TestAlluxioEngine_transformJournal(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.AlluxioRuntime
	}
	type args struct {
		value *Alluxio
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test0",
			fields: fields{
				&datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							Replicas: 3,
							Properties: map[string]string{
								"alluxio.master.journal.type": "EMBEDDED",
							},
							Journal: datav1alpha1.Journal{
								VolumeType: "memory",
							},
						},
					},
				},
			},
			args: args{
				value: &Alluxio{
					Journal: Journal{
						Type:       JOURNAL_TYPE_EMBEDDED,
						VolumeType: "emptyDir",
						Size:       "30Gi",
						Format:     Format{RunFormat: true},
					},
				},
			},
		},
		{
			name: "test1",
			fields: fields{
				&datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							Journal: datav1alpha1.Journal{
								VolumeType:   "pvc",
								StorageClass: "standard",
							},
							Replicas: 3,
						},
						Properties: map[string]string{
							"alluxio.master.journal.type": "EMBEDDED",
						},
					},
				},
			},
			args: args{
				value: &Alluxio{
					Journal: Journal{
						Type:         JOURNAL_TYPE_EMBEDDED,
						VolumeType:   "persistentVolumeClaim",
						StorageClass: "standard",
						Format:       Format{RunFormat: true},
					},
				},
			},
		},
		{
			name: "test2",
			fields: fields{
				&datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				},
			},
			args: args{
				value: &Alluxio{
					Journal: Journal{
						Type:       JOURNAL_TYPE_UFS,
						VolumeType: "emptyDir",
						Size:       "30Gi",
						UFSType:    "local",
						Format:     Format{RunFormat: false},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime: tt.fields.runtime,
			}
			value := &Alluxio{}
			e.transformJournal(e.runtime, value)
			if !reflect.DeepEqual(value, tt.args.value) {
				t.Errorf("transformJournal() value = %v, want %v", value, tt.args.value)
			}
		})
	}
}
