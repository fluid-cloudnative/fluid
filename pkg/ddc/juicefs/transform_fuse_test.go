/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"encoding/base64"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformFuse(t *testing.T) {
	juicefsSecret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"metaurl":    []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	juicefsSecret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret1).DeepCopy(), juicefsSecret2.DeepCopy())

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
	}

	var tests = []struct {
		name         string
		runtime      *datav1alpha1.JuiceFSRuntime
		dataset      *datav1alpha1.Dataset
		juicefsValue *JuiceFS
		expect       string
		wantErr      bool
	}{
		{
			name: "test-secret-right",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
				}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test1",
						Options: map[string]string{
							"storage": "test1",
							"bucket":  "test1",
						},
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "access-key",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "test1",
										Key:  "access-key",
									}},
							}, {
								Name: "secret-key",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "test1",
										Key:  "secret-key",
									}},
							}, {
								Name: "metaurl",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "test1",
										Key:  "metaurl",
									}},
							}},
					}},
				}},
			juicefsValue: &JuiceFS{},
			expect:       "",
			wantErr:      false,
		},
		{
			name: "test-secret-wrong-1",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{Fuse: datav1alpha1.JuiceFSFuseSpec{}}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test2",
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "test1",
									Key:  "metaurl",
								},
							},
						}},
					}},
				}}, juicefsValue: &JuiceFS{}, expect: "", wantErr: true,
		},
		{
			name: "test-secret-wrong-2",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
				}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test3",
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "test1",
									Key:  "metaurl",
								},
							},
						}},
					}}}},
			juicefsValue: &JuiceFS{},
			expect:       "",
			wantErr:      true,
		},
		{
			name: "test-options",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: "MEM",
							Path:       "/data",
							Low:        "0.7",
						}},
					},
				}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test2",
						Options:    map[string]string{"debug": ""},
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "test1",
									Key:  "metaurl",
								},
							},
						}},
					}}}},
			juicefsValue: &JuiceFS{},
			expect:       "",
			wantErr:      false,
		},
		{
			name:    "test-no-mount",
			runtime: &datav1alpha1.JuiceFSRuntime{},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-no-mount",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
			juicefsValue: &JuiceFS{},
			expect:       "",
			wantErr:      true,
		},
		{
			name:    "test-no-secret",
			runtime: &datav1alpha1.JuiceFSRuntime{},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test2",
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "not-exist",
									Key:  "metaurl",
								},
							},
						}},
					}},
				}}, juicefsValue: &JuiceFS{}, expect: "", wantErr: true,
		},
		{
			name:    "test-no-metaurl",
			runtime: &datav1alpha1.JuiceFSRuntime{},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test",
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "no-metaurl",
									Key:  "metaurl",
								},
							},
						}},
					}},
				}}, juicefsValue: &JuiceFS{}, expect: "", wantErr: true,
		},
		{
			name: "test-tiredstore",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: "SSD",
							Path:       "/data",
							Low:        "0.7",
							Quota:      resource.NewQuantity(10, resource.BinarySI),
						}},
					},
				}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test",
						Options:    map[string]string{"debug": ""},
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "test1",
									Key:  "metaurl",
								},
							},
						}},
					}}}},
			juicefsValue: &JuiceFS{},
			expect:       "",
			wantErr:      false,
		},
	}
	for _, test := range tests {
		err := engine.transformFuse(test.runtime, test.dataset, test.juicefsValue)
		if (err != nil) && !test.wantErr {
			t.Errorf("Got err %v", err)
		}
	}
}

