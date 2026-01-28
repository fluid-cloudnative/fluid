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
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var dummy = func(client client.Client) (ports []int, err error) {
	return []int{14000, 14001}, nil
}

var _ = Describe("JuiceFSEngine Transform", func() {
	Describe("transform", func() {
		It("should transform fuse configuration correctly", func() {
			juicefsSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Data: map[string][]byte{
					"metaurl": []byte("test"),
				},
			}
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, (*juicefsSecret).DeepCopy())

			fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "juicefs")
			Expect(err).NotTo(HaveOccurred())

			engine := JuiceFSEngine{
				name:      "test",
				namespace: "fluid",
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{},
					},
				},
				runtimeInfo: runtimeInfo,
			}
			ctrl.SetLogger(zap.New(func(o *zap.Options) {
				o.Development = true
			}))

			runtime := &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{},
					Worker: datav1alpha1.JuiceFSCompTemplateSpec{
						Replicas:     2,
						Resources:    corev1.ResourceRequirements{},
						Options:      nil,
						Env:          nil,
						Enabled:      false,
						NodeSelector: nil,
					},
				},
			}
			dataset := &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "juicefs:///mnt/test",
						Name:       "test",
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "metaurl",
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: "test",
									Key:  "metaurl",
								},
							},
						}},
					}},
				},
			}
			value := &JuiceFS{}

			err = engine.transformFuse(runtime, dataset, value)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("transformTolerations", func() {
		It("should transform tolerations correctly", func() {
			j := &JuiceFSEngine{
				name:      "",
				namespace: "",
			}
			dataset := &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "a",
						Operator: corev1.TolerationOpEqual,
						Value:    "b",
					}},
				},
			}
			value := &JuiceFS{}

			j.transformTolerations(dataset, value)
			Expect(len(value.Tolerations)).To(Equal(len(dataset.Spec.Tolerations)))
		})
	})

	Describe("transformPodMetadata", func() {
		var engine *JuiceFSEngine

		BeforeEach(func() {
			engine = &JuiceFSEngine{Log: fake.NullLogger()}
		})

		type testCase struct {
			Name      string
			Runtime   *datav1alpha1.JuiceFSRuntime
			Value     *JuiceFS
			wantValue *JuiceFS
		}

		DescribeTable("should transform pod metadata correctly",
			func(tc testCase) {
				err := engine.transformPodMetadata(tc.Runtime, tc.Value)
				Expect(err).NotTo(HaveOccurred())
				Expect(tc.Value).To(Equal(tc.wantValue))
			},
			Entry("set common labels and annotations", testCase{
				Name: "set_common_labels_and_annotations",
				Runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
					},
				},
				Value: &JuiceFS{},
				wantValue: &JuiceFS{
					Worker: Worker{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
					Fuse: Fuse{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
				},
			}),
			Entry("set master and workers labels and annotations", testCase{
				Name: "set_master_and_workers_labels_and_annotations",
				Runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							PodMetadata: datav1alpha1.PodMetadata{
								Labels:      map[string]string{"common-key": "worker-value"},
								Annotations: map[string]string{"common-annotation": "worker-val"},
							},
						},
					},
				},
				Value: &JuiceFS{},
				wantValue: &JuiceFS{
					Worker: Worker{
						Labels:      map[string]string{"common-key": "worker-value"},
						Annotations: map[string]string{"common-annotation": "worker-val"},
					},
					Fuse: Fuse{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
				},
			}),
		)
	})

	Describe("allocatePorts", func() {
		BeforeEach(func() {
			pr := net.ParsePortRangeOrDie("14000-15999")
			err := portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummy)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should allocate ports for community edition", func() {
			j := &JuiceFSEngine{}
			value := &JuiceFS{
				Edition: CommunityEdition,
			}
			runtime := &datav1alpha1.JuiceFSRuntime{}

			err := j.allocatePorts(runtime, value)
			Expect(err).NotTo(HaveOccurred())

			Expect(value.Worker.MetricsPort).NotTo(BeNil())
			Expect(*value.Worker.MetricsPort).To(BeNumerically(">=", 14000))
			Expect(*value.Worker.MetricsPort).To(BeNumerically("<=", 15999))

			Expect(value.Fuse.MetricsPort).NotTo(BeNil())
			Expect(*value.Fuse.MetricsPort).To(BeNumerically(">=", 14000))
			Expect(*value.Fuse.MetricsPort).To(BeNumerically("<=", 15999))
		})
	})

	Describe("genWorkerMount", func() {
		type testFields struct {
			name      string
			namespace string
			Log       logr.Logger
		}

		type testArgs struct {
			value   *JuiceFS
			runtime *datav1alpha1.JuiceFSRuntime
		}

		type testCase struct {
			name              string
			fields            testFields
			args              testArgs
			wantWorkerCommand string
			wantWorkerStatCmd string
		}

		DescribeTable("should generate worker mount command correctly",
			func(tc testCase) {
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tc.fields.name,
						Namespace: tc.fields.namespace,
					},
				}
				fakeClient := fake.NewFakeClientWithScheme(testScheme, dataset)
				j := &JuiceFSEngine{
					name:      tc.fields.name,
					namespace: tc.fields.namespace,
					Log:       tc.fields.Log,
					Client:    fakeClient,
				}
				j.genWorkerMount(tc.args.value, tc.args.runtime.Spec.Worker.Options)
				Expect(len(tc.args.value.Worker.Command)).To(Equal(len(tc.wantWorkerCommand)))
				Expect(tc.args.value.Worker.StatCmd).To(Equal(tc.wantWorkerStatCmd))
			},
			Entry("test-community", testCase{
				name: "test-community",
				fields: testFields{
					name:      "test",
					namespace: "fluid",
					Log:       fake.NullLogger(),
				},
				args: testArgs{
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
						Worker: Worker{
							MountPath: "/test-worker",
						},
					},
					runtime: &datav1alpha1.JuiceFSRuntime{},
				},
				wantWorkerCommand: "exec /bin/mount.juicefs redis://127.0.0.1:6379 /test-worker -o metrics=0.0.0.0:9567",
				wantWorkerStatCmd: "stat -c %i /test-worker",
			}),
			Entry("test-community-options", testCase{
				name: "test-community-options",
				fields: testFields{
					name:      "test",
					namespace: "fluid",
					Log:       fake.NullLogger(),
				},
				args: testArgs{
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
					runtime: &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Worker: datav1alpha1.JuiceFSCompTemplateSpec{
						Options: map[string]string{"metrics": "127.0.0.1:9567"},
					}}},
				},
				wantWorkerCommand: "exec /bin/mount.juicefs redis://127.0.0.1:6379 /test-worker -o metrics=127.0.0.1:9567",
				wantWorkerStatCmd: "stat -c %i /test-worker",
			}),
			Entry("test-enterprise", testCase{
				name: "test-enterprise",
				fields: testFields{
					name:      "test",
					namespace: "fluid",
					Log:       fake.NullLogger(),
				},
				args: testArgs{
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
					runtime: &datav1alpha1.JuiceFSRuntime{},
				},
				wantWorkerCommand: "exec /sbin/mount.juicefs test-enterprise /test -o foreground,no-update,cache-group=fluid-test-enterprise",
				wantWorkerStatCmd: "stat -c %i /test",
			}),
			Entry("test-enterprise-options", testCase{
				name: "test-enterprise-options",
				fields: testFields{
					name:      "test",
					namespace: "fluid",
					Log:       fake.NullLogger(),
				},
				args: testArgs{
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
					runtime: &datav1alpha1.JuiceFSRuntime{Spec: datav1alpha1.JuiceFSRuntimeSpec{Worker: datav1alpha1.JuiceFSCompTemplateSpec{
						Options: map[string]string{"verbose": "", "cache-group": "test"},
					}}},
				},
				wantWorkerCommand: "exec /sbin/mount.juicefs test-enterprise /test -o verbose,foreground,no-update,cache-group=test",
				wantWorkerStatCmd: "stat -c %i /test",
			}),
		)
	})

	Describe("genEdition", func() {
		type testArgs struct {
			mount                datav1alpha1.Mount
			value                *JuiceFS
			SharedEncryptOptions []datav1alpha1.EncryptOption
		}

		type testCase struct {
			name        string
			args        testArgs
			wantEdition string
		}

		DescribeTable("should generate correct edition",
			func(tc testCase) {
				j := &JuiceFSEngine{}
				j.genEdition(tc.args.mount, tc.args.value, tc.args.SharedEncryptOptions)
				Expect(tc.args.value.Edition).To(Equal(tc.wantEdition))
			},
			Entry("test-community-1", testCase{
				name: "test-community-1",
				args: testArgs{
					mount: datav1alpha1.Mount{
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name:      JuiceMetaUrl,
							ValueFrom: datav1alpha1.EncryptOptionSource{},
						}},
					},
					value:                &JuiceFS{},
					SharedEncryptOptions: nil,
				},
				wantEdition: "community",
			}),
			Entry("test-community-2", testCase{
				name: "test-community-2",
				args: testArgs{
					mount: datav1alpha1.Mount{},
					value: &JuiceFS{},
					SharedEncryptOptions: []datav1alpha1.EncryptOption{{
						Name:      JuiceMetaUrl,
						ValueFrom: datav1alpha1.EncryptOptionSource{},
					}},
				},
				wantEdition: "community",
			}),
			Entry("test-enterprise-1", testCase{
				name: "test-enterprise-1",
				args: testArgs{
					mount: datav1alpha1.Mount{
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name:      JuiceToken,
							ValueFrom: datav1alpha1.EncryptOptionSource{},
						}},
					},
					value:                &JuiceFS{},
					SharedEncryptOptions: nil,
				},
				wantEdition: "enterprise",
			}),
			Entry("test-enterprise-2", testCase{
				name: "test-enterprise-2",
				args: testArgs{
					mount: datav1alpha1.Mount{},
					value: &JuiceFS{},
					SharedEncryptOptions: []datav1alpha1.EncryptOption{{
						Name:      JuiceToken,
						ValueFrom: datav1alpha1.EncryptOptionSource{},
					}},
				},
				wantEdition: "enterprise",
			}),
		)
	})
})
