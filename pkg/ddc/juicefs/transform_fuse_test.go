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

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	mockNamespace        = "fluid"
	mockSecretTest1      = "test1"
	mockSecretTest2      = "test2"
	mockSecretTest3      = "test3"
	mockSecretCommunity  = "test-community"
	mockSecretEnterprise = "test-enterprise"
	mockSecretMirror     = "test-mirror-buckets"
	mockMountPoint       = "juicefs:///mnt/test"
	mockMountPointRoot   = "juicefs:///"
	mockMountPointTest   = "juicefs:///test"
	mockMetaurlKey       = "metaurl"
	mockAccessKey        = "access-key"
	mockSecretKey        = "secret-key"
	mockTokenKey         = "token"
	mockStorageValue     = "test1"
	mockBucketValue      = "test1"
	mockTestPath         = "/test"
	mockTestWorkerPath   = "/test-worker"
	mockDataPath         = "/data"
	mockDevPath          = "/dev"
	mockMountPathTest    = "/test"
	mockCachePath        = "/abc"
	mockRedisSource      = "redis://127.0.0.1:6379"
	mockBucketURL        = "http://127.0.0.1:9000/minio/test"
	mockBucket2URL       = "http://127.0.0.1:9001/minio/test"
	mockStorageMinio     = "minio"
	mockEncodedTest      = "test"
	mockEncodedTest2     = "test2"
)

