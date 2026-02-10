/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var _ = Describe("JindoCacheEngine Transform", func() {
	Describe("transformTolerations", func() {
		var (
			engine     *JindoCacheEngine
			dataset    *datav1alpha1.Dataset
			jindoValue *Jindo
		)

		BeforeEach(func() {
			engine = &JindoCacheEngine{
				Log: fake.NullLogger(),
			}
			jindoValue = &Jindo{}
		})

		Context("when dataset have tolerations", func() {
			BeforeEach(func() {
				dataset = &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{{
							MountPoint: "local:///mnt/test",
							Name:       "test",
						}},
						Tolerations: []corev1.Toleration{{
							Key:      "jindo",
							Operator: "Equals",
							Value:    "true",
						}},
					},
				}
			})
			Context("and runtime have tolerations", func() {
				It("should correctly combine tolerations from dataset and runtime", func() {
					result := resource.MustParse("20Gi")
					testRuntime := &datav1alpha1.JindoRuntime{
						Spec: datav1alpha1.JindoRuntimeSpec{
							Secret: "secret",
							TieredStore: datav1alpha1.TieredStore{
								Levels: []datav1alpha1.Level{{
									MediumType: common.Memory,
									Quota:      &result,
									High:       "0.8",
									Low:        "0.1",
								}},
							},
							Master: datav1alpha1.JindoCompTemplateSpec{
								Tolerations: []corev1.Toleration{{
									Key:      "master",
									Operator: "Equals",
									Value:    "true",
								}},
							},
							Worker: datav1alpha1.JindoCompTemplateSpec{
								Tolerations: []corev1.Toleration{{
									Key:      "worker",
									Operator: "Equals",
									Value:    "true",
								}},
							},
							Fuse: datav1alpha1.JindoFuseSpec{
								Tolerations: []corev1.Toleration{{
									Key:      "fuse",
									Operator: "Equals",
									Value:    "true",
								}},
							},
						},
					}

					engine.transformTolerations(dataset, testRuntime, jindoValue)

					Expect(len(jindoValue.Master.Tolerations)).To(Equal(2))
					Expect(len(jindoValue.Worker.Tolerations)).To(Equal(2))
					Expect(len(jindoValue.Fuse.Tolerations)).To(Equal(2))
				})
			})

			Context("and runtime does not have tolerations", func() {
				It("should contains tolerations defined in dataset", func() {
					result := resource.MustParse("20Gi")
					testRuntime := &datav1alpha1.JindoRuntime{
						Spec: datav1alpha1.JindoRuntimeSpec{
							Secret: "secret",
							TieredStore: datav1alpha1.TieredStore{
								Levels: []datav1alpha1.Level{{
									MediumType: common.Memory,
									Quota:      &result,
									High:       "0.8",
									Low:        "0.1",
								}},
							},
						},
					}
					engine.transformTolerations(dataset, testRuntime, jindoValue)

					Expect(len(jindoValue.Master.Tolerations)).To(Equal(1))
					Expect(len(jindoValue.Worker.Tolerations)).To(Equal(1))
					Expect(len(jindoValue.Fuse.Tolerations)).To(Equal(1))
				})
			})
		})

		Context("when only runtime have tolerations", func() {
			It("should contains tolerations defined in runtime", func() {
				dataset = &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{{
							MountPoint: "local:///mnt/test",
							Name:       "test",
						}},
					},
				}

				result := resource.MustParse("20Gi")
				testRuntime := &datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						Master: datav1alpha1.JindoCompTemplateSpec{
							Tolerations: []corev1.Toleration{
								{
									Key:      "toleration1",
									Operator: "Equals",
									Value:    "master",
								},
								{
									Key:      "toleration2",
									Operator: "Equals",
									Value:    "master",
								},
							},
						},
					},
				}
				engine.transformTolerations(dataset, testRuntime, jindoValue)

				Expect(len(jindoValue.Master.Tolerations)).To(Equal(2))
				Expect(jindoValue.Master.Tolerations[0].Key).To(Equal("toleration1"))
				Expect(jindoValue.Master.Tolerations[0].Value).To(Equal("master"))
				Expect(len(jindoValue.Worker.Tolerations)).To(Equal(0))
				Expect(len(jindoValue.Fuse.Tolerations)).To(Equal(0))
			})
		})

		It("should handle table-driven tolerations test", func() {
			resources := corev1.ResourceRequirements{}
			resources.Limits = make(corev1.ResourceList)
			resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

			result := resource.MustParse("20Gi")
			tests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				dataset     *datav1alpha1.Dataset
				jindoValue  *Jindo
				expect      int
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						Master: datav1alpha1.JindoCompTemplateSpec{
							Tolerations: []corev1.Toleration{{
								Key:      "master",
								Operator: "Equals",
								Value:    "true",
							}},
						},
						Worker: datav1alpha1.JindoCompTemplateSpec{
							Tolerations: []corev1.Toleration{{
								Key:      "worker",
								Operator: "Equals",
								Value:    "true",
							}},
						},
						Fuse: datav1alpha1.JindoFuseSpec{
							Tolerations: []corev1.Toleration{{
								Key:      "fuse",
								Operator: "Equals",
								Value:    "true",
							}},
						},
					},
				}, &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{{
							MountPoint: "local:///mnt/test",
							Name:       "test",
						}},
						Tolerations: []corev1.Toleration{{
							Key:      "jindo",
							Operator: "Equals",
							Value:    "true",
						}},
					}}, &Jindo{}, 2,
				},
			}
			for _, test := range tests {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				engine.transformTolerations(test.dataset, test.testRuntime, test.jindoValue)
				Expect(len(test.jindoValue.Master.Tolerations)).To(Equal(test.expect))
				Expect(len(test.jindoValue.Worker.Tolerations)).To(Equal(test.expect))
				Expect(len(test.jindoValue.Fuse.Tolerations)).To(Equal(test.expect))
			}
		})
	})

	Describe("parseSmartDataImage", func() {
		DescribeTable("should parse smart data image correctly",
			func(testRuntime *datav1alpha1.JindoRuntime, expect, expectImagePullPolicy, expectDnsServer string) {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				smartdataConfig := engine.getSmartDataConfigs(testRuntime)
				registryVersion := smartdataConfig.image + ":" + smartdataConfig.imageTag
				Expect(registryVersion).To(Equal(expect))
				Expect(smartdataConfig.imagePullPolicy).To(Equal(expectImagePullPolicy))
				Expect(smartdataConfig.dnsServer).To(Equal(expectDnsServer))
			},
			Entry("default image",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
					}},
				"registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:6.2.0",
				"Always",
				"1.1.1.1",
			),
			Entry("custom image",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						JindoVersion: datav1alpha1.VersionSpec{
							Image:           "jindofs/smartdata",
							ImageTag:        "testtag",
							ImagePullPolicy: "IfNotPresent",
						},
					}},
				"jindofs/smartdata:testtag",
				"IfNotPresent",
				"1.1.1.1",
			),
		)
	})

	Describe("transformHostNetWork", func() {
		It("should transform host network correctly", func() {
			result := resource.MustParse("20Gi")
			tests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				jindoValue  *Jindo
				expect      bool
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
					},
				}, &Jindo{}, true},
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						NetworkMode: "HostNetwork",
					},
				}, &Jindo{}, true},
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						NetworkMode: "ContainerNetwork",
					},
				}, &Jindo{}, false},
			}
			for _, test := range tests {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				engine.transformNetworkMode(test.testRuntime, test.jindoValue)
				Expect(test.jindoValue.UseHostNetwork).To(Equal(test.expect))
			}
		})

		It("should handle invalid network mode", func() {
			result := resource.MustParse("20Gi")
			errortests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				jindoValue  *Jindo
				expect      bool
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						NetworkMode: "Non",
					},
				}, &Jindo{}, false},
			}
			for _, test := range errortests {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				engine.transformNetworkMode(test.testRuntime, test.jindoValue)
				Expect(test.jindoValue.UseHostNetwork).To(Equal(test.expect))
			}
		})
	})

	Describe("transformAllocatePorts", func() {
		It("should allocate ports correctly", func() {
			result := resource.MustParse("20Gi")
			tests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				jindoValue  *Jindo
				expect      int
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						NetworkMode: "ContainerNetwork",
					},
				}, &Jindo{}, 8101},
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Secret: "secret",
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &result,
								High:       "0.8",
								Low:        "0.1",
							}},
						},
						NetworkMode: "ContainerNetwork",
						Replicas:    3,
					},
				}, &Jindo{}, 8101},
			}
			for _, test := range tests {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				engine.transformNetworkMode(test.testRuntime, test.jindoValue)
				test.jindoValue.Master.ReplicaCount = 3
				err := engine.allocatePorts(test.testRuntime, test.jindoValue)
				if test.jindoValue.Master.Port.Rpc != test.expect && err != nil {
					Fail("expected port allocation to match")
				}
			}
		})
	})

	Describe("transformMasterResources", func() {
		It("should transform master resources correctly", func() {
			_ = os.Setenv("USE_DEFAULT_MEM_LIMIT", "true")
			quotas := []resource.Quantity{resource.MustParse("200Gi"), resource.MustParse("10Gi")}
			tests := []struct {
				name                string
				nameField           string
				namespace           string
				testRuntime         *datav1alpha1.JindoRuntime
				value               *Jindo
				userQuotas          string
				wantErr             bool
				wantRuntimeResource corev1.ResourceRequirements
				wantValue           common.Resources
			}{
				{
					name:      "runtime_resource_is_null",
					nameField: "testNull",
					namespace: "default",
					testRuntime: &datav1alpha1.JindoRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "testNull",
							Namespace: "default",
						},
						Spec: datav1alpha1.JindoRuntimeSpec{
							TieredStore: datav1alpha1.TieredStore{
								Levels: []datav1alpha1.Level{{
									MediumType: common.Memory,
									Quota:      &quotas[0],
									High:       "0.8",
									Low:        "0.1",
								}},
							},
						},
					},
					value:      &Jindo{},
					userQuotas: "200g",
					wantErr:    false,
					wantRuntimeResource: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("30Gi"),
						},
					},
					wantValue: common.Resources{
						Requests: common.ResourceList{
							corev1.ResourceMemory: "30Gi",
						},
						Limits: common.ResourceList{},
					},
				},
			}
			for _, tt := range tests {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, tt.testRuntime.DeepCopy())
				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, tt.testRuntime)
				_ = corev1.AddToScheme(s)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				e := &JindoCacheEngine{
					name:      tt.nameField,
					namespace: tt.namespace,
					Client:    client,
				}

				err := e.transformMasterResources(tt.testRuntime, tt.value, tt.userQuotas)
				if tt.wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}

				fetchedRuntime, err := e.getRuntime()
				Expect(err).NotTo(HaveOccurred())

				Expect(utils.ResourceRequirementsEqual(tt.wantRuntimeResource, fetchedRuntime.Spec.Master.Resources)).To(BeTrue())
				Expect(reflect.DeepEqual(tt.wantValue, tt.value.Master.Resources)).To(BeTrue())
			}
		})
	})

	Describe("transform", func() {
		DescribeTable("should transform runtime correctly",
			func(name, namespace string, testRuntime *datav1alpha1.JindoRuntime, dataset *datav1alpha1.Dataset, wantErr bool) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, testRuntime.DeepCopy())
				runtimeObjs = append(runtimeObjs, dataset.DeepCopy())
				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
				s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.DatasetList{})
				_ = corev1.AddToScheme(s)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "jindocache")
				Expect(err).NotTo(HaveOccurred())

				e := &JindoCacheEngine{
					runtime:     testRuntime,
					name:        name,
					namespace:   namespace,
					Client:      client,
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfo,
				}
				err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
				Expect(err).NotTo(HaveOccurred())

				_, err = e.transform(testRuntime)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("fuseOnly",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
				false,
			),
			Entry("pvc",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
							},
						},
					},
				},
				false,
			),
			Entry("pvc-subpath",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data/subpath",
							},
						},
					},
				},
				false,
			),
		)
	})

	Describe("transformDeployMode", func() {
		DescribeTable("should transform deploy mode correctly",
			func(testRuntime *datav1alpha1.JindoRuntime, value *Jindo, wantMaster Master) {
				e := &JindoCacheEngine{
					runtime: testRuntime,
				}
				value.Master.ReplicaCount = e.transformReplicasCount(testRuntime)
				value.Master.ServiceCount = e.transformReplicasCount(testRuntime)
				e.transformDeployMode(testRuntime, value)
				Expect(reflect.DeepEqual(wantMaster, value.Master)).To(BeTrue())
			},
			Entry("replicas is 1, enabled",
				&datav1alpha1.JindoRuntime{},
				&Jindo{},
				Master{
					ServiceCount: 1,
					ReplicaCount: 1,
				},
			),
			Entry("replicas is 1, disabled",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Disabled: true,
						},
					},
				},
				&Jindo{},
				Master{
					ServiceCount: 1,
					ReplicaCount: 0,
				},
			),
			Entry("replicas is 3, disabled",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Disabled: true,
							Replicas: 3,
						},
					},
				},
				&Jindo{},
				Master{
					ServiceCount: 3,
					ReplicaCount: 0,
				},
			),
		)
	})

	Describe("transformPodMetadata", func() {
		It("should transform pod metadata correctly", func() {
			engine := &JindoCacheEngine{Log: fake.NullLogger()}

			testCases := []struct {
				Name      string
				Runtime   *datav1alpha1.JindoRuntime
				Value     *Jindo
				wantValue *Jindo
			}{
				{
					Name: "set_common_labels_and_annotations",
					Runtime: &datav1alpha1.JindoRuntime{
						Spec: datav1alpha1.JindoRuntimeSpec{
							PodMetadata: datav1alpha1.PodMetadata{
								Labels:      map[string]string{"common-key": "common-value"},
								Annotations: map[string]string{"common-annotation": "val"},
							},
						},
					},
					Value: &Jindo{},
					wantValue: &Jindo{
						Master: Master{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
						Worker: Worker{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
						Fuse: Fuse{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
					},
				},
				{
					Name: "set_master_and_workers_labels_and_annotations",
					Runtime: &datav1alpha1.JindoRuntime{
						Spec: datav1alpha1.JindoRuntimeSpec{
							PodMetadata: datav1alpha1.PodMetadata{
								Labels:      map[string]string{"common-key": "common-value"},
								Annotations: map[string]string{"common-annotation": "val"},
							},
							Master: datav1alpha1.JindoCompTemplateSpec{
								PodMetadata: datav1alpha1.PodMetadata{
									Labels:      map[string]string{"common-key": "master-value"},
									Annotations: map[string]string{"common-annotation": "master-val"},
								},
							},
							Worker: datav1alpha1.JindoCompTemplateSpec{
								PodMetadata: datav1alpha1.PodMetadata{
									Labels:      map[string]string{"common-key": "worker-value"},
									Annotations: map[string]string{"common-annotation": "worker-val"},
								},
							},
						},
					},
					Value: &Jindo{},
					wantValue: &Jindo{
						Master: Master{
							Labels:      map[string]string{"common-key": "master-value"},
							Annotations: map[string]string{"common-annotation": "master-val"}},
						Worker: Worker{
							Labels:      map[string]string{"common-key": "worker-value"},
							Annotations: map[string]string{"common-annotation": "worker-val"},
						},
						Fuse: Fuse{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
					},
				},
			}

			for _, tt := range testCases {
				err := engine.transformPodMetadata(tt.Runtime, tt.Value)
				Expect(err).NotTo(HaveOccurred())
				Expect(reflect.DeepEqual(tt.Value, tt.wantValue)).To(BeTrue())
			}
		})
	})

	Describe("transformLogConfig", func() {
		It("should transform log config correctly", func() {
			tests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				jindoValue  *Jindo
				expect      string
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						LogConfig: map[string]string{"logger.level": "6"},
						Fuse: datav1alpha1.JindoFuseSpec{
							LogConfig: map[string]string{"logger.level": "6"},
						},
					},
				}, &Jindo{}, "6"},
			}
			for _, test := range tests {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				engine.transformLogConfig(test.testRuntime, test.jindoValue)
				Expect(test.jindoValue.LogConfig["logger.level"]).To(Equal(test.expect))
				Expect(test.jindoValue.FuseLogConfig["logger.level"]).To(Equal(test.expect))
			}
		})
	})

	Describe("transformEnvVariables", func() {
		DescribeTable("should transform env variables correctly",
			func(name, namespace string, testRuntime *datav1alpha1.JindoRuntime, value *Jindo,
				expectMasterEnvs, expectWorkerEnvs, expectFuseEnvs map[string]string) {
				e := &JindoCacheEngine{
					name:      name,
					namespace: namespace,
					Log:       fake.NullLogger(),
				}
				e.transformEnvVariables(testRuntime, value)
				Expect(reflect.DeepEqual(value.Master.Env, expectMasterEnvs)).To(BeTrue())
				Expect(reflect.DeepEqual(value.Worker.Env, expectWorkerEnvs)).To(BeTrue())
				Expect(reflect.DeepEqual(value.Fuse.Env, expectFuseEnvs)).To(BeTrue())
			},
			Entry("no_env_variable",
				"test-no-env", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-no-env",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&Jindo{},
				nil, nil, nil,
			),
			Entry("all_env_variable_set",
				"test-all-env-set", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-all-env-set",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Env: map[string]string{
								"test-master": "foo",
							},
						},
						Worker: datav1alpha1.JindoCompTemplateSpec{
							Env: map[string]string{
								"test-worker": "bar",
							},
						},
						Fuse: datav1alpha1.JindoFuseSpec{
							Env: map[string]string{
								"test-fuse": "test",
							},
						},
					},
				},
				&Jindo{},
				map[string]string{"test-master": "foo"},
				map[string]string{"test-worker": "bar"},
				map[string]string{"test-fuse": "test"},
			),
		)
	})

	Describe("checkIfSupportSecretMount", func() {
		DescribeTable("should check secret mount support correctly",
			func(testRuntime *datav1alpha1.JindoRuntime, smartdataTag, fuseTag string, expect bool) {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				result := engine.checkIfSupportSecretMount(testRuntime, smartdataTag, fuseTag)
				Expect(result).To(Equal(expect))
			},
			Entry("4.5.2 both tags",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{},
				}, "4.5.2", "4.5.2", false),
			Entry("4.5.2 smartdata, 6.1.1 fuse",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{},
				}, "4.5.2", "6.1.1", false),
			Entry("6.1.1 both tags",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{},
				}, "6.1.1", "6.1.1", true),
			Entry("disabled master and worker",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Disabled: true,
						},
						Worker: datav1alpha1.JindoCompTemplateSpec{
							Disabled: true,
						},
					},
				}, "4.5.2", "6.1.1", true),
		)
	})

	Describe("transformPolicy", func() {
		DescribeTable("should transform policy correctly",
			func(name, namespace string, testRuntime *datav1alpha1.JindoRuntime, dataset *datav1alpha1.Dataset, wantErr bool) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, testRuntime.DeepCopy())
				runtimeObjs = append(runtimeObjs, dataset.DeepCopy())
				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
				s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.DatasetList{})
				_ = corev1.AddToScheme(s)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "jinocache")
				Expect(err).NotTo(HaveOccurred())

				e := &JindoCacheEngine{
					runtime:     testRuntime,
					name:        name,
					namespace:   namespace,
					Client:      client,
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfo,
				}
				err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
				Expect(err).NotTo(HaveOccurred())

				_, err = e.transform(testRuntime)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("WRITE_THROUGH_ALWAYS",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
								Options: map[string]string{
									"writePolicy": "WRITE_THROUGH",
									"metaPolicy":  "ALWAYS",
								},
							},
						},
					},
				},
				false,
			),
			Entry("CACHE_ONLY_ONCE",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
								Options: map[string]string{
									"writePolicy": "CACHE_ONLY",
									"metaPolicy":  "ONCE",
								},
							},
						},
					},
				},
				false,
			),
			Entry("CACHE_ONLY_ALWAYS error",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data/subpath",
								Options: map[string]string{
									"writePolicy": "CACHE_ONLY",
									"metaPolicy":  "ALWAYS",
								},
							},
						},
					},
				},
				true,
			),
		)
	})

	Describe("transformMasterVolume", func() {
		DescribeTable("should transform master volume correctly",
			func(testRuntime *datav1alpha1.JindoRuntime, jindoValue *Jindo, expect int) {
				engine := &JindoCacheEngine{Log: fake.NullLogger()}
				err := engine.transformMasterVolumes(testRuntime, jindoValue)
				if err != nil {
					GinkgoWriter.Println(err)
				}
				Expect(len(jindoValue.Master.VolumeMounts)).To(Equal(expect))
			},
			Entry("matching volume and mount",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "nas",
								MountPath: "test",
								SubPath:   "/test",
							}},
						},
						Volumes: []corev1.Volume{{
							Name: "nas",
						}},
					},
				}, &Jindo{}, 1),
			Entry("non-matching volume and mount",
				&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "nas",
								MountPath: "test",
								SubPath:   "/test",
							}},
						},
						Volumes: []corev1.Volume{{
							Name: "nas-test",
						}},
					},
				}, &Jindo{}, 0),
		)
	})

	Describe("transformCacheSet", func() {
		DescribeTable("should transform cache set correctly",
			func(name, namespace string, testRuntime *datav1alpha1.JindoRuntime, dataset *datav1alpha1.Dataset, wantErr bool) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, testRuntime.DeepCopy())
				runtimeObjs = append(runtimeObjs, dataset.DeepCopy())
				s := runtime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, testRuntime)
				s.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
				s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.DatasetList{})
				_ = corev1.AddToScheme(s)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
				runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "jindocache")
				Expect(err).NotTo(HaveOccurred())

				e := &JindoCacheEngine{
					runtime:     testRuntime,
					name:        name,
					namespace:   namespace,
					Client:      client,
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfo,
				}
				err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
				Expect(err).NotTo(HaveOccurred())

				_, err = e.transform(testRuntime)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("valid cache replica",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
								Options: map[string]string{
									"readCacheReplica":  "1",
									"writeCacheReplica": "1",
								},
							},
						},
					},
				},
				false,
			),
			Entry("invalid read cache replica",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data/subpath",
								Options: map[string]string{
									"readCacheReplica":  "xx",
									"writeCacheReplica": "1",
								},
							},
						},
					},
				},
				true,
			),
			Entry("invalid write cache replica",
				"test", "default",
				&datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				&datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data/subpath",
								Options: map[string]string{
									"readCacheReplica":  "1",
									"writeCacheReplica": "yy",
								},
							},
						},
					},
				},
				true,
			),
		)
	})

	Describe("transformSecret", func() {
		It("should transform secret correctly", func() {
			jindocacheSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Data: map[string][]byte{
					"AccessKeyId":     []byte("test"),
					"AccessKeySecret": []byte("test"),
				},
			}
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, (*jindocacheSecret).DeepCopy())

			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			engine := JindoCacheEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				Log:       fake.NullLogger(),
				runtime: &datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Fuse: datav1alpha1.JindoFuseSpec{},
					},
				},
			}
			ctrl.SetLogger(zap.New(func(o *zap.Options) {
				o.Development = true
			}))

			tests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				dataset     *datav1alpha1.Dataset
				value       *Jindo
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Fuse: datav1alpha1.JindoFuseSpec{},
						Worker: datav1alpha1.JindoCompTemplateSpec{
							Replicas:     2,
							Resources:    corev1.ResourceRequirements{},
							Env:          nil,
							NodeSelector: nil,
						},
					},
				}, &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{{
							MountPoint: "pvc:///mnt/test",
							Name:       "test",
							EncryptOptions: []datav1alpha1.EncryptOption{{
								Name: "fs.oss.accessKeyId",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "test",
										Key:  "AccessKeyId",
									},
								},
							},
								{
									Name: "fs.oss.accessKeySecret",
									ValueFrom: datav1alpha1.EncryptOptionSource{
										SecretKeyRef: datav1alpha1.SecretKeySelector{
											Name: "test",
											Key:  "AccessKeySecret",
										},
									},
								}},
						}},
					},
				}, &Jindo{}},
			}
			for _, test := range tests {
				err := engine.transformMaster(test.testRuntime, "/test", test.value, test.dataset, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.value.SecretKey).To(Equal("AccessKeyId"))
				Expect(test.value.SecretValue).To(Equal("AccessKeySecret"))
				Expect(test.value.Secret).To(Equal("test"))
			}
		})
	})

	Describe("transformMountpoint", func() {
		It("should transform mountpoint correctly", func() {
			jindocacheSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "fluid",
				},
				Data: map[string][]byte{
					"AccessKeyId":     []byte("test"),
					"AccessKeySecret": []byte("test"),
				},
			}
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, (*jindocacheSecret).DeepCopy())

			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			engine := JindoCacheEngine{
				name:      "test",
				namespace: "fluid",
				Client:    client,
				Log:       fake.NullLogger(),
				runtime: &datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Fuse: datav1alpha1.JindoFuseSpec{},
					},
				},
			}
			ctrl.SetLogger(zap.New(func(o *zap.Options) {
				o.Development = true
			}))

			tests := []struct {
				testRuntime *datav1alpha1.JindoRuntime
				dataset     *datav1alpha1.Dataset
				value       *Jindo
			}{
				{&datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Fuse: datav1alpha1.JindoFuseSpec{},
						Worker: datav1alpha1.JindoCompTemplateSpec{
							Replicas:     2,
							Resources:    corev1.ResourceRequirements{},
							Env:          nil,
							NodeSelector: nil,
						},
					},
				}, &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{{
							MountPoint: "dls://test/subdir",
							Name:       "test",
							Options: map[string]string{
								"fs.dls.endpoint": "oss-cn-shanghai.dls.aliyuncs.com",
								"fs.dls.region":   "oss-cn-shanghai",
							},
							EncryptOptions: []datav1alpha1.EncryptOption{{
								Name: "fs.dls.accessKeyId",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "test",
										Key:  "AccessKeyId",
									},
								},
							},
								{
									Name: "fs.dls.accessKeySecret",
									ValueFrom: datav1alpha1.EncryptOptionSource{
										SecretKeyRef: datav1alpha1.SecretKeySelector{
											Name: "test",
											Key:  "AccessKeySecret",
										},
									},
								}},
						}},
					},
				}, &Jindo{}},
			}
			for _, test := range tests {
				err := engine.transformMaster(test.testRuntime, "/test", test.value, test.dataset, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(test.value.SecretKey).To(Equal("AccessKeyId"))
				Expect(test.value.SecretValue).To(Equal("AccessKeySecret"))
				Expect(test.value.Secret).To(Equal("test"))
			}
		})
	})
})
