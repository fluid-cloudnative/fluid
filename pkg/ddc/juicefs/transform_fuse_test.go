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
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

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
			juicefsValue: &JuiceFS{
				Worker: Worker{},
			},
			expect:  "",
			wantErr: false,
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
	q, _ := resource.ParseQuantity("10240000")
	type fields struct {
		runtime     *datav1alpha1.JuiceFSRuntime
		name        string
		namespace   string
		runtimeType string
	}
	type args struct {
		mount                datav1alpha1.Mount
		tiredStoreLevel      *datav1alpha1.Level
		value                *JuiceFS
		sharedOptions        map[string]string
		sharedEncryptOptions []datav1alpha1.EncryptOption
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantValue   *JuiceFS
		wantOptions map[string]string
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
				sharedOptions: map[string]string{"a": "b"},
				sharedEncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "token",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: "test-enterprise",
							Key:  "token",
						},
					},
				}},
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
					Quota:      &q,
				},
				value: &JuiceFS{
					FullnameOverride: "test",
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			wantErr: false,
			wantOptions: map[string]string{
				"a":          "b",
				"cache-dir":  "/dev",
				"cache-size": "9",
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
				sharedOptions: map[string]string{"a": "b"},
				sharedEncryptOptions: []datav1alpha1.EncryptOption{{
					Name: JuiceMetaUrl,
					ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
						Name: "test-community",
						Key:  "access-key",
					}}}, {
					Name: JuiceAccessKey,
					ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
						Name: "test-community",
						Key:  "secret-key",
					}}}, {
					Name: JuiceSecretKey,
					ValueFrom: datav1alpha1.EncryptOptionSource{SecretKeyRef: datav1alpha1.SecretKeySelector{
						Name: "test-community",
						Key:  "metaurl",
					}}},
				},
				mount: datav1alpha1.Mount{
					MountPoint: "juicefs:///test",
					Options:    map[string]string{"a": "c"},
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
			wantOptions: map[string]string{
				"a":         "c",
				"subdir":    "/test",
				"cache-dir": "/dev",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.genValue(tt.args.mount, tt.args.tiredStoreLevel, tt.args.value, tt.args.sharedOptions, tt.args.sharedEncryptOptions)
			if (err != nil) != tt.wantErr {
				t.Errorf("genValue() error = %v, wantErr %v", err, tt.wantErr)
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
		options map[string]string
		runtime *datav1alpha1.JuiceFSRuntime
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantFuseCommand string
		wantFuseStatCmd string
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
					Configs: Configs{
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test-worker",
					},
				},
			},
			wantErr:         false,
			wantFuseCommand: "/bin/mount.juicefs redis://127.0.0.1:6379 /test -o metrics=0.0.0.0:9567",
			wantFuseStatCmd: "stat -c %i /test",
		},
		{
			name: "test-community-options",
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
					Configs: Configs{
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test-worker",
					},
				},
				options: map[string]string{"verbose": ""},
				runtime: &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Worker: datav1alpha1.JuiceFSCompTemplateSpec{
					Options: map[string]string{"metrics": "127.0.0.1:9567"},
				}}},
			},
			wantErr:         false,
			wantFuseCommand: "/bin/mount.juicefs redis://127.0.0.1:6379 /test -o verbose,metrics=0.0.0.0:9567",
			wantFuseStatCmd: "stat -c %i /test",
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
					Configs: Configs{
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test",
					},
				},
			},
			wantErr:         false,
			wantFuseCommand: "/sbin/mount.juicefs test-enterprise /test -o foreground,cache-group=fluid-test-enterprise,no-sharing",
			wantFuseStatCmd: "stat -c %i /test",
		},
		{
			name: "test-enterprise-options",
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
					Configs: Configs{
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
					Worker: Worker{
						MountPath: "/test",
					},
				},
				options: map[string]string{"cache-group": "test", "verbose": ""},
				runtime: &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Worker: datav1alpha1.JuiceFSCompTemplateSpec{
					Options: map[string]string{"no-sharing": ""},
				}}},
			},
			wantErr:         false,
			wantFuseCommand: "/sbin/mount.juicefs test-enterprise /test -o verbose,foreground,cache-group=test,no-sharing",
			wantFuseStatCmd: "stat -c %i /test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.fields.name,
					Namespace: tt.fields.namespace,
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, dataset)
			j := &JuiceFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
				Client:    client,
			}
			if err := j.genFuseMount(tt.args.value, tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("genMount() \nerror = %v\nwantErr = %v", err, tt.wantErr)
			}
			if len(tt.args.value.Fuse.Command) != len(tt.wantFuseCommand) ||
				tt.args.value.Fuse.StatCmd != tt.wantFuseStatCmd {
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
					Configs: Configs{
						Name:            "test-community",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						MetaUrlSecret:   "test",
						Storage:         "minio",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
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
					Configs: Configs{
						Name:            "test-enterprise",
						AccessKeySecret: "test",
						SecretKeySecret: "test",
						Bucket:          "http://127.0.0.1:9000/minio/test",
						TokenSecret:     "test",
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     "/test",
						HostMountPath: "/test",
					},
				},
			},
			wantFormatCmd: "/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://127.0.0.1:9000/minio/test test-enterprise",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{},
				},
			}
			j.genFormatCmd(tt.args.value, j.runtime.Spec.Configs)
			if tt.args.value.Configs.FormatCmd != tt.wantFormatCmd {
				t.Errorf("genMount() value = %v", tt.args.value)
			}
		})
	}
}

