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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("resolveArchitectureApi Tests", Label("pkg.ddc.cache.engine.architecture_test.go"), func() {
	It("should return masterWorkerArchApi when master topology is configured and master is enabled", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-runtime-class",
			},
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
		rt := &datav1alpha1.CacheRuntime{
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-runtime-class",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
			},
		}

		archApi := resolveArchitectureApi("test-runtime", "default", rt, runtimeClass)
		Expect(archApi).NotTo(BeNil())
		_, ok := archApi.(*masterWorkerArchApi)
		Expect(ok).To(BeTrue(), "expected masterWorkerArchApi")
		Expect(archApi.IsMountUFSSupported()).To(BeTrue())
	})

	It("should return workersOnlyArchApi when runtimeClass is nil", func() {
		archApi := resolveArchitectureApi("test-runtime", "default", &datav1alpha1.CacheRuntime{}, nil)
		Expect(archApi).NotTo(BeNil())
		_, ok := archApi.(*workersOnlyArchApi)
		Expect(ok).To(BeTrue(), "expected workersOnlyArchApi when runtimeClass is nil")
		Expect(archApi.IsMountUFSSupported()).To(BeFalse())
	})

	It("should return workersOnlyArchApi when runtimeClass.Topology is nil", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-runtime-class-no-topology",
			},
		}
		rt := &datav1alpha1.CacheRuntime{
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-runtime-class-no-topology",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
			},
		}

		archApi := resolveArchitectureApi("test-runtime", "default", rt, runtimeClass)
		Expect(archApi).NotTo(BeNil())
		_, ok := archApi.(*workersOnlyArchApi)
		Expect(ok).To(BeTrue(), "expected workersOnlyArchApi when topology is nil")
		Expect(archApi.IsMountUFSSupported()).To(BeFalse())
	})

	It("should return workersOnlyArchApi when runtimeClass.Topology.Master is nil (workers only)", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-runtime-class-no-master",
			},
			Topology: &datav1alpha1.RuntimeTopology{
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker"}},
						},
					},
				},
			},
		}
		rt := &datav1alpha1.CacheRuntime{
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-runtime-class-no-master",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
			},
		}

		archApi := resolveArchitectureApi("test-runtime", "default", rt, runtimeClass)
		Expect(archApi).NotTo(BeNil())
		_, ok := archApi.(*workersOnlyArchApi)
		Expect(ok).To(BeTrue(), "expected workersOnlyArchApi")
		Expect(archApi.IsMountUFSSupported()).To(BeFalse())
	})

	It("should return workersOnlyArchApi when runtime is nil", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-runtime-class",
			},
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

		archApi := resolveArchitectureApi("test-runtime", "default", nil, runtimeClass)
		Expect(archApi).NotTo(BeNil())
		_, ok := archApi.(*workersOnlyArchApi)
		Expect(ok).To(BeTrue(), "expected workersOnlyArchApi when runtime is nil")
		Expect(archApi.IsMountUFSSupported()).To(BeFalse())
	})

	It("should return workersOnlyArchApi when master is disabled in runtime spec", func() {
		runtimeClass := &datav1alpha1.CacheRuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-runtime-class-master-disabled",
			},
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
		rt := &datav1alpha1.CacheRuntime{
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-runtime-class-master-disabled",
				Master: datav1alpha1.CacheRuntimeMasterSpec{
					RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
						Disabled: true,
					},
					Replicas: 0,
				},
			},
		}

		archApi := resolveArchitectureApi("test-runtime", "default", rt, runtimeClass)
		Expect(archApi).NotTo(BeNil())
		_, ok := archApi.(*workersOnlyArchApi)
		Expect(ok).To(BeTrue(), "expected workersOnlyArchApi when master is disabled")
		Expect(archApi.IsMountUFSSupported()).To(BeFalse())
	})
})