var _ = Describe("TransformFuse", func() {
	var (
		fakeClient  client.Client
		engine      JuiceFSEngine
		runtimeInfo base.RuntimeInfoInterface
		err         error
	)

	BeforeEach(func() {
		juicefsSecret1 := createTestSecret(mockSecretTest1, mockNamespace, map[string][]byte{
			mockMetaurlKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockAccessKey:  []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockSecretKey:  []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
		})
		juicefsSecret2 := createTestSecret(mockSecretTest2, mockNamespace, map[string][]byte{
			mockAccessKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockSecretKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
		})
		juicefsSecret3 := createTestSecret(mockSecretTest3, mockNamespace, map[string][]byte{
			mockAccessKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockSecretKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			AccessKey2:    []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest2))),
			SecretKey2:    []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest2))),
		})
		testObjs := []runtime.Object{}
		testObjs = append(testObjs,
			juicefsSecret1.DeepCopy(),
			juicefsSecret2.DeepCopy(),
			juicefsSecret3.DeepCopy())

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
		runtimeInfo, err = base.BuildRuntimeInfo(mockEncodedTest, mockNamespace, common.JuiceFSRuntime)
		Expect(err).NotTo(HaveOccurred())

		engine = JuiceFSEngine{
			name:        mockEncodedTest,
			namespace:   mockNamespace,
			Client:      fakeClient,
			Log:         fake.NullLogger(),
			runtime:     &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Fuse: datav1alpha1.JuiceFSFuseSpec{}}},
			runtimeInfo: runtimeInfo,
		}
	})

	DescribeTable("transformFuse",
		func(runtime *datav1alpha1.JuiceFSRuntime, dataset *datav1alpha1.Dataset, juicefsValue *JuiceFS, wantErr bool) {
			err := engine.transformFuse(runtime, dataset, juicefsValue)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("secret configured correctly",
			&datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
				}},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockSecretTest1,
						Options: map[string]string{
							"storage": mockStorageValue,
							"bucket":  mockBucketValue,
						},
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockAccessKey, mockSecretTest1, mockAccessKey),
							createEncryptOption(mockSecretKey, mockSecretTest1, mockSecretKey),
							createEncryptOption(mockMetaurlKey, mockSecretTest1, mockMetaurlKey),
						},
					}},
				}},
			&JuiceFS{Worker: Worker{}},
			false,
		),
		Entry("secret without metaurl",
			&datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{Fuse: datav1alpha1.JuiceFSFuseSpec{}}},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockSecretTest2,
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockMetaurlKey, mockSecretTest1, mockMetaurlKey),
						},
					}},
				}}, &JuiceFS{}, true,
		),
		Entry("secret with access key 2",
			&datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
				}},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockSecretTest3,
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockMetaurlKey, mockSecretTest1, mockMetaurlKey),
						},
					}}}},
			&JuiceFS{},
			true,
		),
		Entry("with debug options",
			&datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: "MEM",
							Path:       mockDataPath,
							Low:        "0.7",
						}},
					},
				}},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockSecretTest2,
						Options:    map[string]string{"debug": ""},
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockMetaurlKey, mockSecretTest1, mockMetaurlKey),
						},
					}}}},
			&JuiceFS{},
			false,
		),
		Entry("no mount defined",
			&datav1alpha1.JuiceFSRuntime{},
			&datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-no-mount",
					Namespace: mockNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
			&JuiceFS{},
			true,
		),
		Entry("non existent secret",
			&datav1alpha1.JuiceFSRuntime{},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockSecretTest2,
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockMetaurlKey, "not-exist", mockMetaurlKey),
						},
					}},
				}}, &JuiceFS{}, true,
		),
		Entry("no metaurl in secret",
			&datav1alpha1.JuiceFSRuntime{},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockEncodedTest,
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockMetaurlKey, "no-metaurl", mockMetaurlKey),
						},
					}},
				}}, &JuiceFS{}, true,
		),
		Entry("tiered store with quota",
			&datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{
						Options: map[string]string{"verbose": ""},
					},
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{{
							MediumType: "SSD",
							Path:       mockDataPath,
							Low:        "0.7",
							Quota:      resource.NewQuantity(10, resource.BinarySI),
						}},
					},
				}},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockEncodedTest,
						Options:    map[string]string{"debug": ""},
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockMetaurlKey, mockSecretTest1, mockMetaurlKey),
						},
					}}}},
			&JuiceFS{},
			false,
		),
		Entry("secret3 with all keys",
			&datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
				}},
			&datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: mockMountPoint,
						Name:       mockSecretTest1,
						Options: map[string]string{
							"storage": mockStorageValue,
							"bucket":  mockBucketValue,
						},
						EncryptOptions: []datav1alpha1.EncryptOption{
							createEncryptOption(mockAccessKey, mockSecretTest3, mockAccessKey),
							createEncryptOption(mockSecretKey, mockSecretTest3, mockSecretKey),
							createEncryptOption(mockMetaurlKey, mockSecretTest3, mockMetaurlKey),
							createEncryptOption(SecretKey2, mockSecretTest3, SecretKey2),
							createEncryptOption(AccessKey2, mockSecretTest3, AccessKey2),
						},
					}},
				}},
			&JuiceFS{Worker: Worker{}},
			false,
		),
	)
})

