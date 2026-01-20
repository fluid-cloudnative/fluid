/*
Copyright 2025 The Fluid Authors.

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

package fileprefetcher

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("injectFilePrefetcherSidecar", func() {
	var prefetcher *FilePrefetcher

	BeforeEach(func() {
		prefetcher = &FilePrefetcher{}
	})

	Context("with matching containers", func() {
		It("should insert after last match", func() {
			oldContainers := []corev1.Container{
				{Name: common.FuseContainerName + "-2"},
				{Name: common.FuseContainerName + "-1"},
				{Name: "other-container"},
				{Name: "another-container"},
			}
			filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

			newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

			Expect(newContainers).To(HaveLen(5))
			Expect(newContainers[0].Name).To(Equal(common.FuseContainerName + "-2"))
			Expect(newContainers[1].Name).To(Equal(common.FuseContainerName + "-1"))
			Expect(newContainers[2].Name).To(Equal("file-prefetcher-ctr"))
			Expect(newContainers[3].Name).To(Equal("other-container"))
			Expect(newContainers[4].Name).To(Equal("another-container"))
		})
	})

	Context("without matching containers", func() {
		It("should insert at beginning", func() {
			oldContainers := []corev1.Container{
				{Name: "other-container"},
				{Name: "another-container"},
			}
			filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

			newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

			Expect(newContainers).To(HaveLen(3))
			Expect(newContainers[0].Name).To(Equal("file-prefetcher-ctr"))
			Expect(newContainers[1].Name).To(Equal("other-container"))
			Expect(newContainers[2].Name).To(Equal("another-container"))
		})
	})

	Context("with empty container list", func() {
		It("should only contain file prefetcher", func() {
			oldContainers := []corev1.Container{}
			filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

			newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

			Expect(newContainers).To(HaveLen(1))
			Expect(newContainers[0].Name).To(Equal("file-prefetcher-ctr"))
		})
	})
})

var _ = Describe("buildFilePrefetcherConfig", func() {
	var (
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
		prefetcher   *FilePrefetcher
	)

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
		}
		runtimeInfos = map[string]base.RuntimeInfoInterface{}
		prefetcher = &FilePrefetcher{}
	})

	Context("when all annotations are provided correctly", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherImage] = "test-image"
			pod.Annotations[AnnotationFilePrefetcherExtraEnvs] = "KEY1=VALUE1 KEY2=VALUE2"
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "true"
			pod.Annotations[AnnotationFilePrefetcherTimeoutSeconds] = "30"
			pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{Name: "myvol", VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "mypvc",
				},
			}})
			runtimeInfos["mypvc"] = &base.RuntimeInfo{}
		})

		It("should build the configuration successfully", func() {
			config, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Image).To(Equal("test-image"))
			Expect(config.ExtraEnvs).To(HaveKeyWithValue("KEY1", "VALUE1"))
			Expect(config.ExtraEnvs).To(HaveKeyWithValue("KEY2", "VALUE2"))
			Expect(config.GlobPaths).To(Equal("/data/myvol/path/to/myfolder/*.pkl"))
			Expect(config.AsyncPrefetch).To(BeTrue())
			Expect(config.TimeoutSeconds).To(Equal(30))
		})

	})

	Context("when the image annotation is missing, with default image set", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherExtraEnvs] = "KEY1=VALUE1 KEY2=VALUE2"
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "true"
			pod.Annotations[AnnotationFilePrefetcherTimeoutSeconds] = "30"
			pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{Name: "myvol", VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: "mypvc",
				},
			}})
			runtimeInfos["mypvc"] = &base.RuntimeInfo{}
			defaultFilePrefetcherImage = "test-image"
		})

		It("should use the default image", func() {
			config, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.Image).To(Equal(defaultFilePrefetcherImage))
			Expect(config.ExtraEnvs).To(HaveKeyWithValue("KEY1", "VALUE1"))
			Expect(config.ExtraEnvs).To(HaveKeyWithValue("KEY2", "VALUE2"))
			Expect(config.GlobPaths).To(Equal("/data/myvol/path/to/myfolder/*.pkl"))
			Expect(config.AsyncPrefetch).To(BeTrue())
			Expect(config.TimeoutSeconds).To(Equal(30))
		})
	})

	Context("when the image annotation is missing, no default image set", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherExtraEnvs] = "KEY1=VALUE1 KEY2=VALUE2"
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "true"
			pod.Annotations[AnnotationFilePrefetcherTimeoutSeconds] = "30"
			defaultFilePrefetcherImage = ""
		})

		It("should return an error", func() {
			_, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the extra envs annotation has an invalid format", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherImage] = "test-image"
			pod.Annotations[AnnotationFilePrefetcherExtraEnvs] = "KEY1=VALUE1 KEY2"
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "true"
			pod.Annotations[AnnotationFilePrefetcherTimeoutSeconds] = "30"
		})

		It("should return an error", func() {
			_, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the async prefetch annotation has an invalid value", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherImage] = "test-image"
			pod.Annotations[AnnotationFilePrefetcherExtraEnvs] = "KEY1=VALUE1 KEY2=VALUE2"
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "invalid"
			pod.Annotations[AnnotationFilePrefetcherTimeoutSeconds] = "30"
		})

		It("should return an error", func() {
			_, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the timeout seconds annotation has an invalid value", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherImage] = "test-image"
			pod.Annotations[AnnotationFilePrefetcherExtraEnvs] = "KEY1=VALUE1 KEY2=VALUE2"
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "true"
			pod.Annotations[AnnotationFilePrefetcherTimeoutSeconds] = "invalid"
		})

		It("should return an error", func() {
			_, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("NewPlugin", func() {
	It("should create a new FilePrefetcher plugin", func() {
		plugin, err := NewPlugin(nil, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(plugin).NotTo(BeNil())
		Expect(plugin.GetName()).To(Equal(Name))
	})
})

var _ = Describe("Mutate", func() {
	var (
		prefetcher   *FilePrefetcher
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
	)

	BeforeEach(func() {
		prefetcher = &FilePrefetcher{
			name: Name,
			log:  ctrl.Log.WithName("FilePrefetcher"),
		}
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "app-container"},
				},
			},
		}
		runtimeInfos = map[string]base.RuntimeInfoInterface{}
		defaultFilePrefetcherImage = "test-default-image"
	})

	Context("when injection annotation is not set", func() {
		It("should skip mutation", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
			Expect(pod.Spec.Containers).To(HaveLen(1))
		})
	})

	Context("when injection annotation is false", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = "false"
		})

		It("should skip mutation", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
		})
	})

	Context("when injection is already done", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
			pod.Annotations[AnnotationFilePrefetcherInjectDone] = common.True
		})

		It("should skip mutation", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
			Expect(pod.Spec.Containers).To(HaveLen(1))
		})
	})

	Context("when config building fails", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
			defaultFilePrefetcherImage = ""
		})

		It("should return error and stop", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(shouldStop).To(BeTrue())
		})
	})

	Context("when file list is empty or invalid", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
			pod.Annotations[AnnotationFilePrefetcherFileList] = "invalid://no-pvc"
			defaultFilePrefetcherImage = "test-image"
		})

		It("should skip injection but not error", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
			Expect(pod.Spec.Containers).To(HaveLen(1))
		})
	})

	Context("when successfully injecting with sync prefetch", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/data/*.pkl"
			pod.Spec.Volumes = []corev1.Volume{
				{
					Name: "mypvc-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "mypvc",
						},
					},
				},
			}
			pod.Spec.Containers = []corev1.Container{
				{Name: common.FuseContainerName + "-1"},
				{Name: "app-container"},
			}
			runtimeInfos["mypvc"] = &base.RuntimeInfo{}
		})

		It("should inject prefetcher container and mark as done", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())
			Expect(pod.Spec.Containers).To(HaveLen(3))
			Expect(pod.Spec.Containers[1].Name).To(Equal(filePrefetcherContainerName))
			Expect(pod.Annotations[AnnotationFilePrefetcherInjectDone]).To(Equal(common.True))
			Expect(pod.Spec.Volumes).To(HaveLen(2))
		})
	})

	Context("when successfully injecting with async prefetch", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
			pod.Annotations[AnnotationFilePrefetcherFileList] = "pvc://mypvc/data/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherAsync] = "true"
			pod.Spec.Volumes = []corev1.Volume{
				{
					Name: "mypvc-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "mypvc",
						},
					},
				},
			}
			pod.Spec.Containers = []corev1.Container{
				{Name: common.FuseContainerName + "-1"},
				{Name: "app-container"},
			}
			runtimeInfos["mypvc"] = &base.RuntimeInfo{}
		})

		It("should inject status volume into non-fuse containers", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			// App container should have status volume mount
			appContainer := pod.Spec.Containers[2]
			Expect(appContainer.Name).To(Equal("app-container"))
			Expect(appContainer.VolumeMounts).To(ContainElement(
				corev1.VolumeMount{
					Name:      filePrefetcherStatusVolumeName,
					MountPath: filePrefetcherStatusVolumeMountPath,
				},
			))

			// FUSE container should not have status volume mount
			fuseContainer := pod.Spec.Containers[0]
			Expect(fuseContainer.VolumeMounts).NotTo(ContainElement(
				corev1.VolumeMount{
					Name:      filePrefetcherStatusVolumeName,
					MountPath: filePrefetcherStatusVolumeMountPath,
				},
			))
		})
	})

	Context("when using default file list with multiple PVCs", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
			pod.Spec.Volumes = []corev1.Volume{
				{
					Name: "pvc1-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc1",
						},
					},
				},
				{
					Name: "pvc2-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc2",
						},
					},
				},
			}
			runtimeInfos["pvc1"] = &base.RuntimeInfo{}
			runtimeInfos["pvc2"] = &base.RuntimeInfo{}
		})

		It("should include all PVCs in file list", func() {
			shouldStop, err := prefetcher.Mutate(pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(shouldStop).To(BeFalse())

			prefetcherContainer := pod.Spec.Containers[0]
			var fileListEnv *corev1.EnvVar
			for _, env := range prefetcherContainer.Env {
				if env.Name == envKeyFilePrefetcherFileList {
					fileListEnv = &env
					break
				}
			}
			Expect(fileListEnv).NotTo(BeNil())
			Expect(fileListEnv.Value).To(ContainSubstring("/data/pvc1-vol/**"))
			Expect(fileListEnv.Value).To(ContainSubstring("/data/pvc2-vol/**"))
		})
	})
})

var _ = Describe("injectFilePrefetcherSidecar edge cases", func() {
	var prefetcher *FilePrefetcher

	BeforeEach(func() {
		prefetcher = &FilePrefetcher{}
	})

	Context("with single FUSE container at the end", func() {
		It("should insert after the FUSE container", func() {
			oldContainers := []corev1.Container{
				{Name: "app-container"},
				{Name: common.FuseContainerName + "-1"},
			}
			filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

			newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

			Expect(newContainers).To(HaveLen(3))
			Expect(newContainers[0].Name).To(Equal("app-container"))
			Expect(newContainers[1].Name).To(Equal(common.FuseContainerName + "-1"))
			Expect(newContainers[2].Name).To(Equal("file-prefetcher-ctr"))
		})
	})

	Context("with multiple non-consecutive FUSE containers", func() {
		It("should insert after the last FUSE container", func() {
			oldContainers := []corev1.Container{
				{Name: common.FuseContainerName + "-1"},
				{Name: "app-container-1"},
				{Name: common.FuseContainerName + "-2"},
				{Name: "app-container-2"},
			}
			filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

			newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

			Expect(newContainers).To(HaveLen(5))
			Expect(newContainers[0].Name).To(Equal(common.FuseContainerName + "-1"))
			Expect(newContainers[1].Name).To(Equal("app-container-1"))
			Expect(newContainers[2].Name).To(Equal(common.FuseContainerName + "-2"))
			Expect(newContainers[3].Name).To(Equal("file-prefetcher-ctr"))
			Expect(newContainers[4].Name).To(Equal("app-container-2"))
		})
	})
})

var _ = Describe("buildFilePrefetcherSidecarContainer with multiple volumes", func() {
	var prefetcher *FilePrefetcher

	BeforeEach(func() {
		prefetcher = &FilePrefetcher{}
	})

	Context("with multiple volume mounts", func() {
		It("should mount all volumes correctly", func() {
			config := filePrefetcherConfig{
				Image:         "test-image",
				AsyncPrefetch: true,
				VolumeMountPaths: map[string]string{
					"vol1": "/data/vol1",
					"vol2": "/data/vol2",
					"vol3": "/data/vol3",
				},
				GlobPaths:      "/data/vol1/**;/data/vol2/**;/data/vol3/**",
				TimeoutSeconds: 60,
				ExtraEnvs:      map[string]string{},
			}

			containerSpec, _ := prefetcher.buildFilePrefetcherSidecarContainer(config)

			Expect(containerSpec.VolumeMounts).To(HaveLen(4)) // 3 data volumes + 1 status volume

			volumeNames := make([]string, 0)
			for _, vm := range containerSpec.VolumeMounts {
				volumeNames = append(volumeNames, vm.Name)
			}
			Expect(volumeNames).To(ContainElements("vol1", "vol2", "vol3", filePrefetcherStatusVolumeName))
		})
	})
})

var _ = Describe("parseGlobPathsFromFileList additional cases", func() {
	var (
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
		prefetcher   *FilePrefetcher
	)

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: "pvc1-vol",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc1",
							},
						},
					},
					{
						Name: "pvc2-vol",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc2",
							},
						},
					},
				},
			},
		}
		runtimeInfos = map[string]base.RuntimeInfoInterface{
			"pvc1": &base.RuntimeInfo{},
			"pvc2": &base.RuntimeInfo{},
		}
		prefetcher = &FilePrefetcher{}
	})

	Context("with mixed valid and invalid paths", func() {
		It("should only parse valid paths", func() {
			fileList := "pvc://pvc1/data/*.pkl;invalid://bad;pvc://pvc2/logs/*.log;pvc://nonexistent/test"

			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(fileList, pod, runtimeInfos)

			Expect(volumeMountPaths).To(HaveLen(2))
			Expect(globPaths).To(HaveLen(2))
			Expect(globPaths).To(ContainElements(
				"/data/pvc1-vol/data/*.pkl",
				"/data/pvc2-vol/logs/*.log",
			))
		})
	})
})

var _ = Describe("buildFilePrefetcherConfig with defaults", func() {
	var (
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
		prefetcher   *FilePrefetcher
	)

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: "vol",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pvc",
							},
						},
					},
				},
			},
		}
		runtimeInfos = map[string]base.RuntimeInfoInterface{
			"pvc": &base.RuntimeInfo{},
		}
		prefetcher = &FilePrefetcher{}
		defaultFilePrefetcherImage = "default-image"
	})

	Context("when only required annotations are set", func() {
		BeforeEach(func() {
			pod.Annotations[AnnotationFilePrefetcherInject] = common.True
		})

		It("should use default values for optional fields", func() {
			config, err := prefetcher.buildFilePrefetcherConfig(pod, runtimeInfos)

			Expect(err).NotTo(HaveOccurred())
			Expect(config.Image).To(Equal("default-image"))
			Expect(config.AsyncPrefetch).To(BeFalse())
			Expect(config.ExtraEnvs).To(BeEmpty())
			Expect(config.GlobPaths).To(Equal("/data/vol/**"))
		})
	})
})

var _ = Describe("parseGlobPathsFromFileList", func() {
	var (
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
		prefetcher   *FilePrefetcher
	)

	BeforeEach(func() {
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{},
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: "mypvc-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "mypvc",
							},
						},
					},
				},
			},
		}
		runtimeInfos = map[string]base.RuntimeInfoInterface{
			"mypvc": &base.RuntimeInfo{},
		}
		prefetcher = &FilePrefetcher{}
	})

	Context("when the file list is valid", func() {
		BeforeEach(func() {
			fileList := "pvc://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should parse the glob paths correctly", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(HaveKeyWithValue("mypvc-volume", "/data/mypvc-volume"))
			Expect(globPaths).To(ConsistOf("/data/mypvc-volume/path/to/myfolder/*.pkl"))
		})
	})

	Context("when the file list contains multiple valid paths", func() {
		BeforeEach(func() {
			fileList := "pvc://mypvc/path/to/myfolder/*.pkl;pvc://mypvc/another/path/*.txt"
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should parse the glob paths correctly", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(HaveKeyWithValue("mypvc-volume", "/data/mypvc-volume"))
			Expect(globPaths).To(ConsistOf("/data/mypvc-volume/path/to/myfolder/*.pkl", "/data/mypvc-volume/another/path/*.txt"))
		})
	})

	Context("when the file list contains an invalid URI path", func() {
		BeforeEach(func() {
			fileList := "invalid://mypvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should skip the invalid path", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(BeEmpty())
			Expect(globPaths).To(BeEmpty())
		})
	})

	Context("when the file list contains a path without a valid PVC", func() {
		BeforeEach(func() {
			fileList := "pvc://nonexistent-pvc/path/to/myfolder/*.pkl"
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should skip the path with the nonexistent PVC", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(BeEmpty())
			Expect(globPaths).To(BeEmpty())
		})
	})

	Context("when the file list is empty", func() {
		BeforeEach(func() {
			fileList := ""
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should return empty volume mount paths and glob paths", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(BeEmpty())
			Expect(globPaths).To(BeEmpty())
		})
	})

	Context("when the file list contains a path without a subpath", func() {
		BeforeEach(func() {
			fileList := "pvc://mypvc"
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should default the glob path to **", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(HaveKeyWithValue("mypvc-volume", "/data/mypvc-volume"))
			Expect(globPaths).To(ConsistOf("/data/mypvc-volume/**"))
		})
	})

	Context("when the file list contains relative path like \"..\" and \".\"", func() {
		BeforeEach(func() {
			fileList := "pvc://mypvc/../path/to/myfolder/.././mydir*/test1"
			pod.Annotations[AnnotationFilePrefetcherFileList] = fileList
		})

		It("should clean the file list", func() {
			volumeMountPaths, globPaths := prefetcher.parseGlobPathsFromFileList(pod.Annotations[AnnotationFilePrefetcherFileList], pod, runtimeInfos)
			Expect(volumeMountPaths).To(HaveKeyWithValue("mypvc-volume", "/data/mypvc-volume"))
			Expect(globPaths).To(ConsistOf("/data/mypvc-volume/path/to/mydir*/test1"))
		})
	})
})

