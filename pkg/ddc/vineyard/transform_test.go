/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	transformTestNamespace    = "fluid"
	transformTestName         = "test"
	transformTestVineyard     = "vineyard"
	transformTestImageFromEnv = "image-from-env"
	transformTestTagFromEnv   = "image-tag-from-env"
)

var (
	LocalMount = datav1alpha1.Mount{
		MountPoint: "local:///mnt/test",
		Name:       "local",
	}
)

var _ = Describe("VineyardEngine Transform", Label("pkg.ddc.vineyard.transform_test.go"), func() {
	Describe("transformFuse", func() {
		BeforeEach(func() {
			ctrl.SetLogger(zap.New(func(o *zap.Options) {
				o.Development = true
			}))
		})

		Context("when image, imageTag and pullPolicy are set directly", func() {
			It("should use the specified values", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Fuse: datav1alpha1.VineyardClientSocketSpec{
							Image:           "dummy-fuse-image",
							ImageTag:        "dummy-tag",
							ImagePullPolicy: "IfNotPresent",
							Env:             map[string]string{"TEST_ENV": "true"},
							CleanPolicy:     "OnRuntimeDeleted",
						},
					},
				}
				value := &Vineyard{}

				runtimeInfo, err := base.BuildRuntimeInfo(transformTestName, transformTestNamespace, transformTestVineyard)
				Expect(err).NotTo(HaveOccurred())

				engine := &VineyardEngine{
					runtimeInfo: runtimeInfo,
					Log:         ctrl.Log,
				}
				engine.transformFuse(runtime, value)

				Expect(value.Fuse.Image).To(Equal("dummy-fuse-image"))
				Expect(value.Fuse.ImageTag).To(Equal("dummy-tag"))
				Expect(value.Fuse.ImagePullPolicy).To(Equal("IfNotPresent"))
				Expect(string(value.Fuse.CleanPolicy)).To(Equal("OnRuntimeDeleted"))
				Expect(value.Fuse.Env).To(Equal(map[string]string{"TEST_ENV": "true"}))
			})
		})

		Context("when image, imageTag come from environment", func() {
			BeforeEach(func() {
				err := os.Setenv("VINEYARD_FUSE_IMAGE_ENV", "image-from-env:image-tag-from-env")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.Unsetenv("VINEYARD_FUSE_IMAGE_ENV")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use values from environment", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Fuse: datav1alpha1.VineyardClientSocketSpec{
							ImagePullPolicy: "IfNotPresent",
							Env:             map[string]string{"TEST_ENV": "true"},
							CleanPolicy:     "OnRuntimeDeleted",
						},
					},
				}
				value := &Vineyard{}

				runtimeInfo, err := base.BuildRuntimeInfo(transformTestName, transformTestNamespace, transformTestVineyard)
				Expect(err).NotTo(HaveOccurred())

				engine := &VineyardEngine{
					runtimeInfo: runtimeInfo,
					Log:         ctrl.Log,
				}
				engine.transformFuse(runtime, value)

				Expect(value.Fuse.Image).To(Equal(transformTestImageFromEnv))
				Expect(value.Fuse.ImageTag).To(Equal(transformTestTagFromEnv))
				Expect(value.Fuse.ImagePullPolicy).To(Equal("IfNotPresent"))
				Expect(string(value.Fuse.CleanPolicy)).To(Equal("OnRuntimeDeleted"))
				Expect(value.Fuse.Env).To(Equal(map[string]string{"TEST_ENV": "true"}))
			})
		})
	})

	Describe("transformMasters", func() {
		Context("when image, imageTag and pullPolicy are set directly", func() {
			It("should use the specified values", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								Image:           "test-image",
								ImageTag:        "test-tag",
								ImagePullPolicy: "IfNotPresent",
							},
						},
					},
				}

				engine := &VineyardEngine{Log: fake.NullLogger()}
				ds := &datav1alpha1.Dataset{}
				gotValue := &Vineyard{}

				err := engine.transformMasters(runtime, ds, gotValue)
				Expect(err).NotTo(HaveOccurred())
				Expect(gotValue.Master.Image).To(Equal("test-image"))
				Expect(gotValue.Master.ImageTag).To(Equal("test-tag"))
				Expect(gotValue.Master.ImagePullPolicy).To(Equal("IfNotPresent"))
			})
		})

		Context("when image, imageTag come from environment", func() {
			BeforeEach(func() {
				err := os.Setenv("VINEYARD_MASTER_IMAGE_ENV", "image-from-env:image-tag-from-env")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.Unsetenv("VINEYARD_MASTER_IMAGE_ENV")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use values from environment", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								ImagePullPolicy: "IfNotPresent",
							},
						},
					},
				}

				engine := &VineyardEngine{Log: fake.NullLogger()}
				ds := &datav1alpha1.Dataset{}
				gotValue := &Vineyard{}

				err := engine.transformMasters(runtime, ds, gotValue)
				Expect(err).NotTo(HaveOccurred())
				Expect(gotValue.Master.Image).To(Equal(transformTestImageFromEnv))
				Expect(gotValue.Master.ImageTag).To(Equal(transformTestTagFromEnv))
				Expect(gotValue.Master.ImagePullPolicy).To(Equal("IfNotPresent"))
			})
		})
	})

	Describe("transformMasterSelector", func() {
		Context("when no NodeSelector is set", func() {
			It("should return empty map", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformMasterSelector(runtime)
				Expect(actual).To(BeEmpty())
			})
		})

		Context("when NodeSelector is set", func() {
			It("should return the specified selector", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								NodeSelector: map[string]string{"disktype": "ssd"},
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformMasterSelector(runtime)
				Expect(actual).To(HaveKeyWithValue("disktype", "ssd"))
			})
		})
	})

	Describe("transformMasterPorts", func() {
		Context("when no ports are set", func() {
			It("should return default ports", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								Ports: map[string]int{},
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformMasterPorts(runtime)
				Expect(actual).To(HaveKeyWithValue("client", 2379))
				Expect(actual).To(HaveKeyWithValue("peer", 2380))
			})
		})

		Context("when ports are set", func() {
			It("should return the specified ports", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								Ports: map[string]int{
									"client": 1234,
									"peer":   5678,
								},
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformMasterPorts(runtime)
				Expect(actual).To(HaveKeyWithValue("client", 1234))
				Expect(actual).To(HaveKeyWithValue("peer", 5678))
			})
		})
	})

	Describe("transformMasterOptions", func() {
		Context("when no options are set", func() {
			It("should return default options", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								Options: map[string]string{},
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformMasterOptions(runtime)
				Expect(actual).To(HaveKeyWithValue("vineyardd.reserve.memory", "true"))
				Expect(actual).To(HaveKeyWithValue("etcd.prefix", "/vineyard"))
			})
		})

		Context("when options are set", func() {
			It("should return the specified options", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								Options: map[string]string{
									"vineyardd.reserve.memory": "false",
									"etcd.prefix":              "/vineyard-test",
								},
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformMasterOptions(runtime)
				Expect(actual).To(HaveKeyWithValue("vineyardd.reserve.memory", "false"))
				Expect(actual).To(HaveKeyWithValue("etcd.prefix", "/vineyard-test"))
			})
		})
	})

	Describe("transformWorkers", func() {
		Context("when image, imageTag and pullPolicy are set directly", func() {
			It("should use the specified values", func() {
				vineyardRuntime := &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      transformTestName,
						Namespace: transformTestNamespace,
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Replicas: 3,
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: "MEM",
									Quota:      resource.NewQuantity(1024*1024*1024, resource.BinarySI),
								},
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas:        2,
							Image:           "test-image",
							ImageTag:        "test-tag",
							ImagePullPolicy: "IfNotPresent",
						},
					},
				}

				runtimeInfo, err := base.BuildRuntimeInfo(transformTestName, transformTestNamespace, transformTestVineyard,
					base.WithTieredStore(vineyardRuntime.Spec.TieredStore))
				Expect(err).NotTo(HaveOccurred())

				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{},
				}
				runtimeInfo.SetupWithDataset(dataset)

				s := k8sruntime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, vineyardRuntime)
				err = datav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime)

				engine := &VineyardEngine{
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfo,
					Client:      mockClient,
					name:        transformTestName,
					namespace:   transformTestNamespace,
				}

				value := &Vineyard{}
				err = engine.transformWorkers(vineyardRuntime, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Worker.Image).To(Equal("test-image"))
				Expect(value.Worker.ImageTag).To(Equal("test-tag"))
				Expect(value.Worker.ImagePullPolicy).To(Equal("IfNotPresent"))
			})
		})

		Context("when image and tag come from environment", func() {
			BeforeEach(func() {
				err := os.Setenv("VINEYARD_WORKER_IMAGE_ENV", "image-from-env:image-tag-from-env")
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				err := os.Unsetenv("VINEYARD_WORKER_IMAGE_ENV")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should use image from environment", func() {
				vineyardRuntime := &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      transformTestName,
						Namespace: transformTestNamespace,
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Replicas: 3,
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: "MEM",
									Quota:      resource.NewQuantity(1024*1024*1024, resource.BinarySI),
								},
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas:        2,
							ImagePullPolicy: "IfNotPresent",
						},
					},
				}

				runtimeInfo, err := base.BuildRuntimeInfo(transformTestName, transformTestNamespace, transformTestVineyard,
					base.WithTieredStore(vineyardRuntime.Spec.TieredStore))
				Expect(err).NotTo(HaveOccurred())

				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{},
				}
				runtimeInfo.SetupWithDataset(dataset)

				s := k8sruntime.NewScheme()
				s.AddKnownTypes(datav1alpha1.GroupVersion, vineyardRuntime)
				err = datav1alpha1.AddToScheme(s)
				Expect(err).NotTo(HaveOccurred())

				mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime)

				engine := &VineyardEngine{
					Log:         fake.NullLogger(),
					runtimeInfo: runtimeInfo,
					Client:      mockClient,
					name:        transformTestName,
					namespace:   transformTestNamespace,
				}

				value := &Vineyard{}
				err = engine.transformWorkers(vineyardRuntime, value)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Worker.Image).To(Equal("image-from-env"))
				Expect(value.Worker.ImageTag).To(Equal("image-tag-from-env"))
				Expect(value.Worker.ImagePullPolicy).To(Equal("IfNotPresent"))
			})
		})
	})

	Describe("transformWorkerOptions", func() {
		Context("when no options are set", func() {
			It("should return empty map", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Options: map[string]string{},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformWorkerOptions(runtime)
				Expect(actual).To(BeEmpty())
			})
		})

		Context("when options are set", func() {
			It("should return the specified options", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Options: map[string]string{
								"dummy-key": "dummy-value",
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformWorkerOptions(runtime)
				Expect(actual).To(HaveKeyWithValue("dummy-key", "dummy-value"))
			})
		})
	})

	Describe("transformFuseOptions", func() {
		Context("when no fuse options are set", func() {
			It("should return default options", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Fuse: datav1alpha1.VineyardClientSocketSpec{
							Options: map[string]string{},
						},
					},
				}
				value := &Vineyard{
					FullnameOverride: transformTestVineyard,
					Master: Master{
						Ports: map[string]int{
							"client": 2379,
						},
					},
				}

				engine := &VineyardEngine{
					namespace: "default",
				}
				actual := engine.transformFuseOptions(runtime, value)
				Expect(actual).To(HaveKeyWithValue("size", "0"))
				Expect(actual).To(HaveKeyWithValue("etcd_endpoint", "http://vineyard-master-0.vineyard-master.default:2379"))
				Expect(actual).To(HaveKeyWithValue("etcd_prefix", "/vineyard"))
			})
		})

		Context("when fuse options are set", func() {
			It("should return the specified options", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Fuse: datav1alpha1.VineyardClientSocketSpec{
							Options: map[string]string{
								"size":           "10Gi",
								"etcd_endpoint":  "http://vineyard-master-0.vineyard-master.default:12379",
								"reserve_memory": "true",
							},
						},
					},
				}
				value := &Vineyard{
					Master: Master{
						Ports: map[string]int{
							"client": 12379,
						},
					},
				}

				engine := &VineyardEngine{
					namespace: "default",
				}
				actual := engine.transformFuseOptions(runtime, value)
				Expect(actual).To(HaveKeyWithValue("size", "10Gi"))
				Expect(actual).To(HaveKeyWithValue("etcd_endpoint", "http://vineyard-master-0.vineyard-master.default:12379"))
				Expect(actual).To(HaveKeyWithValue("etcd_prefix", "/vineyard"))
				Expect(actual).To(HaveKeyWithValue("reserve_memory", "true"))
			})
		})
	})

	Describe("transformWorkerPorts", func() {
		Context("when no ports are set", func() {
			It("should return default ports", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Ports: map[string]int{},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformWorkerPorts(runtime)
				Expect(actual).To(HaveKeyWithValue("rpc", 9600))
				Expect(actual).To(HaveKeyWithValue("exporter", 9144))
			})
		})

		Context("when ports are set", func() {
			It("should return the specified ports", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Ports: map[string]int{
								"rpc":      1234,
								"exporter": 5678,
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual := engine.transformWorkerPorts(runtime)
				Expect(actual).To(HaveKeyWithValue("rpc", 1234))
				Expect(actual).To(HaveKeyWithValue("exporter", 5678))
			})
		})
	})

	Describe("transformFuseNodeSelector", func() {
		Context("when no NodeSelector is set", func() {
			It("should return fluid label", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      transformTestVineyard,
						Namespace: transformTestNamespace,
					},
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							NodeSelector: map[string]string{},
						},
					},
				}

				runtimeInfo, err := base.BuildRuntimeInfo(transformTestName, transformTestNamespace, transformTestVineyard)
				Expect(err).NotTo(HaveOccurred())

				engine := &VineyardEngine{
					name:        transformTestVineyard,
					namespace:   transformTestNamespace,
					runtimeInfo: runtimeInfo,
				}
				actual := engine.transformFuseNodeSelector(runtime)
				Expect(actual).To(Equal(map[string]string{"fluid.io/f-fluid-vineyard": "true"}))
			})
		})
	})

	Describe("transformTieredStore", func() {
		Context("when no tiered store is set", func() {
			It("should return error", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{},
					},
				}

				engine := &VineyardEngine{}
				_, err := engine.transformTieredStore(runtime)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when tiered store is set", func() {
			It("should return the specified tiered store", func() {
				quota := resource.MustParse("20Gi")
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: "MEM",
									Quota:      &quota,
								},
							},
						},
					},
				}

				engine := &VineyardEngine{}
				actual, err := engine.transformTieredStore(runtime)
				Expect(err).NotTo(HaveOccurred())
				Expect(actual.Levels).To(HaveLen(1))
				Expect(string(actual.Levels[0].MediumType)).To(Equal("MEM"))
				Expect(actual.Levels[0].Quota.String()).To(Equal("20Gi"))
			})
		})
	})

	Describe("allocatePorts", func() {
		BeforeEach(func() {
			pr := net.ParsePortRangeOrDie("14000-16000")
			dummyPorts := func(client client.Client) (ports []int, err error) {
				return []int{14000, 14001, 14002, 14003}, nil
			}
			err := portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummyPorts)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when using ContainerNetwork", func() {
			It("should not allocate ports", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								NetworkMode: "ContainerNetwork",
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							NetworkMode: "ContainerNetwork",
						},
					},
				}
				value := &Vineyard{
					Master: Master{
						Ports: map[string]int{
							MasterClientName: MasterClientPort,
							MasterPeerName:   MasterPeerPort,
						},
					},
					Worker: Worker{
						Ports: map[string]int{
							WorkerRPCName:      WorkerRPCPort,
							WorkerExporterName: WorkerExporterPort,
						},
					},
				}

				engine := &VineyardEngine{}
				err := engine.allocatePorts(value, runtime)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Master.Ports).To(HaveKeyWithValue(MasterClientName, MasterClientPort))
				Expect(value.Master.Ports).To(HaveKeyWithValue(MasterPeerName, MasterPeerPort))
				Expect(value.Worker.Ports).To(HaveKeyWithValue(WorkerRPCName, WorkerRPCPort))
				Expect(value.Worker.Ports).To(HaveKeyWithValue(WorkerExporterName, WorkerExporterPort))
			})
		})

		Context("when using HostNetwork", func() {
			It("should allocate ports within range", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								NetworkMode: "HostNetwork",
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							NetworkMode: "HostNetwork",
						},
					},
				}
				value := &Vineyard{
					Master: Master{
						Ports: map[string]int{
							MasterClientName: MasterClientPort,
							MasterPeerName:   MasterPeerPort,
						},
					},
					Worker: Worker{
						Ports: map[string]int{
							WorkerRPCName:      WorkerRPCPort,
							WorkerExporterName: WorkerExporterPort,
						},
					},
				}

				engine := &VineyardEngine{}
				err := engine.allocatePorts(value, runtime)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(value.Master.Ports)).To(Equal(2))
				Expect(len(value.Worker.Ports)).To(Equal(2))
				Expect(value.Master.Ports[MasterClientName]).To(BeNumerically(">=", 14000))
				Expect(value.Master.Ports[MasterClientName]).To(BeNumerically("<=", 16000))
				Expect(value.Master.Ports[MasterPeerName]).To(BeNumerically(">=", 14000))
				Expect(value.Master.Ports[MasterPeerName]).To(BeNumerically("<=", 16000))
				Expect(value.Worker.Ports[WorkerRPCName]).To(BeNumerically(">=", 14000))
				Expect(value.Worker.Ports[WorkerRPCName]).To(BeNumerically("<=", 16000))
				Expect(value.Worker.Ports[WorkerExporterName]).To(BeNumerically(">=", 14000))
				Expect(value.Worker.Ports[WorkerExporterName]).To(BeNumerically("<=", 16000))
			})
		})
	})

	Describe("transformPodMetadata", func() {
		Context("when setting common labels and annotations", func() {
			It("should apply to all components", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
					},
				}
				value := &Vineyard{}

				engine := &VineyardEngine{Log: fake.NullLogger()}
				engine.transformPodMetadata(runtime, value)

				Expect(value.Master.Labels).To(HaveKeyWithValue("common-key", "common-value"))
				Expect(value.Master.Annotations).To(HaveKeyWithValue("common-annotation", "val"))
				Expect(value.Worker.Labels).To(HaveKeyWithValue("common-key", "common-value"))
				Expect(value.Worker.Annotations).To(HaveKeyWithValue("common-annotation", "val"))
				Expect(value.Fuse.Labels).To(HaveKeyWithValue("common-key", "common-value"))
				Expect(value.Fuse.Annotations).To(HaveKeyWithValue("common-annotation", "val"))
			})
		})

		Context("when setting component-specific labels and annotations", func() {
			It("should override common values", func() {
				runtime := &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "common-value"},
							Annotations: map[string]string{"common-annotation": "val"},
						},
						Master: datav1alpha1.MasterSpec{
							VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
								PodMetadata: datav1alpha1.PodMetadata{
									Labels:      map[string]string{"common-key": "master-value"},
									Annotations: map[string]string{"common-annotation": "master-val"},
								},
							},
						},
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							PodMetadata: datav1alpha1.PodMetadata{
								Labels:      map[string]string{"common-key": "worker-value"},
								Annotations: map[string]string{"common-annotation": "worker-val"},
							},
						},
					},
				}
				value := &Vineyard{}

				engine := &VineyardEngine{Log: fake.NullLogger()}
				engine.transformPodMetadata(runtime, value)

				Expect(value.Master.Labels).To(HaveKeyWithValue("common-key", "master-value"))
				Expect(value.Master.Annotations).To(HaveKeyWithValue("common-annotation", "master-val"))
				Expect(value.Worker.Labels).To(HaveKeyWithValue("common-key", "worker-value"))
				Expect(value.Worker.Annotations).To(HaveKeyWithValue("common-annotation", "worker-val"))
				Expect(value.Fuse.Labels).To(HaveKeyWithValue("common-key", "common-value"))
				Expect(value.Fuse.Annotations).To(HaveKeyWithValue("common-annotation", "val"))
			})
		})
	})
})