func TestJuiceFSEngine_genValue(t *testing.T) {
	juicefsSecret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-community",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"metaurl":    []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	juicefsSecret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-enterprise",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"token":      []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"access-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
			"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("test"))),
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret1).DeepCopy(), juicefsSecret2.DeepCopy())

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
	}
	type fields struct {
		runtime     *datav1alpha1.JuiceFSRuntime
		name        string
		namespace   string
		runtimeType string
	}
	type args struct {
		mount           datav1alpha1.Mount
		tiredStoreLevel *datav1alpha1.Level
		value           *JuiceFS
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantValue *JuiceFS
	}{
		{
			name: "test",
			fields: fields{
				runtime:     nil,
				name:        "test",
				namespace:   "fluid",
				runtimeType: common.JuiceFSRuntime,
			},
			args: args{
				mount: datav1alpha1.Mount{
					MountPoint: "juicefs:///",
					Options:    nil,
					Name:       "test",
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name: "token",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "test-enterprise",
								Key:  "token",
							},
						},
					}},
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: "SSD",
					Path:       "/dev",
				},
				value: &JuiceFS{
					FullnameOverride: "test",
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			wantErr: false,
			wantValue: &JuiceFS{
				FullnameOverride: "test",
				Fuse: Fuse{
					SubPath:       "/",
					TokenSecret:   "test-enterprise",
					MountPath:     "/juicefs/fluid/test/juicefs-fuse",
					CacheDir:      "/dev",
					HostMountPath: "/juicefs/fluid/test",
					//Command:       "/sbin/mount.juicefs test /juicefs/fluid/test/juicefs-fuse -o subdir=/,cache-dir=/dev,foreground,cache-group=test,no-sharing",
					//StatCmd:       "stat -c %i /juicefs/fluid/test/juicefs-fuse",
					//FormatCmd:     "/usr/bin/juicefs auth --token=${TOKEN} test",
				},
				Worker: Worker{
					//Command: "/sbin/mount.juicefs test /juicefs/fluid/test/juicefs-fuse -o subdir=/,cache-dir=/dev,foreground,cache-group=test",
				},
			},
		},
		{
			name: "test-community",
			fields: fields{
				name:        "test-community",
				namespace:   "fluid",
				runtimeType: common.JuiceFSRuntime,
			},
			args: args{
				mount: datav1alpha1.Mount{
					MountPoint: "juicefs:///",
					Options:    map[string]string{},
					Name:       "test-community",
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name: JuiceMetaUrl,
						ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: "test-community",
							Key:  "metaurl",
						}}}, {
						Name: JuiceAccessKey,
						ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: "test-community",
							Key:  "access-key",
						}}}, {
						Name: JuiceSecretKey,
						ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: "test-community",
							Key:  "secret-key",
						}}},
					},
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: "SSD",
					Path:       "/dev",
				},
				value: &JuiceFS{
					FullnameOverride: "test-community",
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			wantErr: false,
			wantValue: &JuiceFS{
				Fuse: Fuse{
					SubPath:       "/",
					TokenSecret:   "test-enterprise",
					MountPath:     "/juicefs/fluid/test/juicefs-fuse",
					CacheDir:      "/dev",
					HostMountPath: "/juicefs/fluid/test",
				},
				Worker: Worker{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := engine.genValue(tt.args.mount, tt.args.tiredStoreLevel, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("genMount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantValue != nil {
				if tt.wantValue.Fuse.Command != tt.args.value.Fuse.Command &&
					tt.wantValue.Fuse.FormatCmd != tt.args.value.Fuse.FormatCmd &&
					tt.wantValue.Worker.Command != tt.args.value.Worker.Command {
					t.Errorf("genMount() got = %v, want = %v", tt.args.value, tt.wantValue)
				}
			}
		})
	}
}

func TestJuiceFSEngine_genMount(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
	}
	type args struct {
		value   *JuiceFS
		options []string
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantErr           bool
		wantWorkerCommand string
		wantFuseCommand   string
		wantStatCmd       string
	}{
		{
			name: "test-community",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-community",
					Edition:          "community",
					Source:           "redis://127.0.0.1:6379",
					Fuse: Fuse{
						SubPath:         "/",
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
						MountPath:       "/test",
						CacheDir:        "/cache",
						HostMountPath:   "/test",
					},
				},
			},
			wantErr:           false,
			wantWorkerCommand: "/bin/mount.juicefs redis://127.0.0.1:6379 /test -o metrics=0.0.0.0:9567",
			wantFuseCommand:   "/bin/mount.juicefs redis://127.0.0.1:6379 /test -o metrics=0.0.0.0:9567",
			wantStatCmd:       "stat -c %i /test",
		},
		{
			name: "test-enterprise",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-enterprise",
					Edition:          "enterprise",
					Source:           "test-enterprise",
					Fuse: Fuse{
						SubPath:         "/",
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
						MountPath:       "/test",
						CacheDir:        "/cache",
						HostMountPath:   "/test",
					},
				},
			},
			wantErr:           false,
			wantWorkerCommand: "/sbin/mount.juicefs test-enterprise /test -o foreground,cache-group=test-enterprise",
			wantFuseCommand:   "/sbin/mount.juicefs test-enterprise /test -o foreground,cache-group=test-enterprise,no-sharing",
			wantStatCmd:       "stat -c %i /test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}
			if err := j.genMount(tt.args.value, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("genMount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.value.Fuse.Command != tt.wantFuseCommand ||
				tt.args.value.Fuse.StatCmd != tt.wantStatCmd ||
				tt.args.value.Worker.Command != tt.wantWorkerCommand {
				t.Errorf("genMount() value = %v", tt.args.value)
			}
		})
	}
}

