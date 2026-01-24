/*
Copyright 2020 The Fluid Authors.

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
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// Constants for test values used multiple times
const (
	testJnifuseKey           = "alluxio.fuse.jnifuse.enabled"
	testBlockSizeKey         = "alluxio.user.block.size.bytes.default"
	testMasterRpcPortKey     = "alluxio.master.rpc.port"
	testMountPath            = "/mnt/runtime"
	testExpectedBlockSize    = "256MB"
	testExpectedJnifuseTrue  = "true"
	testExpectedJnifuseFalse = "false"
)

var _ = Describe("AlluxioEngine Transform Optimization Tests", Label("pkg.ddc.alluxio.transform_optimization_test.go"), func() {
	var engine *AlluxioEngine

	BeforeEach(func() {
		engine = &AlluxioEngine{}
	})

	Describe("optimizeDefaultProperties", func() {
		Context("when no properties are set in runtime", func() {
			It("should set default jnifuse property to true", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Properties: map[string]string{},
					},
				}
				alluxioValue := &Alluxio{}

				engine.optimizeDefaultProperties(runtime, alluxioValue)

				Expect(alluxioValue.Properties[testJnifuseKey]).To(Equal(testExpectedJnifuseTrue))
			})
		})

		Context("when jnifuse property is already set to false", func() {
			It("should preserve the existing value", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Properties: map[string]string{
							testJnifuseKey: testExpectedJnifuseFalse,
						},
					},
				}
				alluxioValue := &Alluxio{}

				engine.optimizeDefaultProperties(runtime, alluxioValue)

				Expect(alluxioValue.Properties[testJnifuseKey]).To(Equal(testExpectedJnifuseFalse))
			})
		})
	})

	Describe("optimizeDefaultPropertiesAndFuseForHTTP", func() {
		Context("when dataset has HTTPS mount point", func() {
			It("should set block size property for HTTP", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Properties: map[string]string{},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
					Fuse: Fuse{
						Args: []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072,attr_timeout=7200,entry_timeout=7200,nonempty"},
					},
				}
				dataset := &datav1alpha1.Dataset{
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{MountPoint: "https://mirrors.bit.edu.cn/apache/zookeeper/zookeeper-3.6.2/"},
						},
					},
				}

				engine.optimizeDefaultProperties(runtime, alluxioValue)
				engine.optimizeDefaultPropertiesAndFuseForHTTP(runtime, dataset, alluxioValue)

				Expect(alluxioValue.Properties[testBlockSizeKey]).To(Equal(testExpectedBlockSize))
			})
		})
	})

	Describe("setDefaultProperties", func() {
		Context("when property is not set in runtime", func() {
			It("should set the default value in alluxioValue", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Properties: map[string]string{},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}

				setDefaultProperties(runtime, alluxioValue, testJnifuseKey, testExpectedJnifuseTrue)

				Expect(alluxioValue.Properties[testJnifuseKey]).To(Equal(testExpectedJnifuseTrue))
			})
		})

		Context("when property is already set in runtime", func() {
			It("should NOT set the default value in alluxioValue (leave it empty)", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Properties: map[string]string{
							testJnifuseKey: testExpectedJnifuseFalse,
						},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}

				setDefaultProperties(runtime, alluxioValue, testJnifuseKey, testExpectedJnifuseTrue)

				// When property exists in runtime, setDefaultProperties does NOT set it in alluxioValue
				Expect(alluxioValue.Properties[testJnifuseKey]).To(BeEmpty())
			})
		})
	})

	Describe("optimizeDefaultForMaster", func() {
		Context("when no JVM options are set in runtime", func() {
			It("should set default JVM options including UnlockExperimentalVMOptions", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}

				engine.optimizeDefaultForMaster(runtime, alluxioValue)

				Expect(alluxioValue.Master.JvmOptions).To(HaveLen(2))
				Expect(alluxioValue.Master.JvmOptions[1]).To(Equal("-XX:+UnlockExperimentalVMOptions"))
			})
		})

		Context("when JVM options are specified in runtime", func() {
			It("should use the runtime JVM options", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Master: datav1alpha1.AlluxioCompTemplateSpec{
							JvmOptions: []string{"-Xmx4G"},
						},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
					Master:     Master{},
				}

				engine.optimizeDefaultForMaster(runtime, alluxioValue)

				Expect(alluxioValue.Master.JvmOptions).To(HaveLen(1))
				Expect(alluxioValue.Master.JvmOptions[0]).To(Equal("-Xmx4G"))
			})
		})
	})

	Describe("optimizeDefaultForWorker", func() {
		Context("when no JVM options are set in runtime", func() {
			It("should set default JVM options including UnlockExperimentalVMOptions", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}

				engine.optimizeDefaultForWorker(runtime, alluxioValue)

				Expect(alluxioValue.Worker.JvmOptions).To(HaveLen(3))
				Expect(alluxioValue.Worker.JvmOptions[1]).To(Equal("-XX:+UnlockExperimentalVMOptions"))
			})
		})

		Context("when JVM options are specified in runtime", func() {
			It("should use the runtime JVM options", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Worker: datav1alpha1.AlluxioCompTemplateSpec{
							JvmOptions: []string{"-Xmx4G"},
						},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}

				engine.optimizeDefaultForWorker(runtime, alluxioValue)

				Expect(alluxioValue.Worker.JvmOptions).To(HaveLen(1))
				Expect(alluxioValue.Worker.JvmOptions[0]).To(Equal("-Xmx4G"))
			})
		})
	})

	Describe("optimizeDefaultFuse", func() {
		Context("when no JVM options are set with new fuse arg version", func() {
			It("should set default JVM options and append mount path to args", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
					Fuse: Fuse{
						MountPath: testMountPath,
					},
				}
				isNewFuseArgVersion := true

				engine.optimizeDefaultFuse(runtime, alluxioValue, isNewFuseArgVersion)

				expectedJvmOptions := []string{
					"-Xmx16G",
					"-Xms16G",
					"-XX:+UseG1GC",
					"-XX:MaxDirectMemorySize=32g",
					"-XX:+UnlockExperimentalVMOptions",
				}
				expectedArgs := []string{"fuse", "--fuse-opts=kernel_cache,rw", testMountPath, "/"}

				Expect(alluxioValue.Fuse.JvmOptions).To(Equal(expectedJvmOptions))
				Expect(alluxioValue.Fuse.Args).To(Equal(expectedArgs))
			})
		})

		Context("when no JVM options are set with old fuse arg version", func() {
			It("should set default JVM options without mount path in args", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
					Fuse: Fuse{
						MountPath: testMountPath,
					},
				}
				isNewFuseArgVersion := false

				engine.optimizeDefaultFuse(runtime, alluxioValue, isNewFuseArgVersion)

				expectedJvmOptions := []string{
					"-Xmx16G",
					"-Xms16G",
					"-XX:+UseG1GC",
					"-XX:MaxDirectMemorySize=32g",
					"-XX:+UnlockExperimentalVMOptions",
				}
				expectedArgs := []string{"fuse", "--fuse-opts=kernel_cache,rw"}

				Expect(alluxioValue.Fuse.JvmOptions).To(Equal(expectedJvmOptions))
				Expect(alluxioValue.Fuse.Args).To(Equal(expectedArgs))
			})
		})

		Context("when JVM options are specified in runtime", func() {
			It("should use the runtime JVM options", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Fuse: datav1alpha1.AlluxioFuseSpec{
							JvmOptions: []string{"-Xmx4G"},
						},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}
				isNewFuseArgVersion := true

				engine.optimizeDefaultFuse(runtime, alluxioValue, isNewFuseArgVersion)

				Expect(alluxioValue.Fuse.JvmOptions).To(HaveLen(1))
				Expect(alluxioValue.Fuse.JvmOptions[0]).To(Equal("-Xmx4G"))
			})
		})

		Context("when fuse args are specified with new fuse arg version", func() {
			It("should append mount path and root to args", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Fuse: datav1alpha1.AlluxioFuseSpec{
							Args: []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"},
						},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
					Fuse: Fuse{
						MountPath: testMountPath,
					},
				}
				isNewFuseArgVersion := true

				engine.optimizeDefaultFuse(runtime, alluxioValue, isNewFuseArgVersion)

				expectedArgs := []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072", testMountPath, "/"}
				Expect(alluxioValue.Fuse.Args).To(Equal(expectedArgs))
			})
		})

		Context("when fuse args are specified with old fuse arg version", func() {
			It("should not append mount path to args", func() {
				runtime := &datav1alpha1.AlluxioRuntime{
					Spec: datav1alpha1.AlluxioRuntimeSpec{
						Fuse: datav1alpha1.AlluxioFuseSpec{
							Args: []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"},
						},
					},
				}
				alluxioValue := &Alluxio{
					Properties: map[string]string{},
				}
				isNewFuseArgVersion := false

				engine.optimizeDefaultFuse(runtime, alluxioValue, isNewFuseArgVersion)

				expectedArgs := []string{"fuse", "--fuse-opts=kernel_cache,rw,max_read=131072"}
				Expect(alluxioValue.Fuse.Args).To(Equal(expectedArgs))
			})
		})
	})

	Describe("setPortProperties", func() {
		Context("when port values are specified in alluxioValue", func() {
			It("should set all port properties correctly", func() {
				port := 20000
				runtime := &datav1alpha1.AlluxioRuntime{}
				alluxioValue := &Alluxio{
					Master: Master{
						Ports: Ports{
							Rpc:      port,
							Web:      port,
							Embedded: 0,
						},
					},
					Worker: Worker{
						Ports: Ports{
							Rpc: port,
							Web: port,
						},
					},
					JobMaster: JobMaster{
						Ports: Ports{
							Rpc:      port,
							Web:      port,
							Embedded: 0,
						},
						Resources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
						},
					},
					JobWorker: JobWorker{
						Ports: Ports{
							Rpc:  port,
							Web:  port,
							Data: port,
						},
						Resources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
						},
					},
					Properties: map[string]string{},
				}

				testEngine := &AlluxioEngine{
					runtime: runtime,
				}
				testEngine.setPortProperties(runtime, alluxioValue)

				Expect(alluxioValue.Properties[testMasterRpcPortKey]).To(Equal(strconv.Itoa(port)))
			})
		})
	})
})