var _ = Describe("buildFilePrefetcherSidecarContainer", func() {
	var (
		config     filePrefetcherConfig
		prefetcher *FilePrefetcher
	)

	BeforeEach(func() {
		prefetcher = &FilePrefetcher{}
		config = filePrefetcherConfig{
			Image:            "test-image",
			AsyncPrefetch:    false,
			VolumeMountPaths: map[string]string{"mypvc-volume": "/data/mypvc-volume"},
			GlobPaths:        "/data/mypvc-volume/path/to/myfolder/*.pkl",
			TimeoutSeconds:   30,
			ExtraEnvs:        map[string]string{"KEY1": "VALUE1", "KEY2": "VALUE2"},
		}
	})

	Context("when the configuration is valid", func() {
		It("should build the sidecar container correctly", func() {
			containerSpec, statusFileVolume := prefetcher.buildFilePrefetcherSidecarContainer(config)
			Expect(containerSpec.Name).To(Equal(filePrefetcherContainerName))
			Expect(containerSpec.Image).To(Equal("test-image"))
			Expect(containerSpec.Env).To(ContainElements(
				corev1.EnvVar{Name: envKeyFilePrefetcherFileList, Value: "/data/mypvc-volume/path/to/myfolder/*.pkl"},
				corev1.EnvVar{Name: envKeyFilePrefetcherAsyncPrefetch, Value: "false"},
				corev1.EnvVar{Name: envKeyFilePrefetcherTimeoutSeconds, Value: "30"},
				corev1.EnvVar{Name: "KEY1", Value: "VALUE1"},
				corev1.EnvVar{Name: "KEY2", Value: "VALUE2"},
			))
			Expect(containerSpec.VolumeMounts).To(ContainElements(
				corev1.VolumeMount{Name: "mypvc-volume", MountPath: "/data/mypvc-volume"},
				corev1.VolumeMount{Name: filePrefetcherStatusVolumeName, MountPath: filePrefetcherStatusVolumeMountPath},
			))
			Expect(statusFileVolume.Name).To(Equal(filePrefetcherStatusVolumeName))
			Expect(statusFileVolume.VolumeSource.EmptyDir).NotTo(BeNil())
		})
	})

	Context("when async prefetch is enabled", func() {
		BeforeEach(func() {
			config.AsyncPrefetch = true
		})

		It("should not have a PostStart lifecycle hook", func() {
			containerSpec, _ := prefetcher.buildFilePrefetcherSidecarContainer(config)
			Expect(containerSpec.Lifecycle).To(BeNil())
		})
	})

	Context("when async prefetch is disabled", func() {
		BeforeEach(func() {
			config.AsyncPrefetch = false
		})

		It("should have a PostStart lifecycle hook", func() {
			containerSpec, _ := prefetcher.buildFilePrefetcherSidecarContainer(config)
			Expect(containerSpec.Lifecycle).NotTo(BeNil())
			Expect(containerSpec.Lifecycle.PostStart.Exec.Command).To(Equal([]string{
				"bash",
				"-c",
				`cnt=0; while [[ $cnt -lt $FILE_PREFETCHER_TIMEOUT_SECONDS ]]; do if [[ -e "/tmp/fluid-file-prefetcher/status/prefetcher.status" ]]; then exit 0; fi; cnt=$(expr $cnt + 1); sleep 1; done; echo "time out waiting for prefetching done"; exit 1`,
			}))
		})
	})

	Context("when there are no extra environment variables", func() {
		BeforeEach(func() {
			config.ExtraEnvs = map[string]string{}
		})

		It("should not include extra environment variables", func() {
			containerSpec, _ := prefetcher.buildFilePrefetcherSidecarContainer(config)
			Expect(containerSpec.Env).To(ContainElements(
				corev1.EnvVar{Name: envKeyFilePrefetcherFileList, Value: "/data/mypvc-volume/path/to/myfolder/*.pkl"},
				corev1.EnvVar{Name: envKeyFilePrefetcherAsyncPrefetch, Value: "false"},
				corev1.EnvVar{Name: envKeyFilePrefetcherTimeoutSeconds, Value: "30"},
			))
			Expect(containerSpec.Env).NotTo(ContainElement(corev1.EnvVar{Name: "KEY1", Value: "VALUE1"}))
			Expect(containerSpec.Env).NotTo(ContainElement(corev1.EnvVar{Name: "KEY2", Value: "VALUE2"}))
		})
	})
})

func TestFilePrefetcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FilePrefetcher")
}
