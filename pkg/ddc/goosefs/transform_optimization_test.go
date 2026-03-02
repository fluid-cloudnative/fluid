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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GooseFSEngine Property Optimization", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	Context("single master replica", func() {
		It("should set journal type to UFS", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Properties: map[string]string{},
				},
			}
			value := &GooseFS{}

			engine.optimizeDefaultProperties(runtime, value)

			Expect(value.Properties["goosefs.master.journal.type"]).To(Equal("UFS"))
		})
	})

	Context("multiple master replicas", func() {
		It("should set journal type to EMBEDDED", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Properties: map[string]string{},
				},
			}
			value := &GooseFS{
				Master: Master{Replicas: 3},
			}

			engine.optimizeDefaultProperties(runtime, value)

			Expect(value.Properties["goosefs.master.journal.type"]).To(Equal("EMBEDDED"))
		})
	})

	Context("property already set in runtime", func() {
		It("should preserve existing property value", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Properties: map[string]string{
						"goosefs.fuse.jnifuse.enabled": "false",
					},
				},
			}
			value := &GooseFS{}

			engine.optimizeDefaultProperties(runtime, value)

			Expect(value.Properties["goosefs.fuse.jnifuse.enabled"]).To(Equal("false"))
		})
	})
})

var _ = Describe("GooseFSEngine HTTP Mount Optimization", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	It("should set block size for HTTP mount", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Properties: map[string]string{},
			},
		}
		value := &GooseFS{
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

		engine.optimizeDefaultProperties(runtime, value)
		engine.optimizeDefaultPropertiesAndFuseForHTTP(runtime, dataset, value)

		Expect(value.Properties["goosefs.user.block.size.bytes.default"]).To(Equal("256MB"))
	})
})

var _ = Describe("GooseFSEngine Default Property Setting", func() {
	Context("property not set in runtime", func() {
		It("should use default value", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Properties: map[string]string{},
				},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			setDefaultProperties(runtime, value, "goosefs.fuse.jnifuse.enabled", "true")

			Expect(value.Properties["goosefs.fuse.jnifuse.enabled"]).To(Equal("true"))
		})
	})

	Context("property already set in runtime", func() {
		It("should not set default when runtime has property", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Properties: map[string]string{
						"goosefs.fuse.jnifuse.enabled": "false",
					},
				},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			setDefaultProperties(runtime, value, "goosefs.fuse.jnifuse.enabled", "true")

			Expect(value.Properties).NotTo(HaveKey("goosefs.fuse.jnifuse.enabled"))
		})
	})
})

var _ = Describe("GooseFSEngine Master JVM Optimization", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	Context("no JVM options set", func() {
		It("should set default JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			engine.optimizeDefaultForMaster(runtime, value)

			Expect(value.Master.JvmOptions).To(ContainElement("-Xmx16G"))
			Expect(value.Master.JvmOptions).To(ContainElement("-XX:+UnlockExperimentalVMOptions"))
		})
	})

	Context("JVM options already set", func() {
		It("should use runtime JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Master: datav1alpha1.GooseFSCompTemplateSpec{
						JvmOptions: []string{"-Xmx4G"},
					},
				},
			}
			value := &GooseFS{
				Properties: map[string]string{},
				Master:     Master{},
			}

			engine.optimizeDefaultForMaster(runtime, value)

			Expect(value.Master.JvmOptions).To(HaveLen(1))
			Expect(value.Master.JvmOptions[0]).To(Equal("-Xmx4G"))
		})
	})
})

var _ = Describe("GooseFSEngine Worker JVM Optimization", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	Context("no JVM options set", func() {
		It("should set default JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			engine.optimizeDefaultForWorker(runtime, value)

			Expect(value.Worker.JvmOptions).To(ContainElement("-Xmx12G"))
			Expect(value.Worker.JvmOptions).To(ContainElement("-XX:+UnlockExperimentalVMOptions"))
			Expect(value.Worker.JvmOptions).To(ContainElement("-XX:MaxDirectMemorySize=32g"))
		})
	})

	Context("JVM options already set", func() {
		It("should use runtime JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Worker: datav1alpha1.GooseFSCompTemplateSpec{
						JvmOptions: []string{"-Xmx4G"},
					},
				},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			engine.optimizeDefaultForWorker(runtime, value)

			Expect(value.Worker.JvmOptions).To(HaveLen(1))
			Expect(value.Worker.JvmOptions[0]).To(Equal("-Xmx4G"))
		})
	})
})

var _ = Describe("GooseFSEngine Fuse JVM Optimization", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	Context("no JVM options set", func() {
		It("should set default JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			engine.optimizeDefaultFuse(runtime, value)

			Expect(value.Fuse.JvmOptions).To(ContainElement("-Xmx16G"))
			Expect(value.Fuse.JvmOptions).To(ContainElement("-Xms16G"))
			Expect(value.Fuse.JvmOptions).To(ContainElement("-XX:+UseG1GC"))
			Expect(value.Fuse.JvmOptions).To(ContainElement("-XX:MaxDirectMemorySize=32g"))
			Expect(value.Fuse.JvmOptions).To(ContainElement("-XX:+UnlockExperimentalVMOptions"))
		})
	})

	Context("JVM options already set", func() {
		It("should use runtime JVM options", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					Fuse: datav1alpha1.GooseFSFuseSpec{
						JvmOptions: []string{"-Xmx4G"},
					},
				},
			}
			value := &GooseFS{
				Properties: map[string]string{},
			}

			engine.optimizeDefaultFuse(runtime, value)

			Expect(value.Fuse.JvmOptions).To(HaveLen(1))
			Expect(value.Fuse.JvmOptions[0]).To(Equal("-Xmx4G"))
		})
	})
})