func TestJuiceFSEngine_genFormat(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
	}
	type args struct {
		value   *JuiceFS
		options []string
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantErr           bool
		wantWorkerCommand string
		wantFuseCommand   string
		wantStatCmd       string
	}{
		{
			name: "test-community",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-community",
					Edition:          "community",
					Source:           "redis://127.0.0.1:6379",
					Fuse: Fuse{
						SubPath:         "/",
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
						MountPath:       "/test",
						CacheDir:        "/cache",
						HostMountPath:   "/test",
					},
				},
			},
			wantErr:           false,
			wantWorkerCommand: "/bin/mount.juicefs redis://127.0.0.1:6379 /test -o metrics=0.0.0.0:9567",
			wantFuseCommand:   "/bin/mount.juicefs redis://127.0.0.1:6379 /test -o metrics=0.0.0.0:9567",
			wantStatCmd:       "stat -c %i /test",
		},
		{
			name: "test-enterprise",
			fields: fields{
				name:      "test",
				namespace: "fluid",
				Log:       fake.NullLogger(),
			},
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-enterprise",
					Edition:          "enterprise",
					Source:           "test-enterprise",
					Fuse: Fuse{
						SubPath:         "/",
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
						MountPath:       "/test",
						CacheDir:        "/cache",
						HostMountPath:   "/test",
					},
				},
			},
			wantErr:           false,
			wantWorkerCommand: "/sbin/mount.juicefs test-enterprise /test -o foreground,cache-group=test-enterprise",
			wantFuseCommand:   "/sbin/mount.juicefs test-enterprise /test -o foreground,cache-group=test-enterprise,no-sharing",
			wantStatCmd:       "stat -c %i /test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}
			if err := j.genMount(tt.args.value, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("genMount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.value.Fuse.Command != tt.wantFuseCommand ||
				tt.args.value.Fuse.StatCmd != tt.wantStatCmd ||
				tt.args.value.Worker.Command != tt.wantWorkerCommand {
				t.Errorf("genMount() value = %v", tt.args.value)
			}
		})
	}
}

func TestJuiceFSEngine_genFormatCmd(t *testing.T) {
	type args struct {
		value *JuiceFS
	}
	tests := []struct {
		name          string
		args          args
		wantFormatCmd string
	}{
		{
			name: "test-community",
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-community",
					Edition:          "community",
					Source:           "redis://127.0.0.1:6379",
					Fuse: Fuse{
						SubPath:         "/",
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
						MountPath:       "/test",
						CacheDir:        "/cache",
						HostMountPath:   "/test",
					},
				},
			},
			wantFormatCmd: "/usr/local/bin/juicefs format --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY} --storage=minio --bucket=http://127.0.0.1:9000/minio/test redis://127.0.0.1:6379 test-community",
		},
		{
			name: "test-enterprise",
			args: args{
				value: &JuiceFS{
					FullnameOverride: "test-enterprise",
					Edition:          "enterprise",
					Source:           "test-enterprise",
					Fuse: Fuse{
						SubPath:         "/",
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
						MountPath:       "/test",
						CacheDir:        "/cache",
						HostMountPath:   "/test",
					},
				},
			},
			wantFormatCmd: "/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://127.0.0.1:9000/minio/test test-enterprise",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			j.genFormatCmd(tt.args.value)
			if tt.args.value.Fuse.FormatCmd != tt.wantFormatCmd {
				t.Errorf("genMount() value = %v", tt.args.value)
			}
		})
	}
}
