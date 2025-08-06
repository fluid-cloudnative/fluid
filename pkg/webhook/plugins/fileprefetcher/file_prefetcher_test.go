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
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInjectFilePrefetcherSidecar_WithMatchingContainers_ShouldInsertAfterLastMatch(t *testing.T) {
	oldContainers := []corev1.Container{
		{Name: common.FuseContainerName + "-2"},
		{Name: common.FuseContainerName + "-1"},
		{Name: "other-container"},
		{Name: "another-container"},
	}
	prefetcher := &FilePrefetcher{}
	filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

	newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

	assert.Equal(t, 5, len(newContainers))
	assert.Equal(t, common.FuseContainerName+"-2", newContainers[0].Name)
	assert.Equal(t, common.FuseContainerName+"-1", newContainers[1].Name)
	assert.Equal(t, "file-prefetcher-ctr", newContainers[2].Name)
	assert.Equal(t, "other-container", newContainers[3].Name)
	assert.Equal(t, "another-container", newContainers[4].Name)
}

func TestInjectFilePrefetcherSidecar_WithoutMatchingContainers_ShouldInsertAtBeginning(t *testing.T) {
	oldContainers := []corev1.Container{
		{Name: "other-container"},
		{Name: "another-container"},
	}
	prefetcher := &FilePrefetcher{}
	filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

	newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

	assert.Equal(t, 3, len(newContainers))
	assert.Equal(t, "file-prefetcher-ctr", newContainers[0].Name)
	assert.Equal(t, "other-container", newContainers[1].Name)
	assert.Equal(t, "another-container", newContainers[2].Name)
}

func TestInjectFilePrefetcherSidecar_EmptyContainerList_ShouldOnlyContainFilePrefetcher(t *testing.T) {
	oldContainers := []corev1.Container{}
	filePrefetcherCtr := corev1.Container{Name: "file-prefetcher-ctr"}

	prefetcher := &FilePrefetcher{}
	newContainers := prefetcher.injectFilePrefetcherSidecar(oldContainers, filePrefetcherCtr)

	assert.Equal(t, 1, len(newContainers))
	assert.Equal(t, "file-prefetcher-ctr", newContainers[0].Name)
}

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
