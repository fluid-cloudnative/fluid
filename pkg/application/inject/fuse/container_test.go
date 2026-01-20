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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("findInjectedSidecars", func() {
	Context("when pod has no injected sidecars", func() {
		It("should return empty slice", func() {
			pod1 := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test",
						},
						{
							Name: "test2",
						},
					},
				},
			}
			podObjs, err := pod.NewApplication(pod1).GetPodSpecs()
			Expect(err).NotTo(HaveOccurred())

			injectedSidecars, err := findInjectedSidecars(podObjs[0])
			
			Expect(err).NotTo(HaveOccurred())
			Expect(injectedSidecars).To(BeEmpty())
		})
	})

	Context("when pod has one injected sidecar", func() {
		It("should return the injected sidecar", func() {
			pod2 := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "fluid-fuse-0",
						},
						{
							Name: "test",
						},
					},
				},
			}
			podObjs, err := pod.NewApplication(pod2).GetPodSpecs()
			Expect(err).NotTo(HaveOccurred())

			injectedSidecars, err := findInjectedSidecars(podObjs[0])
			
			Expect(err).NotTo(HaveOccurred())
			Expect(injectedSidecars).To(HaveLen(1))
			Expect(injectedSidecars[0].Name).To(Equal("fluid-fuse-0"))
		})
	})

	Context("when pod has multiple injected sidecars", func() {
		It("should return all injected sidecars", func() {
			pod3 := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "fluid-fuse-0",
						},
						{
							Name: "test",
						},
						{
							Name: "fluid-fuse-1",
						},
					},
				},
			}
			podObjs, err := pod.NewApplication(pod3).GetPodSpecs()
			Expect(err).NotTo(HaveOccurred())

			injectedSidecars, err := findInjectedSidecars(podObjs[0])
			
			Expect(err).NotTo(HaveOccurred())
			Expect(injectedSidecars).To(HaveLen(2))
			Expect(injectedSidecars[0].Name).To(Equal("fluid-fuse-0"))
			Expect(injectedSidecars[1].Name).To(Equal("fluid-fuse-1"))
		})
	})

	Context("when GetContainers returns an error", func() {
		It("should return the error immediately", func() {
			// Create a mock FluidObject that returns an error
			mockPod := &mockFluidObject{
				shouldError: true,
			}

			injectedSidecars, err := findInjectedSidecars(mockPod)
			
			Expect(err).To(HaveOccurred())
			Expect(injectedSidecars).To(BeEmpty())
		})
	})
})

// mockFluidObject is a mock implementation of common.FluidObject for testing error handling
type mockFluidObject struct {
	shouldError bool
	containers  []corev1.Container
}

func (m *mockFluidObject) GetContainers() ([]corev1.Container, error) {
	if m.shouldError {
		return nil, ErrMockGetContainers
	}
	return m.containers, nil
}

// Implement other required methods from common.FluidObject interface
// (these are placeholders and should be adjusted based on the actual interface)
func (m *mockFluidObject) GetName() string {
	return "mock-pod"
}

func (m *mockFluidObject) GetNamespace() string {
	return "default"
}

var ErrMockGetContainers = errors.New("mock error from GetContainers")