func Test_genOption(t *testing.T) {
	type args struct {
		optionMap map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test",
			args: args{
				optionMap: map[string]string{"a": "b", "c": ""},
			},
			want: []string{"a=b", "c"},
		},
		{
			name: "test-empty",
			args: args{
				optionMap: nil,
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := genArgs(tt.args.optionMap)
			if !isSliceEqual(got, tt.want) {
				t.Errorf("genOption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func isSliceEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	diff := make(map[string]int, len(got))
	for _, v := range got {
		diff[v]++
	}
	for _, v := range want {
		if _, ok := diff[v]; !ok {
			return false
		}
		diff[v] -= 1
		if diff[v] == 0 {
			delete(diff, v)
		}
	}
	return len(diff) == 0
}

func isMapEqual(got, want map[string]string) bool {
	if len(got) != len(want) {
		return false
	}

	for k, v := range got {
		if want[k] != v {
			return false
		}
	}
	return true
}

func TestJuiceFSEngine_getQuota(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "test-1Gi",
			args: args{
				v: "1Gi",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "test-10Gi",
			args: args{
				v: "10Gi",
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "test-1Mi",
			args: args{
				v: "1Mi",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			got, err := j.getQuota(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("getQuota() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getQuota() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseImageTag(t *testing.T) {
	type args struct {
		imageTag string
	}
	tests := []struct {
		name    string
		args    args
		want    *ClientVersion
		want1   *ClientVersion
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				imageTag: "v1.0.4-4.9.0",
			},
			want: &ClientVersion{
				Major: 1,
				Minor: 0,
				Patch: 4,
				Tag:   "",
			},
			want1: &ClientVersion{
				Major: 4,
				Minor: 9,
				Patch: 0,
				Tag:   "",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				imageTag: "nightly",
			},
			want: &ClientVersion{
				Major: 0,
				Minor: 0,
				Patch: 0,
				Tag:   "nightly",
			},
			want1: &ClientVersion{
				Major: 0,
				Minor: 0,
				Patch: 0,
				Tag:   "nightly",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseImageTag(tt.args.imageTag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImageTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseImageTag() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ParseImageTag() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestClientVersion_LessThan(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
		Tag   string
	}
	type args struct {
		other *ClientVersion
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "less",
			fields: fields{
				Major: 1,
				Minor: 0,
				Patch: 0,
				Tag:   "",
			},
			args: args{
				other: &ClientVersion{
					Major: 1,
					Minor: 0,
					Patch: 1,
					Tag:   "",
				},
			},
			want: true,
		},
		{
			name: "more",
			fields: fields{
				Major: 1,
				Minor: 0,
				Patch: 0,
				Tag:   "",
			},
			args: args{
				other: &ClientVersion{
					Major: 0,
					Minor: 1,
					Patch: 0,
					Tag:   "",
				},
			},
			want: false,
		},
		{
			name: "nightly",
			fields: fields{
				Tag: "nightly",
			},
			args: args{
				other: &ClientVersion{
					Major: 1,
					Minor: 0,
					Patch: 0,
					Tag:   "",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &ClientVersion{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
				Tag:   tt.fields.Tag,
			}
			if got := v.LessThan(tt.args.other); got != tt.want {
				t.Errorf("LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJuiceFSEngine_genQuotaCmd(t *testing.T) {
	type args struct {
		value *JuiceFS
		mount datav1alpha1.Mount
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantQuotaCmd string
	}{
		{
			name: "test-ce",
			args: args{
				value: &JuiceFS{
					Edition: CommunityEdition,
					Configs: Configs{},
					Source:  "redis://127.0.0.1:6379",
					Fuse: Fuse{
						ImageTag: "v1.1.4-4.9.2",
						SubPath:  "/demo",
					},
				},
				mount: datav1alpha1.Mount{
					Options: map[string]string{
						"quota": "1Gi",
					},
				},
			},
			wantErr:      false,
			wantQuotaCmd: "/usr/local/bin/juicefs quota set redis://127.0.0.1:6379 --path /demo --capacity 1",
		},
		{
			name: "test-ee",
			args: args{
				value: &JuiceFS{
					Edition: EnterpriseEdition,
					Configs: Configs{},
					Source:  "test",
					Fuse: Fuse{
						ImageTag: "v1.1.4-4.9.2",
						SubPath:  "/demo",
					},
				},
				mount: datav1alpha1.Mount{
					Options: map[string]string{
						"quota": "1Gi",
					},
				},
			},
			wantErr:      false,
			wantQuotaCmd: "/usr/bin/juicefs quota set test --path /demo --capacity 1",
		},
		{
			name: "test-ce-err",
			args: args{
				value: &JuiceFS{
					Edition: CommunityEdition,
					Configs: Configs{},
					Source:  "test",
					Fuse: Fuse{
						ImageTag: "v1.0.4-4.9.1",
						SubPath:  "/demo",
					},
				},
				mount: datav1alpha1.Mount{
					Options: map[string]string{
						"quota": "1Gi",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test-ee-err",
			args: args{
				value: &JuiceFS{
					Edition: EnterpriseEdition,
					Configs: Configs{},
					Source:  "test",
					Fuse: Fuse{
						ImageTag: "v1.1.4-4.9.1",
						SubPath:  "/demo",
					},
				},
				mount: datav1alpha1.Mount{
					Options: map[string]string{
						"quota": "1Gi",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "test-no-subpath",
			args: args{
				value: &JuiceFS{
					Edition: CommunityEdition,
					Configs: Configs{},
					Source:  "test",
					Fuse: Fuse{
						ImageTag: "v1.1.4-4.9.2",
					},
				},
				mount: datav1alpha1.Mount{
					Options: map[string]string{
						"quota": "1Gi",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			if err := j.genQuotaCmd(tt.args.value, tt.args.mount); (err != nil) != tt.wantErr {
				t.Errorf("genQuotaCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantQuotaCmd != tt.args.value.Configs.QuotaCmd {
				t.Errorf("genQuotaCmd() got cmd = %v, want %v", tt.args.value.Configs.QuotaCmd, tt.wantQuotaCmd)
			}
		})
	}
}

func TestJuiceFSEngine_genMountOptions(t *testing.T) {
	result := resource.MustParse("20Gi")
	type args struct {
		mount           datav1alpha1.Mount
		tiredStoreLevel *datav1alpha1.Level
	}
	tests := []struct {
		name        string
		args        args
		wantOptions map[string]string
		wantErr     bool
	}{
		{
			name: "test1",
			args: args{
				mount: datav1alpha1.Mount{
					MountPoint: "juicefs:///test",
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: common.SSD,
					VolumeType: common.VolumeTypeHostPath,
					Path:       "/abc",
					Quota:      &result,
				},
			},
			wantOptions: map[string]string{
				"subdir":     "/test",
				"cache-dir":  "/abc",
				"cache-size": "20480",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			gotOptions, err := j.genMountOptions(tt.args.mount, tt.args.tiredStoreLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("genMountOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !isMapEqual(gotOptions, tt.wantOptions) {
				t.Errorf("genMountOptions() gotOptions = %v, want %v", gotOptions, tt.wantOptions)
			}
		})
	}
}
