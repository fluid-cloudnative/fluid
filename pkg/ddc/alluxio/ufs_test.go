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
	"strings"
	"testing"

	"reflect"

	. "github.com/agiledragon/gomonkey"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
			e := &AlluxioEngine{}
			gotValue, err := e.UsedStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.UsedStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.UsedStorageBytes() = %v, want %v", gotValue, tt.wantValue)
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
			e := &AlluxioEngine{}
			gotValue, err := e.FreeStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.FreeStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.FreeStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTotalStorageBytes(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.AlluxioRuntime
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
				runtime: &datav1alpha1.AlluxioRuntime{
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
			e := &AlluxioEngine{
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
				t.Errorf("AlluxioEngine.TotalStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.TotalStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTotalFileNums(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.AlluxioRuntime
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
				runtime: &datav1alpha1.AlluxioRuntime{
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
			e := &AlluxioEngine{
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
				t.Errorf("AlluxioEngine.TotalFileNums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.TotalFileNums() = %v, want %v", gotValue, tt.wantValue)
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
			e := &AlluxioEngine{}
			gotShould, err := e.ShouldCheckUFS()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.ShouldCheckUFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("AlluxioEngine.ShouldCheckUFS() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestPrepareUFS(t *testing.T) {
	type fields struct {
		runtime            *datav1alpha1.AlluxioRuntime
		dataset            *datav1alpha1.Dataset
		name               string
		namespace          string
		Log                logr.Logger
		MetadataSyncDoneCh chan MetadataSyncResult
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{},
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
				Log:       log.NullLogger{},
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

			var afsUtils operations.AlluxioFileUtils
			patch1 := ApplyMethod(reflect.TypeOf(afsUtils), "Ready", func(_ operations.AlluxioFileUtils) bool {
				return true
			})
			defer patch1.Reset()

			patch2 := ApplyMethod(reflect.TypeOf(afsUtils), "IsMounted", func(_ operations.AlluxioFileUtils, AlluxioPath string) (bool, error) {
				return false, nil
			})
			defer patch2.Reset()

			patch3 := ApplyMethod(reflect.TypeOf(afsUtils), "Mount", func(_ operations.AlluxioFileUtils, alluxioPath string, ufsPath string, options map[string]string, readOnly bool, shared bool) error {
				return nil
			})
			defer patch3.Reset()

			patch4 := ApplyMethod(reflect.TypeOf(afsUtils), "QueryMetaDataInfoIntoFile", func(_ operations.AlluxioFileUtils, key operations.KeyOfMetaDataFile, filename string) (string, error) {
				return "10000", nil
			})
			defer patch4.Reset()

			e := &AlluxioEngine{
				runtime:            tt.fields.runtime,
				name:               tt.fields.name,
				namespace:          tt.fields.namespace,
				Log:                tt.fields.Log,
				Client:             mockClient,
				MetadataSyncDoneCh: tt.fields.MetadataSyncDoneCh,
			}
			if err := e.PrepareUFS(); (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.PrepareUFS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindUnmountedUFS(t *testing.T) {
	
	type fields struct {
		mountPoints          []datav1alpha1.Mount
		wantedUnmountedPaths []string
	}
	
	tests := []fields {
		{
			mountPoints:           []datav1alpha1.Mount{
				{
					MountPoint: "s3://bucket/path/train",
					Path: "/path1",
				},
			},
			wantedUnmountedPaths:  []string{"/path1"},
		},
		{
			mountPoints:           []datav1alpha1.Mount{
				{
					MountPoint: "local://mnt/test",
					Path: "/path2",
				},
			},
			wantedUnmountedPaths:  []string{},
		},
		{
			mountPoints:          []datav1alpha1.Mount{
				{
					MountPoint: "s3://bucket/path/train",
					Path: "/path1",
				},
				{
					MountPoint: "local://mnt/test",
					Path: "/path2",
				},
				{
					MountPoint: "hdfs://endpoint/path/train",
					Path: "/path3",
				},
			},
			wantedUnmountedPaths:  []string{"/path1", "/path3"},
		},
	}

	for index, test := range tests {
		t.Run("test", func(t *testing.T) {
			s := runtime.NewScheme()
			runtime := datav1alpha1.AlluxioRuntime{}
			dataset := datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: test.mountPoints,
				},
			}

			s.AddKnownTypes(datav1alpha1.GroupVersion, &runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, &dataset)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, &runtime, &dataset)

			var afsUtils operations.AlluxioFileUtils
			patch1 := ApplyMethod(reflect.TypeOf(afsUtils), "Ready", func(_ operations.AlluxioFileUtils) bool {
				return true
			})
			defer patch1.Reset()

			patch2 := ApplyMethod(reflect.TypeOf(afsUtils), "FindUnmountedAlluxioPaths", func(_ operations.AlluxioFileUtils, alluxioPaths []string) ([]string, error) {
				return alluxioPaths, nil
			})
			defer patch2.Reset()

			e := &AlluxioEngine{
				runtime:            &runtime,
				name:               "test",
				namespace:          "default",
				Log:                log.NullLogger{},
				Client:             mockClient,
				MetadataSyncDoneCh: nil,
			}

			unmountedPaths, err := e.FindUnmountedUFS()
			if err != nil{
				t.Errorf("AlluxioEngine.FindUnmountedUFS() error = %v", err)
				return
			}
			if (len(unmountedPaths) != 0 || len(test.wantedUnmountedPaths) != 0 ) && 
			       !reflect.DeepEqual(unmountedPaths,test.wantedUnmountedPaths) {
				t.Errorf("%d check failure, want: %s, got: %s", index, strings.Join(test.wantedUnmountedPaths, ","), strings.Join(unmountedPaths, ","))
				return
		}
		})
	}
}