var _ = Describe("GooseFSEngine Port Configuration", func() {
	const testPort = 20000

	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	It("should set port properties correctly", func() {
		runtime := &datav1alpha1.GooseFSRuntime{}
		value := &GooseFS{
			Master: Master{
				Ports: Ports{
					Rpc:      testPort,
					Web:      testPort,
					Embedded: 0,
				},
			},
			Worker: Worker{
				Ports: Ports{
					Rpc: testPort,
					Web: testPort,
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
					Rpc:      testPort,
					Web:      testPort,
					Embedded: 0,
				},
			},
			JobWorker: JobWorker{
				Ports: Ports{
					Rpc:  testPort,
					Web:  testPort,
					Data: testPort,
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

		engine.setPortProperties(runtime, value)

		Expect(value.Properties["goosefs.master.rpc.port"]).To(Equal(strconv.Itoa(testPort)))
	})

	It("should set embedded journal ports when configured", func() {
		runtime := &datav1alpha1.GooseFSRuntime{}
		value := &GooseFS{
			Master: Master{
				Ports: Ports{
					Rpc:      testPort,
					Web:      testPort,
					Embedded: 19200,
				},
			},
			Worker: Worker{
				Ports: Ports{Rpc: testPort, Web: testPort},
			},
			JobMaster: JobMaster{
				Ports: Ports{
					Rpc:      testPort,
					Web:      testPort,
					Embedded: 19201,
				},
			},
			JobWorker: JobWorker{
				Ports: Ports{Rpc: testPort, Web: testPort, Data: testPort},
			},
			Properties: map[string]string{},
		}

		engine.setPortProperties(runtime, value)

		Expect(value.Properties["goosefs.master.embedded.journal.port"]).To(Equal("19200"))
		Expect(value.Properties["goosefs.job.master.embedded.journal.port"]).To(Equal("19201"))
	})

	It("should set API gateway port when enabled", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
					Enabled: true,
				},
			},
		}
		value := &GooseFS{
			Master:     Master{Ports: Ports{Rpc: testPort, Web: testPort}},
			Worker:     Worker{Ports: Ports{Rpc: testPort, Web: testPort}},
			JobMaster:  JobMaster{Ports: Ports{Rpc: testPort, Web: testPort}},
			JobWorker:  JobWorker{Ports: Ports{Rpc: testPort, Web: testPort, Data: testPort}},
			APIGateway: APIGateway{Ports: Ports{Rest: 39999}},
			Properties: map[string]string{},
		}

		engine.setPortProperties(runtime, value)

		Expect(value.Properties["goosefs.proxy.web.port"]).To(Equal("39999"))
	})

	It("should set journal addresses for HA mode", func() {
		runtime := &datav1alpha1.GooseFSRuntime{}
		value := &GooseFS{
			FullnameOverride: "test-goosefs",
			Master: Master{
				Replicas: 3,
				Ports:    Ports{Rpc: testPort, Web: testPort, Embedded: 19200},
			},
			Worker:     Worker{Ports: Ports{Rpc: testPort, Web: testPort}},
			JobMaster:  JobMaster{Ports: Ports{Rpc: testPort, Web: testPort, Embedded: 19201}},
			JobWorker:  JobWorker{Ports: Ports{Rpc: testPort, Web: testPort, Data: testPort}},
			Properties: map[string]string{},
		}

		engine.setPortProperties(runtime, value)

		Expect(value.Properties["goosefs.master.embedded.journal.addresses"]).To(ContainSubstring("test-goosefs-master-0:19200"))
		Expect(value.Properties["goosefs.master.embedded.journal.addresses"]).To(ContainSubstring("test-goosefs-master-1:19200"))
		Expect(value.Properties["goosefs.master.embedded.journal.addresses"]).To(ContainSubstring("test-goosefs-master-2:19200"))
	})
})

var _ = Describe("GooseFSEngine Fuse Args Configuration", func() {
	var engine *GooseFSEngine

	BeforeEach(func() {
		engine = &GooseFSEngine{}
	})

	It("should use runtime fuse args when set", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Fuse: datav1alpha1.GooseFSFuseSpec{
					Args: []string{"fuse", "--custom-args"},
				},
			},
		}
		value := &GooseFS{
			Properties: map[string]string{},
		}

		engine.optimizeDefaultFuse(runtime, value)

		Expect(value.Fuse.Args).To(Equal([]string{"fuse", "--custom-args"}))
	})

	It("should set default rw fuse args when not set", func() {
		runtime := &datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}
		value := &GooseFS{
			Properties: map[string]string{},
		}

		engine.optimizeDefaultFuse(runtime, value)

		Expect(value.Fuse.Args).To(ContainElement("fuse"))
		Expect(len(value.Fuse.Args)).To(BeNumerically(">=", 2))
		Expect(value.Fuse.Args[1]).To(ContainSubstring("rw"))
	})
})