var _ = Describe("GenValue", func() {
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

	var (
		fakeClient client.Client
		engine     JuiceFSEngine
		q          resource.Quantity
	)

	BeforeEach(func() {
		juicefsSecret1 := createTestSecret(mockSecretCommunity, mockNamespace, map[string][]byte{
			mockMetaurlKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockAccessKey:  []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockSecretKey:  []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
		})
		juicefsSecret2 := createTestSecret(mockSecretEnterprise, mockNamespace, map[string][]byte{
			mockTokenKey:  []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockAccessKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockSecretKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
		})
		juicefsSecret3 := createTestSecret(mockSecretMirror, mockNamespace, map[string][]byte{
			mockTokenKey:  []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockAccessKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			mockSecretKey: []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest))),
			AccessKey2:    []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest2))),
			SecretKey2:    []byte(base64.StdEncoding.EncodeToString([]byte(mockEncodedTest2))),
		})

		testObjs := []runtime.Object{}
		testObjs = append(testObjs,
			juicefsSecret1.DeepCopy(),
			juicefsSecret2.DeepCopy(),
			juicefsSecret3.DeepCopy())

		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
		q, _ = resource.ParseQuantity("10240000")
	})

	DescribeTable("genValue",
		func(f fields, a args, wantErr bool, wantOptions map[string]string) {
			engine = JuiceFSEngine{
				name:      f.name,
				namespace: f.namespace,
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{},
					},
				},
			}
			opt, err := engine.genValue(a.mount, a.tiredStoreLevel, a.value, a.sharedOptions, a.sharedEncryptOptions)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(opt).To(Equal(wantOptions))
		},
		Entry("enterprise with token",
			fields{
				runtime:     nil,
				name:        mockEncodedTest,
				namespace:   mockNamespace,
				runtimeType: common.JuiceFSRuntime,
			},
			args{
				sharedOptions: map[string]string{"a": "b"},
				sharedEncryptOptions: []datav1alpha1.EncryptOption{
					createEncryptOption(mockTokenKey, mockSecretEnterprise, mockTokenKey),
				},
				mount: datav1alpha1.Mount{
					MountPoint: mockMountPointRoot,
					Options:    nil,
					Name:       mockEncodedTest,
					EncryptOptions: []datav1alpha1.EncryptOption{
						createEncryptOption(mockTokenKey, mockSecretEnterprise, mockTokenKey),
					},
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: "SSD",
					Path:       mockDevPath,
					Quota:      &q,
				},
				value: &JuiceFS{
					FullnameOverride: mockEncodedTest,
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			false,
			map[string]string{"a": "b"},
		),
		Entry("community with metaurl",
			fields{
				name:        mockSecretCommunity,
				namespace:   mockNamespace,
				runtimeType: common.JuiceFSRuntime,
			},
			args{
				sharedOptions: map[string]string{"a": "b"},
				sharedEncryptOptions: []datav1alpha1.EncryptOption{
					createEncryptOption(JuiceMetaUrl, mockSecretCommunity, mockAccessKey),
					createEncryptOption(JuiceAccessKey, mockSecretCommunity, mockSecretKey),
					createEncryptOption(JuiceSecretKey, mockSecretCommunity, mockMetaurlKey),
				},
				mount: datav1alpha1.Mount{
					MountPoint: mockMountPointTest,
					Options:    map[string]string{"a": "c"},
					Name:       mockSecretCommunity,
					EncryptOptions: []datav1alpha1.EncryptOption{
						createEncryptOption(JuiceMetaUrl, mockSecretCommunity, mockMetaurlKey),
						createEncryptOption(JuiceAccessKey, mockSecretCommunity, mockAccessKey),
						createEncryptOption(JuiceSecretKey, mockSecretCommunity, mockSecretKey),
					},
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: "SSD",
					Path:       mockDevPath,
				},
				value: &JuiceFS{
					FullnameOverride: mockSecretCommunity,
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			false,
			map[string]string{"a": "c"},
		),
		Entry("mirror buckets",
			fields{
				name:        mockSecretMirror,
				namespace:   mockNamespace,
				runtimeType: common.JuiceFSRuntime,
			},
			args{
				sharedOptions: map[string]string{"a": "b"},
				sharedEncryptOptions: []datav1alpha1.EncryptOption{
					createEncryptOption(JuiceMetaUrl, mockSecretMirror, mockAccessKey),
					createEncryptOption(JuiceAccessKey, mockSecretMirror, mockSecretKey),
					createEncryptOption(JuiceSecretKey, mockSecretMirror, mockMetaurlKey),
				},
				mount: datav1alpha1.Mount{
					MountPoint: mockMountPointTest,
					Options:    map[string]string{"a": "c"},
					Name:       mockSecretMirror,
					EncryptOptions: []datav1alpha1.EncryptOption{
						createEncryptOption(JuiceMetaUrl, mockSecretMirror, mockMetaurlKey),
						createEncryptOption(JuiceAccessKey, mockSecretMirror, mockAccessKey),
						createEncryptOption(JuiceSecretKey, mockSecretMirror, mockSecretKey),
						createEncryptOption(AccessKey2, mockSecretMirror, AccessKey2),
						createEncryptOption(SecretKey2, mockSecretMirror, SecretKey2),
					},
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: "SSD",
					Path:       mockDevPath,
				},
				value: &JuiceFS{
					FullnameOverride: mockSecretMirror,
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			false,
			map[string]string{"a": "c"},
		),
		Entry("shared mirror buckets",
			fields{
				name:        mockSecretMirror,
				namespace:   mockNamespace,
				runtimeType: common.JuiceFSRuntime,
			},
			args{
				sharedOptions: map[string]string{"a": "b"},
				sharedEncryptOptions: []datav1alpha1.EncryptOption{
					createEncryptOption(JuiceMetaUrl, mockSecretMirror, mockAccessKey),
					createEncryptOption(JuiceAccessKey, mockSecretMirror, mockSecretKey),
					createEncryptOption(JuiceSecretKey, mockSecretMirror, mockMetaurlKey),
					createEncryptOption(AccessKey2, mockSecretMirror, AccessKey2),
					createEncryptOption(SecretKey2, mockSecretMirror, SecretKey2),
				},
				mount: datav1alpha1.Mount{
					MountPoint: mockMountPointTest,
					Options:    map[string]string{"a": "c"},
					Name:       mockSecretMirror,
					EncryptOptions: []datav1alpha1.EncryptOption{
						createEncryptOption(JuiceMetaUrl, mockSecretMirror, mockMetaurlKey),
						createEncryptOption(JuiceAccessKey, mockSecretMirror, mockAccessKey),
						createEncryptOption(JuiceSecretKey, mockSecretMirror, mockSecretKey),
					},
				},
				tiredStoreLevel: &datav1alpha1.Level{
					MediumType: "SSD",
					Path:       mockDevPath,
				},
				value: &JuiceFS{
					FullnameOverride: mockSecretMirror,
					Fuse:             Fuse{},
					Worker:           Worker{},
				},
			},
			false,
			map[string]string{"a": "c"},
		),
	)
})

var _ = Describe("GenFuseMount", func() {
	type testCase struct {
		name            string
		engineName      string
		engineNamespace string
		log             logr.Logger
		value           *JuiceFS
		options         map[string]string
		wantErr         bool
		wantFuseCommand string
		wantFuseStatCmd string
	}

	DescribeTable("genFuseMount",
		func(tc testCase) {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tc.engineName,
					Namespace: tc.engineNamespace,
				},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, dataset)
			j := &JuiceFSEngine{
				name:      tc.engineName,
				namespace: tc.engineNamespace,
				Log:       tc.log,
				Client:    fakeClient,
			}
			err := j.genFuseMount(tc.value, tc.options)
			if tc.wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(len(tc.value.Fuse.Command)).To(Equal(len(tc.wantFuseCommand)))
			Expect(tc.value.Fuse.StatCmd).To(Equal(tc.wantFuseStatCmd))
		},
		Entry("community edition",
			testCase{
				name:            "community",
				engineName:      mockEncodedTest,
				engineNamespace: mockNamespace,
				log:             fake.NullLogger(),
				value: &JuiceFS{
					FullnameOverride: mockSecretCommunity,
					Edition:          CommunityEdition,
					Source:           mockRedisSource,
					Configs: Configs{
						Name:            mockSecretCommunity,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						MetaUrlSecret:   mockEncodedTest,
						Storage:         mockStorageMinio,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
					Worker: Worker{
						MountPath: mockTestWorkerPath,
					},
				},
				wantErr:         false,
				wantFuseCommand: "exec /bin/mount.juicefs redis://127.0.0.1:6379 /test -o metrics=0.0.0.0:9567",
				wantFuseStatCmd: "stat -c %i /test",
			},
		),
		Entry("community edition with options",
			testCase{
				name:            "community with options",
				engineName:      mockEncodedTest,
				engineNamespace: mockNamespace,
				log:             fake.NullLogger(),
				value: &JuiceFS{
					FullnameOverride: mockSecretCommunity,
					Edition:          CommunityEdition,
					Source:           mockRedisSource,
					Configs: Configs{
						Name:            mockSecretCommunity,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						MetaUrlSecret:   mockEncodedTest,
						Storage:         mockStorageMinio,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
					Worker: Worker{
						MountPath: mockTestWorkerPath,
					},
				},
				options:         map[string]string{"verbose": ""},
				wantErr:         false,
				wantFuseCommand: "exec /bin/mount.juicefs redis://127.0.0.1:6379 /test -o verbose,metrics=0.0.0.0:9567",
				wantFuseStatCmd: "stat -c %i /test",
			},
		),
		Entry("enterprise edition",
			testCase{
				name:            "enterprise",
				engineName:      mockEncodedTest,
				engineNamespace: mockNamespace,
				log:             fake.NullLogger(),
				value: &JuiceFS{
					FullnameOverride: mockSecretEnterprise,
					Edition:          EnterpriseEdition,
					Source:           mockSecretEnterprise,
					Configs: Configs{
						Name:            mockSecretEnterprise,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						TokenSecret:     mockEncodedTest,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
					Worker: Worker{
						MountPath: mockMountPathTest,
					},
				},
				wantErr:         false,
				wantFuseCommand: "exec /sbin/mount.juicefs test-enterprise /test -o foreground,no-update,cache-group=fluid-test-enterprise,no-sharing",
				wantFuseStatCmd: "stat -c %i /test",
			},
		),
		Entry("enterprise edition with options",
			testCase{
				name:            "enterprise with options",
				engineName:      mockEncodedTest,
				engineNamespace: mockNamespace,
				log:             fake.NullLogger(),
				value: &JuiceFS{
					FullnameOverride: mockSecretEnterprise,
					Edition:          EnterpriseEdition,
					Source:           mockSecretEnterprise,
					Configs: Configs{
						Name:            mockSecretEnterprise,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						TokenSecret:     mockEncodedTest,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
					Worker: Worker{
						MountPath: mockMountPathTest,
					},
				},
				options:         map[string]string{"cache-group": mockEncodedTest, "verbose": ""},
				wantErr:         false,
				wantFuseCommand: "exec /sbin/mount.juicefs test-enterprise /test -o verbose,foreground,no-update,cache-group=test,no-sharing",
				wantFuseStatCmd: "stat -c %i /test",
			},
		),
		Entry("enterprise edition with bucket2",
			testCase{
				name:            "enterprise with bucket2",
				engineName:      mockEncodedTest,
				engineNamespace: mockNamespace,
				log:             fake.NullLogger(),
				value: &JuiceFS{
					FullnameOverride: mockSecretEnterprise,
					Edition:          EnterpriseEdition,
					Source:           mockSecretEnterprise,
					Configs: Configs{
						Name:            mockSecretEnterprise,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						TokenSecret:     mockEncodedTest,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
					Worker: Worker{
						MountPath: mockMountPathTest,
					},
				},
				options:         map[string]string{"cache-group": mockEncodedTest, "verbose": "", JuiceBucket2: "bucket2"},
				wantErr:         false,
				wantFuseCommand: "exec /sbin/mount.juicefs test-enterprise /test -o verbose,foreground,no-update,cache-group=test,no-sharing",
				wantFuseStatCmd: "stat -c %i /test",
			},
		),
	)
})

var _ = Describe("GenFormatCmd", func() {
	type testCase struct {
		name          string
		value         *JuiceFS
		options       map[string]string
		wantFormatCmd string
	}

	DescribeTable("genFormatCmd",
		func(tc testCase) {
			j := &JuiceFSEngine{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{},
				},
			}
			j.genFormatCmd(tc.value, j.runtime.Spec.Configs, tc.options)
			Expect(tc.value.Configs.FormatCmd).To(Equal(tc.wantFormatCmd))
		},
		Entry("community edition",
			testCase{
				name: "community",
				value: &JuiceFS{
					FullnameOverride: mockSecretCommunity,
					Edition:          CommunityEdition,
					Source:           mockRedisSource,
					Configs: Configs{
						Name:            mockSecretCommunity,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						MetaUrlSecret:   mockEncodedTest,
						Storage:         mockStorageMinio,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
				},
				options:       map[string]string{},
				wantFormatCmd: "/usr/local/bin/juicefs format --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY} --storage=minio --bucket=http://127.0.0.1:9000/minio/test redis://127.0.0.1:6379 test-community",
			},
		),
		Entry("enterprise edition",
			testCase{
				name: "enterprise",
				value: &JuiceFS{
					FullnameOverride: mockSecretEnterprise,
					Edition:          EnterpriseEdition,
					Source:           mockSecretEnterprise,
					Configs: Configs{
						Name:            mockSecretEnterprise,
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						TokenSecret:     mockEncodedTest,
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
				},
				options:       map[string]string{},
				wantFormatCmd: "/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://127.0.0.1:9000/minio/test test-enterprise",
			},
		),
		Entry("mirror bucket",
			testCase{
				name: "mirror bucket",
				value: &JuiceFS{
					FullnameOverride: "test-mirror-bucket",
					Edition:          EnterpriseEdition,
					Source:           "test-mirror-bucket",
					Configs: Configs{
						Name:            "test-mirror-bucket",
						AccessKeySecret: mockEncodedTest,
						SecretKeySecret: mockEncodedTest,
						Bucket:          mockBucketURL,
						TokenSecret:     mockEncodedTest,
						EncryptEnvOptions: []EncryptEnvOption{
							{
								Name:    AccessKey2,
								EnvName: "access_key2",
							},
							{
								Name:    SecretKey2,
								EnvName: "secret_key2",
							},
						},
					},
					Fuse: Fuse{
						SubPath:       "/",
						MountPath:     mockMountPathTest,
						HostMountPath: mockMountPathTest,
					},
				},
				options: map[string]string{
					JuiceBucket2: mockBucket2URL,
				},
				wantFormatCmd: "/usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://127.0.0.1:9000/minio/test --bucket2=http://127.0.0.1:9001/minio/test --access-key2=${access_key2} --secret-key2=${secret_key2} test-mirror-bucket",
			},
		),
	)
})

var _ = Describe("GenArgs", func() {
	DescribeTable("genArgs",
		func(optionMap map[string]string, wantContains []string) {
			got := genArgs(optionMap)
			Expect(isSliceEqual(got, wantContains)).To(BeTrue())
		},
		Entry("with values",
			map[string]string{"a": "b", "c": ""},
			[]string{"a=b", "c"},
		),
		Entry("empty map",
			nil,
			[]string{},
		),
	)
})

var _ = Describe("GetQuota", func() {
	DescribeTable("getQuota",
		func(input string, want int64, wantErr bool) {
			j := &JuiceFSEngine{}
			got, err := j.getQuota(input)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(want))
			}
		},
		Entry("1Gi quota", "1Gi", int64(1), false),
		Entry("10Gi quota", "10Gi", int64(10), false),
		Entry("1Mi quota too small", "1Mi", int64(0), true),
	)
})

var _ = Describe("ParseImageTag", func() {
	DescribeTable("parseImageTag",
		func(imageTag string, wantCE *ClientVersion, wantEE *ClientVersion, wantErr bool) {
			gotCE, gotEE, err := ParseImageTag(imageTag)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(reflect.DeepEqual(gotCE, wantCE)).To(BeTrue())
				Expect(reflect.DeepEqual(gotEE, wantEE)).To(BeTrue())
			}
		},
		Entry("version format",
			"v1.0.4-4.9.0",
			&ClientVersion{Major: 1, Minor: 0, Patch: 4, Tag: ""},
			&ClientVersion{Major: 4, Minor: 9, Patch: 0, Tag: ""},
			false,
		),
		Entry("nightly tag",
			"nightly",
			&ClientVersion{Major: 0, Minor: 0, Patch: 0, Tag: "nightly"},
			&ClientVersion{Major: 0, Minor: 0, Patch: 0, Tag: "nightly"},
			false,
		),
	)
})

var _ = Describe("ClientVersionLessThan", func() {
	DescribeTable("lessThan",
		func(v *ClientVersion, other *ClientVersion, want bool) {
			got := v.LessThan(other)
			Expect(got).To(Equal(want))
		},
		Entry("less than",
			&ClientVersion{Major: 1, Minor: 0, Patch: 0, Tag: ""},
			&ClientVersion{Major: 1, Minor: 0, Patch: 1, Tag: ""},
			true,
		),
		Entry("greater than",
			&ClientVersion{Major: 1, Minor: 0, Patch: 0, Tag: ""},
			&ClientVersion{Major: 0, Minor: 1, Patch: 0, Tag: ""},
			false,
		),
		Entry("nightly tag",
			&ClientVersion{Tag: "nightly"},
			&ClientVersion{Major: 1, Minor: 0, Patch: 0, Tag: ""},
			false,
		),
	)
})

var _ = Describe("GenQuotaCmd", func() {
	DescribeTable("genQuotaCmd",
		func(value *JuiceFS, mount datav1alpha1.Mount, wantErr bool, wantQuotaCmd string) {
			j := &JuiceFSEngine{}
			err := j.genQuotaCmd(value, mount)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Configs.QuotaCmd).To(Equal(wantQuotaCmd))
			}
		},
		Entry("community edition",
			&JuiceFS{
				Edition: CommunityEdition,
				Configs: Configs{},
				Source:  mockRedisSource,
				Fuse: Fuse{
					ImageTag: "v1.1.4-4.9.2",
					SubPath:  "/demo",
				},
			},
			datav1alpha1.Mount{
				Options: map[string]string{
					"quota": "1Gi",
				},
			},
			false,
			"/usr/local/bin/juicefs quota set redis://127.0.0.1:6379 --path /demo --capacity 1",
		),
		Entry("enterprise edition",
			&JuiceFS{
				Edition: EnterpriseEdition,
				Configs: Configs{},
				Source:  mockEncodedTest,
				Fuse: Fuse{
					ImageTag: "v1.1.4-4.9.2",
					SubPath:  "/demo",
				},
			},
			datav1alpha1.Mount{
				Options: map[string]string{
					"quota": "1Gi",
				},
			},
			false,
			"/usr/bin/juicefs quota set test --path /demo --capacity 1",
		),
	)
})

var _ = Describe("GenMountOptions", func() {
	DescribeTable("genMountOptions",
		func(mount datav1alpha1.Mount, tiredStoreLevel *datav1alpha1.Level, wantOptions map[string]string, wantErr bool) {
			j := &JuiceFSEngine{}
			gotOptions, err := j.genMountOptions(mount, tiredStoreLevel)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(isMapEqual(gotOptions, wantOptions)).To(BeTrue())
			}
		},
		Entry("with tiered store",
			datav1alpha1.Mount{
				MountPoint: mockMountPointTest,
			},
			&datav1alpha1.Level{
				MediumType: common.SSD,
				VolumeType: common.VolumeTypeHostPath,
				Path:       mockCachePath,
				Quota:      resource.NewQuantity(20*1024*1024*1024, resource.BinarySI),
			},
			map[string]string{
				"subdir":     "/test",
				"cache-dir":  mockCachePath,
				"cache-size": "20480",
			},
			false,
		),
	)
})

func createTestSecret(name string, namespace string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func createEncryptOption(name string, secretName string, secretKey string) datav1alpha1.EncryptOption {
	return datav1alpha1.EncryptOption{
		Name: name,
		ValueFrom: datav1alpha1.EncryptOptionSource{
			SecretKeyRef: datav1alpha1.SecretKeySelector{
				Name: secretName,
				Key:  secretKey,
			},
		},
	}
}

func isSliceEqual(got []string, want []string) bool {
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
		diff[v]--
		if diff[v] == 0 {
			delete(diff, v)
		}
	}
	return len(diff) == 0
}

func isMapEqual(got map[string]string, want map[string]string) bool {
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
