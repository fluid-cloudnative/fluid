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
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

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
		Log:       log.NullLogger{},
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
