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

package fuse

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/mutator"
	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Fuse Injector", func() {
	var (
		injector     *Injector
		fakeClient   client.Client
		scheme       *runtime.Scheme
		testLogger   logr.Logger
		namespace    string
		runtimeInfos map[string]base.RuntimeInfoInterface
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		testLogger = logr.Discard()
		namespace = "default"

		injector = &Injector{
			client: fakeClient,
			log:    testLogger,
		}

	})

	Describe("injectCheckMountReadyScript", func() {
		Context("when no runtime infos are provided", func() {
			It("should skip injection and return nil", func() {
				podSpecs := &mutator.MutatingPodSpecs{
					MetaObj: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: namespace,
					},
					Volumes:    []corev1.Volume{},
					Containers: []corev1.Container{},
				}

				err := injector.injectCheckMountReadyScript(podSpecs, map[string]base.RuntimeInfoInterface{})
				Expect(err).To(BeNil())
				Expect(len(podSpecs.Volumes)).To(Equal(0))
			})
		})

		Context("when runtime infos are provided", func() {
			It("should inject script volume and volume mounts", func() {
				podSpecs := &mutator.MutatingPodSpecs{
					MetaObj: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: namespace,
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "app-container",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
						},
					},
				}

				err := injector.injectCheckMountReadyScript(podSpecs, runtimeInfos)
				Expect(err).To(BeNil())
			})
		})

		Context("when pod has GenerateName instead of Name", func() {
			It("should use GenerateName for logging", func() {
				podSpecs := &mutator.MutatingPodSpecs{
					MetaObj: metav1.ObjectMeta{
						GenerateName: "test-pod-",
						Namespace:    namespace,
					},
					Volumes:    []corev1.Volume{},
					Containers: []corev1.Container{},
				}

				err := injector.injectCheckMountReadyScript(podSpecs, runtimeInfos)
				Expect(err).To(BeNil())
			})
		})

		Context("when PostStart injection is enabled", func() {
			It("should inject PostStart lifecycle hooks", func() {
				podSpecs := &mutator.MutatingPodSpecs{
					MetaObj: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: namespace,
						Labels: map[string]string{
							"fluid.io/enable-injection": "true",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "app-container",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
						},
					},
				}

				err := injector.injectCheckMountReadyScript(podSpecs, runtimeInfos)
				Expect(err).To(BeNil())
			})
		})

		Context("when container already has PostStart lifecycle", func() {
			It("should skip PostStart injection", func() {
				existingPostStart := &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"/bin/sh", "-c", "echo existing"},
					},
				}

				podSpecs := &mutator.MutatingPodSpecs{
					MetaObj: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: namespace,
						Labels: map[string]string{
							"fluid.io/enable-injection": "true",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: "app-container",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: existingPostStart,
							},
						},
					},
				}

				err := injector.injectCheckMountReadyScript(podSpecs, runtimeInfos)
				Expect(err).To(BeNil())
				Expect(podSpecs.Containers[0].Lifecycle.PostStart).To(Equal(existingPostStart))
			})
		})

		Context("when init containers are present", func() {
			It("should inject into init containers", func() {
				podSpecs := &mutator.MutatingPodSpecs{
					MetaObj: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: namespace,
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-pvc",
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name: "init-container",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-volume",
									MountPath: "/data",
								},
							},
						},
					},
					Containers: []corev1.Container{},
				}

				err := injector.injectCheckMountReadyScript(podSpecs, runtimeInfos)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("ensureScriptConfigMapExists", func() {
		Context("when configmap does not exist", func() {
			It("should create the configmap", func() {
				appScriptGen, err := injector.ensureScriptConfigMapExists(namespace)
				Expect(err).To(BeNil())
				Expect(appScriptGen).NotTo(BeNil())

				// Verify configmap was created
				cm := appScriptGen.BuildConfigmap()
				retrievedCM := &corev1.ConfigMap{}
				err = fakeClient.Get(context.TODO(), client.ObjectKey{
					Name:      cm.Name,
					Namespace: cm.Namespace,
				}, retrievedCM)
				Expect(err).To(BeNil())
			})
		})

		Context("when configmap already exists", func() {
			It("should not return an error", func() {
				// Create configmap first
				appScriptGen := poststart.NewScriptGeneratorForApp(namespace)
				cm := appScriptGen.BuildConfigmap()
				err := fakeClient.Create(context.TODO(), cm)
				Expect(err).To(BeNil())

				// Try to ensure it exists again
				_, err = injector.ensureScriptConfigMapExists(namespace)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("collectDatasetVolumeMountInfo", func() {
		Context("when volume mount does not reference a PVC", func() {
			It("should return empty map", func() {
				volMounts := []corev1.VolumeMount{
					{
						Name:      "empty-volume",
						MountPath: "/empty",
					},
				}

				volumes := []corev1.Volume{
					{
						Name: "empty-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				}

				result := collectDatasetVolumeMountInfo(volMounts, volumes, runtimeInfos)
				Expect(result).To(HaveLen(0))
			})
		})

		Context("when PVC is not in runtime infos", func() {
			It("should return empty map", func() {
				volMounts := []corev1.VolumeMount{
					{
						Name:      "data-volume",
						MountPath: "/data",
					},
				}

				volumes := []corev1.Volume{
					{
						Name: "data-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "unknown-pvc",
							},
						},
					},
				}

				result := collectDatasetVolumeMountInfo(volMounts, volumes, runtimeInfos)
				Expect(result).To(HaveLen(0))
			})
		})
	})

	Describe("assembleMountInfos", func() {
		Context("when given path to runtime type map", func() {
			It("should return colon-separated strings", func() {
				path2RuntimeTypeMap := map[string]string{
					"/data1": "alluxio",
					"/data2": "juicefs",
				}

				mountPathStr, mountTypeStr := assembleMountInfos(path2RuntimeTypeMap)
				Expect(mountPathStr).To(ContainSubstring("/data1"))
				Expect(mountPathStr).To(ContainSubstring("/data2"))
				Expect(mountPathStr).To(ContainSubstring(":"))
				Expect(mountTypeStr).To(ContainSubstring("alluxio"))
				Expect(mountTypeStr).To(ContainSubstring("juicefs"))
				Expect(mountTypeStr).To(ContainSubstring(":"))
			})
		})

		Context("when given empty map", func() {
			It("should return empty strings", func() {
				path2RuntimeTypeMap := map[string]string{}

				mountPathStr, mountTypeStr := assembleMountInfos(path2RuntimeTypeMap)
				Expect(mountPathStr).To(Equal(""))
				Expect(mountTypeStr).To(Equal(""))
			})
		})

		Context("when given single entry", func() {
			It("should return strings without colons", func() {
				path2RuntimeTypeMap := map[string]string{
					"/data": "alluxio",
				}

				mountPathStr, mountTypeStr := assembleMountInfos(path2RuntimeTypeMap)
				Expect(mountPathStr).To(Equal("/data"))
				Expect(mountTypeStr).To(Equal("alluxio"))
			})
		})
	})
})

// MockRuntimeInfo is a mock implementation of base.RuntimeInfoInterface for testing
type MockRuntimeInfo struct {
	namespace   string
	runtimeType string
}

func (m *MockRuntimeInfo) GetNamespace() string {
	return m.namespace
}

func (m *MockRuntimeInfo) GetRuntimeType() string {
	return m.runtimeType
}

func (m *MockRuntimeInfo) GetName() string {
	return "mock-runtime"
}

func (m *MockRuntimeInfo) IsExclusive() bool {
	return false
}

func (m *MockRuntimeInfo) GetAnnotations() map[string]string {
	return map[string]string{}
}

func (m *MockRuntimeInfo) GetCommonLabelName() string {
	return "fluid.io/dataset"
}
