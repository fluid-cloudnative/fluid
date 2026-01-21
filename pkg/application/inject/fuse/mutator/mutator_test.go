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

package mutator

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("FindExtraArgsFromMetadata", func() {
	var (
		metaObj  metav1.ObjectMeta
		platform string
		result   map[string]string
	)

	JustBeforeEach(func() {
		result = FindExtraArgsFromMetadata(metaObj, platform)
	})

	Context("when annotations are empty", func() {
		BeforeEach(func() {
			metaObj = metav1.ObjectMeta{
				Annotations: nil,
			}
			platform = "myplatform"
		})

		It("should return empty map", func() {
			Expect(result).To(BeEmpty())
		})
	})

	Context("when annotations exist but without extra args", func() {
		BeforeEach(func() {
			metaObj = metav1.ObjectMeta{
				Annotations: map[string]string{"foo": "bar"},
			}
			platform = "myplatform"
		})

		It("should return empty map", func() {
			Expect(result).To(BeEmpty())
		})
	})

	Context("when annotations contain platform-specific extra args", func() {
		BeforeEach(func() {
			metaObj = metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo":                      "bar",
					"myplatform.fluid.io/key1": "value1",
					"myplatform.fluid.io/key2": "value2",
				},
			}
			platform = "myplatform"
		})

		It("should return only the platform-specific extra args", func() {
			Expect(result).To(HaveLen(2))
			Expect(result).To(HaveKeyWithValue("key1", "value1"))
			Expect(result).To(HaveKeyWithValue("key2", "value2"))
			Expect(result).NotTo(HaveKey("foo"))
		})
	})

	Context("when platform is empty", func() {
		BeforeEach(func() {
			metaObj = metav1.ObjectMeta{
				Annotations: map[string]string{
					"myplatform.fluid.io/key1": "value1",
				},
			}
			platform = ""
		})

		It("should return empty map", func() {
			Expect(result).To(BeEmpty())
		})
	})

	Context("when annotations have mixed platform prefixes", func() {
		BeforeEach(func() {
			metaObj = metav1.ObjectMeta{
				Annotations: map[string]string{
					"myplatform.fluid.io/key1":    "value1",
					"otherplatform.fluid.io/key2": "value2",
					"myplatform.fluid.io/key3":    "value3",
				},
			}
			platform = "myplatform"
		})

		It("should return only args matching the specified platform", func() {
			Expect(result).To(HaveLen(2))
			Expect(result).To(HaveKeyWithValue("key1", "value1"))
			Expect(result).To(HaveKeyWithValue("key3", "value3"))
			Expect(result).NotTo(HaveKey("key2"))
		})
	})

	Context("when annotations have special characters in keys", func() {
		BeforeEach(func() {
			metaObj = metav1.ObjectMeta{
				Annotations: map[string]string{
					"myplatform.fluid.io/key-with-dash":       "value1",
					"myplatform.fluid.io/key_with_underscore": "value2",
					"myplatform.fluid.io/key.with.dot":        "value3",
				},
			}
			platform = "myplatform"
		})

		It("should handle special characters in annotation keys", func() {
			Expect(result).To(HaveLen(3))
			Expect(result).To(HaveKeyWithValue("key-with-dash", "value1"))
			Expect(result).To(HaveKeyWithValue("key_with_underscore", "value2"))
			Expect(result).To(HaveKeyWithValue("key.with.dot", "value3"))
		})
	})
})

var _ = Describe("BuildMutator", func() {
	var (
		args     MutatorBuildArgs
		platform string
		mutator  Mutator
		err      error
	)

	BeforeEach(func() {
		// Setup common args
		args = MutatorBuildArgs{
			Client:    nil,
			Log:       logr.Discard(),
			Specs:     &MutatingPodSpecs{},
			Options:   common.FuseSidecarInjectOption{},
			ExtraArgs: map[string]string{},
		}
	})

	JustBeforeEach(func() {
		mutator, err = BuildMutator(args, platform)
	})

	Context("when platform is default", func() {
		BeforeEach(func() {
			platform = utils.ServerlessPlatformDefault
		})

		It("should build default mutator successfully", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator).NotTo(BeNil())
		})
	})

	Context("when platform is unprivileged", func() {
		BeforeEach(func() {
			platform = utils.ServerlessPlatformUnprivileged
		})

		It("should build unprivileged mutator successfully", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator).NotTo(BeNil())
		})
	})

	Context("when platform is unknown", func() {
		BeforeEach(func() {
			platform = "unknown-platform"
		})

		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fuse sidecar mutator cannot be found for platform"))
			Expect(err.Error()).To(ContainSubstring("unknown-platform"))
			Expect(mutator).To(BeNil())
		})
	})

	Context("when platform is empty", func() {
		BeforeEach(func() {
			platform = ""
		})

		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(mutator).To(BeNil())
		})
	})

	Context("when building mutator with extra args", func() {
		BeforeEach(func() {
			platform = utils.ServerlessPlatformDefault
			args.ExtraArgs = map[string]string{
				"arg1": "value1",
				"arg2": "value2",
			}
		})

		It("should build mutator with extra args", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator).NotTo(BeNil())
		})
	})

	Context("when building mutator with options", func() {
		BeforeEach(func() {
			platform = utils.ServerlessPlatformDefault
			args.Options = common.FuseSidecarInjectOption{
				EnableCacheDir:             true,
				SkipSidecarPostStartInject: false,
			}
		})

		It("should build mutator with options", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(mutator).NotTo(BeNil())
		})
	})
})

