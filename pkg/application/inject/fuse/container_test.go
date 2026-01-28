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
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("findInjectedSidecars", func() {
	Context("when there are no injected sidecars", func() {
		It("should return an empty slice", func() {
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

	Context("when there is one injected sidecar", func() {
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

	Context("when there are multiple injected sidecars", func() {
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
						{
							Name: "fluid-fuse-dataset-xyz",
						},
					},
				},
			}
			podObjs, err := pod.NewApplication(pod3).GetPodSpecs()
			Expect(err).NotTo(HaveOccurred())

			injectedSidecars, err := findInjectedSidecars(podObjs[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(injectedSidecars).To(HaveLen(3))
			Expect(injectedSidecars[0].Name).To(Equal("fluid-fuse-0"))
			Expect(injectedSidecars[1].Name).To(Equal("fluid-fuse-1"))
			Expect(injectedSidecars[2].Name).To(Equal("fluid-fuse-dataset-xyz"))
		})
	})

	Context("when container name contains but doesn't start with fluid-fuse prefix", func() {
		It("should only return containers that start with the prefix", func() {
			pod4 := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test-fluid-fuse",
						},
						{
							Name: "fluid-fuse-0",
						},
					},
				},
			}
			podObjs, err := pod.NewApplication(pod4).GetPodSpecs()
			Expect(err).NotTo(HaveOccurred())

			injectedSidecars, err := findInjectedSidecars(podObjs[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(injectedSidecars).To(HaveLen(1))
			Expect(injectedSidecars[0].Name).To(Equal("fluid-fuse-0"))
		})
	})

})
