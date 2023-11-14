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

package goosefs

import (
	"fmt"
	"testing"

	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func mockExecCommandInContainerForTotalStorageBytes() (stdout string, stderr string, err error) {
	r := `File Count               Folder Count             Folder Size
	50000                    1000                     6706560319`
	return r, "", nil
}

func mockExecCommandInContainerForTotalFileNums() (stdout string, stderr string, err error) {
	r := `Master.FilesCompleted  (Type: COUNTER, Value: 1,331,167)`
	return r, "", nil
}

func TestUsedStorageBytes(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name:      "test",
			fields:    fields{},
			wantValue: 0,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{}
			gotValue, err := e.UsedStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.UsedStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("GooseFSEngine.UsedStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestFreeStorageBytes(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name:      "test",
			fields:    fields{},
			wantValue: 0,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{}
			gotValue, err := e.FreeStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.FreeStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("GooseFSEngine.FreeStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTotalStorageBytes(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.GooseFSRuntime
		name    string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name: "spark",
					},
				},
			},
			wantValue: 6706560319,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime: tt.fields.runtime,
				name:    tt.fields.name,
			}
			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForTotalStorageBytes()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := e.TotalStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.TotalStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("GooseFSEngine.TotalStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTotalFileNums(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.GooseFSRuntime
		name    string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name: "spark",
					},
				},
			},
			wantValue: 1331167,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime: tt.fields.runtime,
				name:    tt.fields.name,
			}
			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForTotalFileNums()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := e.TotalFileNums()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.TotalFileNums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("GooseFSEngine.TotalFileNums() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestShouldCheckUFS(t *testing.T) {
	tests := []struct {
		name       string
		wantShould bool
		wantErr    bool
	}{
		{
			name:       "test",
			wantShould: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{}
			gotShould, err := e.ShouldCheckUFS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.ShouldCheckUFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("GooseFSEngine.ShouldCheckUFS() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestPrepareUFS(t *testing.T) {
	type fields struct {
		runtime            *datav1alpha1.GooseFSRuntime
		dataset            *datav1alpha1.Dataset
		name               string
		namespace          string
		Log                logr.Logger
		MetadataSyncDoneCh chan base.MetadataSyncResult
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "cosn://imagenet-1234567/",
							},
						},
						DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
							Path:     "local:///tmp/restore",
							NodeName: "192.168.0.1",
						},
					},
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: "",
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.dataset)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.fields.dataset)

			var gfsUtils operations.GooseFSFileUtils
			patch1 := ApplyMethod(reflect.TypeOf(gfsUtils), "Ready", func(_ operations.GooseFSFileUtils) bool {
				return true
			})
			defer patch1.Reset()

			patch2 := ApplyMethod(reflect.TypeOf(gfsUtils), "IsMounted", func(_ operations.GooseFSFileUtils, goosefsPath string) (bool, error) {
				return false, nil
			})
			defer patch2.Reset()

			patch3 := ApplyMethod(reflect.TypeOf(gfsUtils), "Mount", func(_ operations.GooseFSFileUtils, goosefsPath string, ufsPath string, options map[string]string, readOnly bool, shared bool) error {
				return nil
			})
			defer patch3.Reset()

			patch4 := ApplyMethod(reflect.TypeOf(gfsUtils), "QueryMetaDataInfoIntoFile", func(_ operations.GooseFSFileUtils, key operations.KeyOfMetaDataFile, filename string) (string, error) {
				return "10000", nil
			})
			defer patch4.Reset()

			e := &GooseFSEngine{
				runtime:            tt.fields.runtime,
				name:               tt.fields.name,
				namespace:          tt.fields.namespace,
				Log:                tt.fields.Log,
				Client:             mockClient,
				MetadataSyncDoneCh: tt.fields.MetadataSyncDoneCh,
			}
			if err := e.PrepareUFS(); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.PrepareUFS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShouldUpdateUFS(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		dataset   *datav1alpha1.Dataset
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name       string
		fields     fields
		wantAdd    []string
		wantRemove []string
	}{
		{
			name: "test0",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "cosn://imagenet-1234567/",
							},
						},
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantAdd:    []string{"/"},
			wantRemove: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.dataset)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.fields.dataset)

			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    mockClient,
			}

			gotUfsToUpdate := e.ShouldUpdateUFS()
			if !reflect.DeepEqual(gotUfsToUpdate.ToAdd(), tt.wantAdd) {
				t.Errorf("GooseFSEngine.ShouldUpdateUFS add = %v, want %v", gotUfsToUpdate.ToAdd(), tt.wantAdd)
			}

			if !reflect.DeepEqual(gotUfsToUpdate.ToRemove(), tt.wantRemove) {
				t.Errorf("GooseFSEngine.ShouldUpdateUFS() remove = %v, want %v", gotUfsToUpdate.ToRemove(), tt.wantRemove)
			}
		})
	}
}

func TestUpdateOnUFSChange(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		dataset   *datav1alpha1.Dataset
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name            string
		fields          fields
		wantUpdateReady bool
		wantErr         bool
		should          bool
		notMount        bool
		Ready           bool
	}{
		{
			name: "test0",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "cosn://imagenet-1234567/",
							},
						},
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr:         false,
			wantUpdateReady: true,
			should:          true,
			Ready:           true,
		},
		{
			name: "test1",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "default",
					},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: v1.ObjectMeta{
						Name:      "hadoop",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "cosn://imagenet-1234567/",
							},
						},
					},
				},
				name:      "hadoop",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr:         false,
			wantUpdateReady: false,
			should:          false,
			Ready:           true,
		},
		{
			name: "test2",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "hbase",
						Namespace: "default",
					},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: v1.ObjectMeta{
						Name:      "hbase",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "cosn://imagenet-1234567/",
							},
						},
					},
				},
				name:      "hbase",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr:         true,
			wantUpdateReady: false,
			should:          true,
			notMount:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.dataset)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.fields.dataset)

			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    mockClient,
			}

			ufs := utils.NewUFSToUpdate(tt.fields.dataset)
			patch1 := ApplyMethod(reflect.TypeOf(ufs), "ShouldUpdate",
				func(_ *utils.UFSToUpdate) bool {
					return tt.should
				})
			defer patch1.Reset()

			var goosefsFileUtils operations.GooseFSFileUtils
			patch2 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Ready", func(_ operations.GooseFSFileUtils) bool {
				return tt.Ready
			})
			defer patch2.Reset()

			patch3 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Mount", func(_ operations.GooseFSFileUtils, goosefsPath string,
				ufsPath string,
				options map[string]string,
				readOnly bool,
				shared bool) error {
				if tt.notMount {
					return fmt.Errorf("Mount Error")
				} else {
					return nil
				}
			})
			defer patch3.Reset()
			gotUpdateReady, err := e.UpdateOnUFSChange(ufs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.UpdateOnUFSChange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUpdateReady != tt.wantUpdateReady {
				t.Errorf("GooseFSEngine.UpdateOnUFSChange() = %v, want %v", gotUpdateReady, tt.wantUpdateReady)
			}
		})
	}
}
