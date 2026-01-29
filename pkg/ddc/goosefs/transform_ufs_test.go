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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("TransformDatasetToVolume", Label("pkg.ddc.goosefs.transform_ufs_test.go"), func() {
	Describe("with local mount", func() {
		type testCase struct {
			runtime        *datav1alpha1.GooseFSRuntime
			dataset        *datav1alpha1.Dataset
			expectPath     string
			expectHostPath string
		}

		DescribeTable("should transform dataset to volume correctly",
			func(tc testCase) {
				value := &GooseFS{}
				engine := &GooseFSEngine{}
				engine.transformDatasetToVolume(tc.runtime, tc.dataset, value)

				Expect(value.UFSPaths).To(HaveLen(1))
				Expect(value.UFSPaths[0].HostPath).To(Equal(tc.expectHostPath))
				Expect(value.UFSPaths[0].ContainerPath).To(Equal(tc.expectPath))
			},
			Entry("local mount without path",
				testCase{
					runtime: &datav1alpha1.GooseFSRuntime{},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "local:///mnt/test",
								Name:       "test",
							}},
						},
					},
					expectPath:     "/underFSStorage/test",
					expectHostPath: "/mnt/test",
				},
			),
			Entry("local mount with root path",
				testCase{
					runtime: &datav1alpha1.GooseFSRuntime{},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "local:///mnt/test",
								Name:       "test",
								Path:       "/",
							}},
						},
					},
					expectPath:     "/underFSStorage",
					expectHostPath: "/mnt/test",
				},
			),
		)
	})

	Describe("with PVC mount", func() {
		type testCase struct {
			runtime       *datav1alpha1.GooseFSRuntime
			dataset       *datav1alpha1.Dataset
			expectName    string
			expectPath    string
			expectSubPath string
		}

		DescribeTable("should transform dataset to PVC correctly",
			func(tc testCase) {
				value := &GooseFS{}
				engine := &GooseFSEngine{}
				engine.transformDatasetToVolume(tc.runtime, tc.dataset, value)

				Expect(value.UFSVolumes).To(HaveLen(1))
				Expect(value.UFSVolumes[0].Name).To(Equal(tc.expectName))
				Expect(value.UFSVolumes[0].ContainerPath).To(Equal(tc.expectPath))
				Expect(value.UFSVolumes[0].SubPath).To(Equal(tc.expectSubPath))
			},
			Entry("PVC mount without path",
				testCase{
					runtime: &datav1alpha1.GooseFSRuntime{},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "pvc://test",
								Name:       "test",
							}},
						},
					},
					expectName:    "test",
					expectPath:    "/underFSStorage/test",
					expectSubPath: "",
				},
			),
			Entry("PVC mount with root path",
				testCase{
					runtime: &datav1alpha1.GooseFSRuntime{},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "pvc://test1",
								Name:       "test1",
								Path:       "/",
							}},
						},
					},
					expectName:    "test1",
					expectPath:    "/underFSStorage",
					expectSubPath: "",
				},
			),
			Entry("PVC mount with subpath",
				testCase{
					runtime: &datav1alpha1.GooseFSRuntime{},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "pvc://test2/subpath",
								Name:       "test2",
							}},
						},
					},
					expectName:    "test2",
					expectPath:    "/underFSStorage/test2",
					expectSubPath: "subpath",
				},
			),
			Entry("PVC mount with subpath and root path",
				testCase{
					runtime: &datav1alpha1.GooseFSRuntime{},
					dataset: &datav1alpha1.Dataset{
						Spec: datav1alpha1.DatasetSpec{
							Mounts: []datav1alpha1.Mount{{
								MountPoint: "pvc://test3/subpath",
								Name:       "test3",
								Path:       "/",
							}},
						},
					},
					expectName:    "test3",
					expectPath:    "/underFSStorage",
					expectSubPath: "subpath",
				},
			),
		)
	})

	Describe("with node affinity", func() {
		It("should set master affinity from dataset", func() {
			runtime := &datav1alpha1.GooseFSRuntime{}
			dataset := &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "local:///mnt/test",
						Name:       "test",
					}},
					NodeAffinity: &datav1alpha1.CacheableNodeAffinity{
						Required: &v1.NodeSelector{
							NodeSelectorTerms: []v1.NodeSelectorTerm{
								{
									MatchExpressions: []v1.NodeSelectorRequirement{
										{
											Operator: v1.NodeSelectorOpIn,
											Values:   []string{"test-label-value"},
										},
									},
								},
							},
						},
					},
				},
			}
			value := &GooseFS{}

			engine := &GooseFSEngine{}
			engine.transformDatasetToVolume(runtime, dataset, value)

			Expect(value.Master.Affinity.NodeAffinity).NotTo(BeNil())
		})
	})
})