var _ = Describe("MutatorBuildArgs", func() {
	Describe("String method", func() {
		It("should format args correctly", func() {
			args := MutatorBuildArgs{
				Options: common.FuseSidecarInjectOption{
					EnableCacheDir: true,
				},
				ExtraArgs: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}

			result := args.String()

			Expect(result).To(ContainSubstring("options:"))
			Expect(result).To(ContainSubstring("extraArgs:"))
		})

		It("should handle empty args", func() {
			args := MutatorBuildArgs{
				Options:   common.FuseSidecarInjectOption{},
				ExtraArgs: map[string]string{},
			}

			result := args.String()

			Expect(result).NotTo(BeEmpty())
			Expect(result).To(ContainSubstring("options:"))
			Expect(result).To(ContainSubstring("extraArgs:"))
		})

		It("should handle nil extra args", func() {
			args := MutatorBuildArgs{
				Options:   common.FuseSidecarInjectOption{},
				ExtraArgs: nil,
			}

			result := args.String()

			Expect(result).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("MutatingPodSpecs", func() {

	Describe("mutating context operations", func() {
		var ctx *mutatingContext

		BeforeEach(func() {
			ctx = &mutatingContext{}
		})

		Context("GetAppendedVolumeNames", func() {
			It("should return nil when appendedVolumeNames is nil", func() {
				nameMapping, err := ctx.GetAppendedVolumeNames()
				Expect(err).NotTo(HaveOccurred())
				Expect(nameMapping).To(BeEmpty())
			})

			It("should return nameMapping when appendedVolumeNames is set", func() {
				ctx.appendedVolumeNames = map[string]string{
					"vol1": "vol1-suffix",
					"vol2": "vol2-suffix",
				}
				nameMapping, err := ctx.GetAppendedVolumeNames()
				Expect(err).NotTo(HaveOccurred())
				Expect(nameMapping).To(HaveLen(2))
				Expect(nameMapping).To(HaveKeyWithValue("vol1", "vol1-suffix"))
				Expect(nameMapping).To(HaveKeyWithValue("vol2", "vol2-suffix"))
			})
		})

		Context("SetAppendedVolumeNames", func() {
			It("should set appendedVolumeNames correctly", func() {
				nameMapping := map[string]string{
					"vol1": "vol1-suffix",
					"vol2": "vol2-suffix",
				}
				ctx.SetAppendedVolumeNames(nameMapping)
				Expect(ctx.appendedVolumeNames).To(Equal(nameMapping))
			})
		})

		Context("GetDatasetUsedInContainers", func() {
			It("should return false when datasetUsedInContainers is nil", func() {
				used, err := ctx.GetDatasetUsedInContainers()
				if err != nil {
					Expect(err).To(HaveOccurred())
					Expect(used).To(BeFalse())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(used).To(BeFalse())
				}
			})

			It("should return datasetUsedInContainers value when set", func() {
				trueVal := true
				ctx.datasetUsedInContainers = &trueVal
				used, err := ctx.GetDatasetUsedInContainers()
				Expect(err).NotTo(HaveOccurred())
				Expect(used).To(BeTrue())

				falseVal := false
				ctx.datasetUsedInContainers = &falseVal
				used, err = ctx.GetDatasetUsedInContainers()
				Expect(err).NotTo(HaveOccurred())
				Expect(used).To(BeFalse())
			})
		})

		Context("SetDatasetUsedInContainers", func() {
			It("should set datasetUsedInContainers to true", func() {
				ctx.SetDatasetUsedInContainers(true)
				Expect(ctx.datasetUsedInContainers).NotTo(BeNil())
				Expect(*ctx.datasetUsedInContainers).To(BeTrue())
			})

			It("should set datasetUsedInContainers to false", func() {
				ctx.SetDatasetUsedInContainers(false)
				Expect(ctx.datasetUsedInContainers).NotTo(BeNil())
				Expect(*ctx.datasetUsedInContainers).To(BeFalse())
			})
		})

		Context("GetDatasetUsedInInitContainers", func() {
			It("should return false when datasetUsedInInitContainers is nil", func() {
				used, err := ctx.GetDatasetUsedInInitContainers()
				if err != nil {
					Expect(err).To(HaveOccurred())
					Expect(used).To(BeFalse())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(used).To(BeFalse())
				}
			})

			It("should return datasetUsedInInitContainers value when set", func() {
				trueVal := true
				ctx.datasetUsedInInitContainers = &trueVal
				used, err := ctx.GetDatasetUsedInInitContainers()
				Expect(err).NotTo(HaveOccurred())
				Expect(used).To(BeTrue())
			})
		})

		Context("SetDatasetUsedInInitContainers", func() {
			It("should set datasetUsedInInitContainers correctly", func() {
				ctx.SetDatasetUsedInInitContainers(true)
				Expect(ctx.datasetUsedInInitContainers).NotTo(BeNil())
				Expect(*ctx.datasetUsedInInitContainers).To(BeTrue())
			})
		})

		Context("generateUniqueHostMountPath", func() {
			It("should generate unique host mount path", func() {
				ctx.generateUniqueHostMountPath = "myplatform"
				path := ctx.generateUniqueHostMountPath
				Expect(path).To(Equal("myplatform"))
			})
		})
	})

	Describe("CollectFluidObjectSpecs", func() {
		var mockPod *mockFluidObject

		BeforeEach(func() {
			mockPod = &mockFluidObject{
				volumes: []corev1.Volume{
					{Name: "vol1"},
				},
				volumeMounts: []corev1.VolumeMount{
					{Name: "vol1", MountPath: "/data"},
				},
				containers: []corev1.Container{
					{Name: "container1"},
				},
				initContainers: []corev1.Container{},
				metaObj: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			}
		})

		It("should collect pod specs successfully", func() {
			podSpecs, err := CollectFluidObjectSpecs(mockPod)
			Expect(err).NotTo(HaveOccurred())
			Expect(podSpecs).NotTo(BeNil())
			Expect(podSpecs.Volumes).To(HaveLen(1))
			Expect(podSpecs.Containers).To(HaveLen(1))
			Expect(podSpecs.MetaObj.Name).To(Equal("test-pod"))
		})

		It("should handle error when getting volumes fails", func() {
			mockPod.shouldErrorOnGetVolumes = true
			podSpecs, err := CollectFluidObjectSpecs(mockPod)
			Expect(err).To(HaveOccurred())
			Expect(podSpecs).To(BeNil())
		})

		It("should handle error when getting volume mounts fails", func() {
			mockPod.shouldErrorOnGetVolumeMounts = true
			podSpecs, err := CollectFluidObjectSpecs(mockPod)
			Expect(err).To(HaveOccurred())
			Expect(podSpecs).To(BeNil())
		})

		It("should handle error when getting containers fails", func() {
			mockPod.shouldErrorOnGetContainers = true
			podSpecs, err := CollectFluidObjectSpecs(mockPod)
			Expect(err).To(HaveOccurred())
			Expect(podSpecs).To(BeNil())
		})

		It("should handle error when getting init containers fails", func() {
			mockPod.shouldErrorOnGetInitContainers = true
			podSpecs, err := CollectFluidObjectSpecs(mockPod)
			Expect(err).To(HaveOccurred())
			Expect(podSpecs).To(BeNil())
		})

		It("should handle error when getting meta object fails", func() {
			mockPod.shouldErrorOnGetMetaObject = true
			podSpecs, err := CollectFluidObjectSpecs(mockPod)
			Expect(err).To(HaveOccurred())
			Expect(podSpecs).To(BeNil())
		})
	})

	Describe("ApplyFluidObjectSpecs", func() {
		var mockPod *mockFluidObject
		var mutatedSpecs *MutatingPodSpecs

		BeforeEach(func() {
			mockPod = &mockFluidObject{
				volumes:        []corev1.Volume{{Name: "vol1"}},
				volumeMounts:   []corev1.VolumeMount{{Name: "vol1", MountPath: "/data"}},
				containers:     []corev1.Container{{Name: "container1"}},
				initContainers: []corev1.Container{},
				metaObj: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
			}

			mutatedSpecs = &MutatingPodSpecs{
				Volumes: []corev1.Volume{
					{Name: "vol1"},
					{Name: "vol2"},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "vol1", MountPath: "/data"},
					{Name: "vol2", MountPath: "/data2"},
				},
				Containers: []corev1.Container{
					{Name: "container1"},
					{Name: "container2"},
				},
				InitContainers: []corev1.Container{
					{Name: "init-container1"},
				},
				MetaObj: metav1.ObjectMeta{
					Name:      "test-pod-mutated",
					Namespace: "default",
					Labels: map[string]string{
						"mutated": "true",
					},
				},
			}
		})

		It("should apply mutated specs successfully", func() {
			err := ApplyFluidObjectSpecs(mockPod, mutatedSpecs)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockPod.setVolumesCalled).To(BeTrue())
			Expect(mockPod.setContainersCalled).To(BeTrue())
			Expect(mockPod.setInitContainersCalled).To(BeTrue())
			Expect(mockPod.setMetaObjectCalled).To(BeTrue())
		})

		It("should handle error when setting volumes fails", func() {
			mockPod.shouldErrorOnSetVolumes = true
			err := ApplyFluidObjectSpecs(mockPod, mutatedSpecs)
			Expect(err).To(HaveOccurred())
		})

		It("should handle error when setting containers fails", func() {
			mockPod.shouldErrorOnSetContainers = true
			err := ApplyFluidObjectSpecs(mockPod, mutatedSpecs)
			Expect(err).To(HaveOccurred())
		})

		It("should handle error when setting init containers fails", func() {
			mockPod.shouldErrorOnSetInitContainers = true
			err := ApplyFluidObjectSpecs(mockPod, mutatedSpecs)
			Expect(err).To(HaveOccurred())
		})

		It("should handle error when setting meta object fails", func() {
			mockPod.shouldErrorOnSetMetaObject = true
			err := ApplyFluidObjectSpecs(mockPod, mutatedSpecs)
			Expect(err).To(HaveOccurred())
		})
	})
})

// Mock implementation of common.FluidObject for testing
type mockFluidObject struct {
	volumes                        []corev1.Volume
	volumeMounts                   []corev1.VolumeMount
	containers                     []corev1.Container
	initContainers                 []corev1.Container
	metaObj                        metav1.ObjectMeta
	shouldErrorOnGetVolumes        bool
	shouldErrorOnGetVolumeMounts   bool
	shouldErrorOnGetContainers     bool
	shouldErrorOnGetInitContainers bool
	shouldErrorOnGetMetaObject     bool
	shouldErrorOnSetVolumes        bool
	shouldErrorOnSetContainers     bool
	shouldErrorOnSetInitContainers bool
	shouldErrorOnSetMetaObject     bool
	setVolumesCalled               bool
	setContainersCalled            bool
	setInitContainersCalled        bool
	setMetaObjectCalled            bool
}

func (m *mockFluidObject) GetRoot() runtime.Object {
	// Return a mock pod as the root object
	return &corev1.Pod{
		ObjectMeta: m.metaObj,
		Spec: corev1.PodSpec{
			Containers:     m.containers,
			InitContainers: m.initContainers,
			Volumes:        m.volumes,
		},
	}
}

func (m *mockFluidObject) GetVolumes() ([]corev1.Volume, error) {
	if m.shouldErrorOnGetVolumes {
		return nil, fmt.Errorf("mock error getting volumes")
	}
	return m.volumes, nil
}

func (m *mockFluidObject) GetVolumeMounts() ([]corev1.VolumeMount, error) {
	if m.shouldErrorOnGetVolumeMounts {
		return nil, fmt.Errorf("mock error getting volume mounts")
	}
	return m.volumeMounts, nil
}

func (m *mockFluidObject) GetContainers() ([]corev1.Container, error) {
	if m.shouldErrorOnGetContainers {
		return nil, fmt.Errorf("mock error getting containers")
	}
	return m.containers, nil
}

func (m *mockFluidObject) GetInitContainers() ([]corev1.Container, error) {
	if m.shouldErrorOnGetInitContainers {
		return nil, fmt.Errorf("mock error getting init containers")
	}
	return m.initContainers, nil
}

func (m *mockFluidObject) GetMetaObject() (metav1.ObjectMeta, error) {
	if m.shouldErrorOnGetMetaObject {
		return metav1.ObjectMeta{}, fmt.Errorf("mock error getting meta object")
	}
	return m.metaObj, nil
}

func (m *mockFluidObject) SetVolumes(volumes []corev1.Volume) error {
	if m.shouldErrorOnSetVolumes {
		return fmt.Errorf("mock error setting volumes")
	}
	m.setVolumesCalled = true
	m.volumes = volumes
	return nil
}

func (m *mockFluidObject) SetContainers(containers []corev1.Container) error {
	if m.shouldErrorOnSetContainers {
		return fmt.Errorf("mock error setting containers")
	}
	m.setContainersCalled = true
	m.containers = containers
	return nil
}

func (m *mockFluidObject) SetInitContainers(containers []corev1.Container) error {
	if m.shouldErrorOnSetInitContainers {
		return fmt.Errorf("mock error setting init containers")
	}
	m.setInitContainersCalled = true
	m.initContainers = containers
	return nil
}

func (m *mockFluidObject) SetMetaObject(metaObj metav1.ObjectMeta) error {
	if m.shouldErrorOnSetMetaObject {
		return fmt.Errorf("mock error setting meta object")
	}
	m.setMetaObjectCalled = true
	m.metaObj = metaObj
	return nil
}