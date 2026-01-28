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
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("OptimizeDefaultProperties", func() {
	It("should set UFS journal type when master replicas is 1", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}
		goosefsValue := &GooseFS{}
		engine := &GooseFSEngine{}

		engine.optimizeDefaultProperties(runtime, goosefsValue)

		Expect(goosefsValue.Properties["goosefs.master.journal.type"]).To(Equal("UFS"))
	})

	It("should set EMBEDDED journal type when master replicas is 3", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}
		goosefsValue := &GooseFS{Master: Master{
			Replicas: 3,
		}}
		engine := &GooseFSEngine{}

		engine.optimizeDefaultProperties(runtime, goosefsValue)

		Expect(goosefsValue.Properties["goosefs.master.journal.type"]).To(Equal("EMBEDDED"))
	})
})

var _ = Describe("OptimizeDefaultPropertiesAndFuseForHTTP", func() {
	It("should set block size for HTTP mount points", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}
		goosefsValue := &GooseFS{
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
		engine := &GooseFSEngine{}

		engine.optimizeDefaultProperties(runtime, goosefsValue)
		engine.optimizeDefaultPropertiesAndFuseForHTTP(runtime, dataset, goosefsValue)

		Expect(goosefsValue.Properties["goosefs.user.block.size.bytes.default"]).To(Equal("256MB"))
	})
})

var _ = Describe("OptimizeDefaultPropertiesWithSet", func() {
	It("should preserve user-set properties", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{
					"goosefs.fuse.jnifuse.enabled": "false",
				},
			},
		}
		goosefsValue := &GooseFS{}
		engine := &GooseFSEngine{}

		engine.optimizeDefaultProperties(runtime, goosefsValue)

		Expect(goosefsValue.Properties["goosefs.fuse.jnifuse.enabled"]).To(Equal("false"))
	})
})

var _ = Describe("SetDefaultPropertiesWithoutSet", func() {
	It("should set default property when not already set", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}
		goosefsValue := &GooseFS{
			Properties: map[string]string{},
		}
		key := "goosefs.fuse.jnifuse.enabled"
		value := "true"

		setDefaultProperties(runtime, goosefsValue, key, value)

		Expect(goosefsValue.Properties[key]).To(Equal("true"))
	})
})

var _ = Describe("SetDefaultPropertiesWithSet", func() {
	It("should not override user-set property", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{
					"goosefs.fuse.jnifuse.enabled": "false",
				},
			},
		}
		goosefsValue := &GooseFS{
			Properties: map[string]string{},
		}
		key := "goosefs.fuse.jnifuse.enabled"
		value := "true"

		setDefaultProperties(runtime, goosefsValue, key, value)

		Expect(goosefsValue.Properties[key]).To(BeEmpty())
	})
})

var _ = Describe("OptimizeDefaultForMaster", func() {
	Context("when no JVM options are set", func() {
		It("should set default JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			goosefsValue := &GooseFS{
				Properties: map[string]string{},
			}
			engine := &GooseFSEngine{}

			engine.optimizeDefaultForMaster(runtime, goosefsValue)

			Expect(goosefsValue.Master.JvmOptions).To(ContainElement("-Xmx16G"))
			Expect(goosefsValue.Master.JvmOptions).To(ContainElement("-XX:+UnlockExperimentalVMOptions"))
		})
	})

	Context("when JVM options are set", func() {
		It("should preserve user-set JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Master: datav1alpha1.GooseFSCompTemplateSpec{
						JvmOptions: []string{"-Xmx4G"},
					},
				},
			}
			goosefsValue := &GooseFS{
				Properties: map[string]string{},
				Master:     Master{},
			}
			engine := &GooseFSEngine{}

			engine.optimizeDefaultForMaster(runtime, goosefsValue)

			Expect(goosefsValue.Master.JvmOptions[0]).To(Equal("-Xmx4G"))
		})
	})
})

var _ = Describe("OptimizeDefaultForWorker", func() {
	Context("when no JVM options are set", func() {
		It("should set default JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			goosefsValue := &GooseFS{
				Properties: map[string]string{},
			}
			engine := &GooseFSEngine{}

			engine.optimizeDefaultForWorker(runtime, goosefsValue)

			Expect(goosefsValue.Worker.JvmOptions).To(ContainElement("-Xmx12G"))
			Expect(goosefsValue.Worker.JvmOptions).To(ContainElement("-XX:+UnlockExperimentalVMOptions"))
			Expect(goosefsValue.Worker.JvmOptions).To(ContainElement("-XX:MaxDirectMemorySize=32g"))
		})
	})

	Context("when JVM options are set", func() {
		It("should preserve user-set JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Worker: datav1alpha1.GooseFSCompTemplateSpec{
						JvmOptions: []string{"-Xmx4G"},
					},
				},
			}
			goosefsValue := &GooseFS{
				Properties: map[string]string{},
			}
			engine := &GooseFSEngine{}

			engine.optimizeDefaultForWorker(runtime, goosefsValue)

			Expect(goosefsValue.Worker.JvmOptions[0]).To(Equal("-Xmx4G"))
		})
	})
})

var _ = Describe("OptimizeDefaultForFuse", func() {
	Context("when no JVM options are set", func() {
		It("should set default JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			goosefsValue := &GooseFS{
				Properties: map[string]string{},
			}
			engine := &GooseFSEngine{}

			engine.optimizeDefaultFuse(runtime, goosefsValue)

			Expect(goosefsValue.Fuse.JvmOptions).To(ContainElement("-Xmx16G"))
			Expect(goosefsValue.Fuse.JvmOptions).To(ContainElement("-Xms16G"))
			Expect(goosefsValue.Fuse.JvmOptions).To(ContainElement("-XX:+UseG1GC"))
			Expect(goosefsValue.Fuse.JvmOptions).To(ContainElement("-XX:MaxDirectMemorySize=32g"))
			Expect(goosefsValue.Fuse.JvmOptions).To(ContainElement("-XX:+UnlockExperimentalVMOptions"))
		})
	})

	Context("when JVM options are set", func() {
		It("should preserve user-set JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{
						JvmOptions: []string{"-Xmx4G"},
					},
				},
			}
			goosefsValue := &GooseFS{
				Properties: map[string]string{},
			}
			engine := &GooseFSEngine{}

			engine.optimizeDefaultFuse(runtime, goosefsValue)

			Expect(goosefsValue.Fuse.JvmOptions[0]).To(Equal("-Xmx4G"))
		})
	})
})

var _ = Describe("GooseFSEngine setPortProperties", func() {
	It("should set port properties correctly", func() {
		var port int = 20000
		runtime := &datav1alpha1.GooseFSRuntime{}
		goosefsValue := &GooseFS{
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
				Resources: common.Resources{
					Requests: common.ResourceList{
						corev1.ResourceCPU:    "100m",
						corev1.ResourceMemory: "100Mi",
					},
				},
			},
			JobMaster: JobMaster{
				Ports: Ports{
					Rpc:      port,
					Web:      port,
					Embedded: 0,
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

		e := &GooseFSEngine{
			runtime:     runtime,
			name:        "",
			namespace:   "",
			runtimeType: "",
		}

		e.setPortProperties(runtime, goosefsValue)

		Expect(goosefsValue.Properties["goosefs.master.rpc.port"]).To(Equal(strconv.Itoa(port)))
	})
})
