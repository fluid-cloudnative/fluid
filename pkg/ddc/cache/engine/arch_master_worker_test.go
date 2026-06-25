/*
Copyright 2026 The Fluid Authors.

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

package engine

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("masterWorkerArchApi Tests", Label("pkg.ddc.cache.engine.arch_master_worker_test.go"), func() {
	It("GetExecutionPodInfo should return correct master pod info", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master-main"}},
						},
					},
				},
			},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}

		podName, containerName, err := api.GetExecutionPodInfo()
		Expect(err).NotTo(HaveOccurred())
		Expect(podName).To(Equal("my-runtime-master-0"))
		Expect(containerName).To(Equal("master-main"))
	})

	It("GetExecutionPodInfo should return error when master template has no containers", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{},
				},
			},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}

		_, _, err := api.GetExecutionPodInfo()
		Expect(err).To(HaveOccurred())
	})

	It("GetExecutionPodInfo should return error when topology is nil", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}

		_, _, err := api.GetExecutionPodInfo()
		Expect(err).To(HaveOccurred())
	})

	It("GetExecutionEntries should return master execution entries", func() {
		entries := &datav1alpha1.ExecutionEntries{
			MountUFS: &datav1alpha1.ExecutionCommonEntry{Command: []string{"/mount.sh"}},
		}
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{ExecutionEntries: entries},
			},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}

		got := api.GetExecutionEntries()
		Expect(got).To(Equal(entries))
	})

	It("GetExecutionEntries should return nil when topology is nil", func() {
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: &datav1alpha1.CacheRuntimeClass{}}
		Expect(api.GetExecutionEntries()).To(BeNil())
	})

	It("GetExecutionEntries should return nil when master topology is nil", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}
		Expect(api.GetExecutionEntries()).To(BeNil())
	})

	It("IsMountUFSSupported should return true when MountUFS entry is configured", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master"}},
						},
					},
					ExecutionEntries: &datav1alpha1.ExecutionEntries{
						MountUFS: &datav1alpha1.ExecutionCommonEntry{Command: []string{"/mount.sh"}},
					},
				},
			},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}
		Expect(api.IsMountUFSSupported()).To(BeTrue())
	})

	It("IsMountUFSSupported should return false when MountUFS entry is nil", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master"}},
						},
					},
					ExecutionEntries: &datav1alpha1.ExecutionEntries{},
				},
			},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}
		Expect(api.IsMountUFSSupported()).To(BeFalse())
	})

	It("IsMountUFSSupported should return false when ExecutionEntries is nil", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master"}},
						},
					},
				},
			},
		}
		api := &masterWorkerArchApi{name: "my-runtime", namespace: "default", runtimeClass: runtimeClass}
		Expect(api.IsMountUFSSupported()).To(BeFalse())
	})
})
