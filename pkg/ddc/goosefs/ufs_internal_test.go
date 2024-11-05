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
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func mockGooseFSFileUtilsCount() (value string, err error) {
	r := `File Count               Folder Count             Folder Size
	6                        1                        571808905`
	return r, nil
}

func TestGooseFSEngine_calculateMountPointsChanges(t *testing.T) {

	testCases := map[string]struct {
		mounted []string
		current []string
		expect  map[string][]string
	}{
		"calculate mount point changes test case 1": {
			mounted: []string{"hadoop3.3.0"},
			current: []string{"hadoopcurrent", "hadoop3.3.0"},
			expect:  map[string][]string{"added": {"hadoopcurrent"}, "removed": {}},
		},
		"calculate mount point changes test case 2": {
			mounted: []string{"hadoopcurrent", "hadoop3.3.0"},
			current: []string{"hadoop3.3.0"},
			expect:  map[string][]string{"added": {}, "removed": {"hadoopcurrent"}},
		},
		"calculate mount point changes test case 3": {
			mounted: []string{"hadoopcurrent", "hadoop3.2.2"},
			current: []string{"hadoop3.3.0", "hadoop3.2.2"},
			expect:  map[string][]string{"added": {"hadoop3.3.0"}, "removed": {"hadoopcurrent"}},
		},
		"calculate mount point changes test case 4": {
			mounted: []string{"hadoop3.3.0"},
			current: []string{"hadoop3.3.0"},
			expect:  map[string][]string{"added": {}, "removed": {}},
		},
		"calculate mount point changes test case 5": {
			mounted: []string{"hadoopcurrent", "hadoop3.2.2"},
			current: []string{"hadoop3.3.0", "hadoop3.2.2", "hadoop3.3.1"},
			expect:  map[string][]string{"added": {"hadoop3.3.0", "hadoop3.3.1"}, "removed": {"hadoopcurrent"}},
		},
	}

	for _, item := range testCases {
		engine := &GooseFSEngine{}
		added, removed := engine.calculateMountPointsChanges(item.mounted, item.current)

		if !ArrayEqual(added, item.expect["added"]) {
			t.Errorf("expected added %v, got %v", item.expect["added"], added)
		}
		if !ArrayEqual(removed, item.expect["removed"]) {
			t.Errorf("expected removed %v, got %v", item.expect["removed"], removed)
		}
	}

}

func ArrayEqual(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for key, val := range a {
		if val != b[key] {
			return false
		}
	}
	return true
}

func TestUsedStorageBytesInternal(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
		Log       logr.Logger
		Client    client.Client
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "todo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    tt.fields.Client,
			}
			gotValue, err := e.usedStorageBytesInternal()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.usedStorageBytesInternal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("GooseFSEngine.usedStorageBytesInternal() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestFreeStorageBytesInternal(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
		Log       logr.Logger
		Client    client.Client
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "todo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    tt.fields.Client,
			}
			gotValue, err := e.freeStorageBytesInternal()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.freeStorageBytesInternal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("GooseFSEngine.freeStorageBytesInternal() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

func TestTotalStorageBytesInternal(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name      string
		fields    fields
		wantTotal int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				name:      "spark",
				namespace: "defaut",
				Log:       fake.NullLogger(),
			},
			wantTotal: 571808905,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}

			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				summary, err := mockGooseFSFileUtilsCount()
				return summary, "", err
			})
			defer patch1.Reset()
			gotTotal, err := e.totalStorageBytesInternal()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.totalStorageBytesInternal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("GooseFSEngine.totalStorageBytesInternal() = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestTotalFileNumsInternal(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name          string
		fields        fields
		wantFileCount int64
		wantErr       bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				name:      "spark",
				namespace: "defaut",
				Log:       fake.NullLogger(),
			},
			wantFileCount: 6,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}

			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				summary, err := mockGooseFSFileUtilsCount()
				return summary, "", err
			})
			defer patch1.Reset()
			gotFileCount, err := e.totalFileNumsInternal()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.totalFileNumsInternal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFileCount != tt.wantFileCount {
				t.Errorf("GooseFSEngine.totalFileNumsInternal() = %v, want %v", gotFileCount, tt.wantFileCount)
			}
		})
	}
}

func TestShouldMountUFS(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		dataset   *datav1alpha1.Dataset
		name      string
		namespace string
		Log       logr.Logger
		Client    client.Client
	}
	tests := []struct {
		name       string
		fields     fields
		wantShould bool
		wantErr    bool
	}{
		{
			name: "test",
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
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantShould: false,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.runtime, tt.fields.dataset)
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				summary := mockGooseFSReportSummary()
				return summary, "", nil
			})
			defer patch1.Reset()
			gotShould, err := e.shouldMountUFS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.shouldMountUFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("GooseFSEngine.shouldMountUFS() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

func TestGetMounts(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		dataset   *datav1alpha1.Dataset
		name      string
		namespace string
		Log       logr.Logger
		Client    client.Client
	}
	tests := []struct {
		name                  string
		fields                fields
		wantResultInCtx       []string
		wantResultHaveMounted []string
		wantErr               bool
	}{
		{
			name: "test",
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
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/spec",
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/spec",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/status",
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/status",
							},
						},
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantResultInCtx:       []string{"/spec", "/spec"},
			wantResultHaveMounted: []string{"/status", "/status"},
			wantErr:               false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.runtime, tt.fields.dataset)
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			var goosefsFileUtils operations.GooseFSFileUtils
			patch1 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Ready", func(_ operations.GooseFSFileUtils) bool {
				return true
			})
			defer patch1.Reset()

			gotResultInCtx, gotResultHaveMounted, err := e.getMounts()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.getMounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResultInCtx, tt.wantResultInCtx) {
				t.Errorf("GooseFSEngine.getMounts() gotResultInCtx = %v, want %v", gotResultInCtx, tt.wantResultInCtx)
			}
			if !reflect.DeepEqual(gotResultHaveMounted, tt.wantResultHaveMounted) {
				t.Errorf("GooseFSEngine.getMounts() gotResultHaveMounted = %v, want %v", gotResultHaveMounted, tt.wantResultHaveMounted)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	type args struct {
		s []string
		v string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{
				s: []string{"1", "2"},
				v: "1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsString(tt.args.s, tt.args.v); got != tt.want {
				t.Errorf("ContainsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessUpdatingUFS(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		dataset   *datav1alpha1.Dataset
		name      string
		namespace string
		Log       logr.Logger
		Client    client.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
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
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/spec",
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/spec",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/status",
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/status",
							},
						},
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr: false,
		},
		{
			name: "test1",
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
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/spec",
								Options: map[string]string{"fs.cosn.bucket.region": "ap-shanghai",
									"fs.cosn.impl":                    "org.apache.hadoop.fs.CosFileSystem",
									"fs.AbstractFileSystem.cosn.impl": "org.apache.hadoop.fs.CosN",
									"fs.cos.app.id":                   "1251707795"},
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "access-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test0",
												Key:  "access-key",
											}},
									}, {
										Name: "secret-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test0",
												Key:  "secret-key",
											}},
									}, {
										Name: "metaurl",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test0",
												Key:  "metaurl",
											}},
									},
								},
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/spec",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/status",
							},
						},
					},
				},
				name:      "hbase",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr: false,
		},
		{
			name: "test2",
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
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/spec",
								Options: map[string]string{"fs.cosn.bucket.region": "ap-shanghai",
									"fs.cosn.impl":                    "org.apache.hadoop.fs.CosFileSystem",
									"fs.AbstractFileSystem.cosn.impl": "org.apache.hadoop.fs.CosN",
									"fs.cos.app.id":                   "1251707795"},
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "access-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test0",
												Key:  "access-key",
											}},
									}, {
										Name: "secret-key",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test0",
												Key:  "secret-key",
											}},
									}, {
										Name: "metaurl",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test0",
												Key:  "metaurl",
											}},
									},
								},
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/spec",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/status",
							},
							{
								Name:       "test1",
								MountPoint: "cos://test1",
								Path:       "/status",
							},
							{
								Name:       "test2",
								MountPoint: "cos://test2",
								Path:       "/status",
							},
						},
					},
				},
				name:      "hadoop",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test0",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"fs.cosn.userinfo.secretKey": []byte("key"),
					"fs.cosn.userinfo.secretId":  []byte("id"),
				},
			}
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.runtime, tt.fields.dataset, &secret)
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			var goosefsFileUtils operations.GooseFSFileUtils
			patch1 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Ready", func(_ operations.GooseFSFileUtils) bool {
				return true
			})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Mount", func(_ operations.GooseFSFileUtils, goosefsPath string,
				ufsPath string,
				options map[string]string,
				readOnly bool,
				shared bool) error {
				return nil
			})
			defer patch2.Reset()
			patch3 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "UnMount", func(_ operations.GooseFSFileUtils, goosefsPath string) error {
				return nil
			})
			defer patch3.Reset()
			patch4 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "LoadMetadataWithoutTimeout", func(_ operations.GooseFSFileUtils, goosefsPath string) error {
				return nil
			})
			defer patch4.Reset()

			ufs := e.ShouldUpdateUFS()
			err := utils.UpdateMountStatus(client, tt.fields.name, tt.fields.namespace, datav1alpha1.UpdatingDatasetPhase)
			if err != nil {
				t.Error("GooseFSEngine.UpdateMountStatus()")
			}
			if err := e.processUpdatingUFS(ufs); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.processUpdatingUFS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMountUFS(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		dataset   *datav1alpha1.Dataset
		name      string
		namespace string
		Log       logr.Logger
		Client    client.Client
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test",
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
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/spec",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test0",
								MountPoint: "cos://test0",
								Path:       "/status",
							},
						},
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
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.runtime, tt.fields.dataset)
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			var goosefsFileUtils operations.GooseFSFileUtils
			patch1 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Ready", func(_ operations.GooseFSFileUtils) bool {
				return true
			})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "Mount", func(_ operations.GooseFSFileUtils, goosefsPath string,
				ufsPath string,
				options map[string]string,
				readOnly bool,
				shared bool) error {
				return nil
			})
			defer patch2.Reset()
			patch3 := ApplyMethod(reflect.TypeOf(goosefsFileUtils), "IsMounted", func(_ operations.GooseFSFileUtils, goosefsPath string,
			) (bool, error) {
				return false, nil
			})
			defer patch3.Reset()
			if err := e.mountUFS(); (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.mountUFS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenUFSMountOptions(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
	}
	type args struct {
		m   datav1alpha1.Mount
		pm  map[string]string
		pme []datav1alpha1.EncryptOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			args: args{
				pm: map[string]string{
					"key1": "value1",
				},
				pme: []datav1alpha1.EncryptOption{
					{
						Name: "key2",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "mysecret",
								Key:  "key2",
							},
						},
					},
				},
				m: datav1alpha1.Mount{
					Options: map[string]string{"fs.cosn.bucket.region": "ap-shanghai",
						"fs.cosn.impl":                    "org.apache.hadoop.fs.CosFileSystem",
						"fs.AbstractFileSystem.cosn.impl": "org.apache.hadoop.fs.CosN",
						"fs.cos.app.id":                   "1251707795"},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: "fs.cosn.userinfo.secretKey",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "mysecret",
									Key:  "fs.cosn.userinfo.secretKey",
								},
							},
						},
						{
							Name: "fs.cosn.userinfo.secretId",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "mysecret",
									Key:  "fs.cosn.userinfo.secretId",
								},
							},
						},
					},
				},
			},
			want: map[string]string{"fs.cosn.bucket.region": "ap-shanghai",
				"fs.cosn.impl":                    "org.apache.hadoop.fs.CosFileSystem",
				"fs.AbstractFileSystem.cosn.impl": "org.apache.hadoop.fs.CosN",
				"fs.cos.app.id":                   "1251707795",
				"fs.cosn.userinfo.secretKey":      "key",
				"fs.cosn.userinfo.secretId":       "id",
				"key1":                            "value1",
				"key2":                            "value2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "mysecret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"fs.cosn.userinfo.secretKey": []byte("key"),
					"fs.cosn.userinfo.secretId":  []byte("id"),
					"key2":                       []byte("value2"),
				},
			}
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, &secret)
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			got, err := e.genUFSMountOptions(tt.args.m, tt.args.pm, tt.args.pme)
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.genUFSMountOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GooseFSEngine.genUFSMountOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
