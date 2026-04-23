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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Dataset", func() {
	Describe("GetPhysicalDatasetFromMounts", func() {
		DescribeTable("counts only dataset reference mounts",
			func(mounts []datav1alpha1.Mount, expected []types.NamespacedName) {
				Expect(GetPhysicalDatasetFromMounts(mounts)).To(Equal(expected))
			},
			Entry("all mounts are dataset references",
				[]datav1alpha1.Mount{
					{MountPoint: "dataset://ns-a/n-a"},
					{MountPoint: "dataset://ns-b/n-b"},
				},
				[]types.NamespacedName{{Namespace: "ns-a", Name: "n-a"}, {Namespace: "ns-b", Name: "n-b"}},
			),
			Entry("mixed mount types only return dataset references",
				[]datav1alpha1.Mount{
					{MountPoint: "dataset://ns-a/n-a"},
					{MountPoint: "http://ns-b/n-b"},
				},
				[]types.NamespacedName{{Namespace: "ns-a", Name: "n-a"}},
			),
			Entry("empty mounts return no dataset references", nil, nil),
		)
	})

	Describe("GetDatasetRefName", func() {
		It("includes both namespace and name in the ref name", func() {
			refNameA := GetDatasetRefName("a-b", "c")
			refNameB := GetDatasetRefName("a", "bc")

			Expect(refNameA).To(Equal("c/a-b"))
			Expect(refNameB).To(Equal("bc/a"))
			Expect(refNameB).NotTo(Equal(refNameA))
		})
	})

	Describe("CheckReferenceDataset", func() {
		DescribeTable("validates reference dataset mounts",
			func(dataset *datav1alpha1.Dataset, wantCheck bool, wantErr bool) {
				gotCheck, err := CheckReferenceDataset(dataset)

				Expect(gotCheck).To(Equal(wantCheck))
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("rejects two dataset mounts",
				&datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://ns-a/n-a"}, {MountPoint: "dataset://ns-b/n-b"}}}},
				false,
				true,
			),
			Entry("rejects mixed dataset and non-dataset mounts",
				&datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://ns-a/n-a"}, {MountPoint: "http://ns-b/n-b"}}}},
				false,
				true,
			),
			Entry("accepts a single dataset mount",
				&datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://ns-a/n-a"}}}},
				true,
				false,
			),
			Entry("accepts datasets with no mounts",
				&datav1alpha1.Dataset{},
				false,
				false,
			),
		)
	})

	Describe("GetPhysicalDatasetSubPath", func() {
		DescribeTable("extracts dataset subpaths",
			func(dataset *datav1alpha1.Dataset, expected []string) {
				Expect(GetPhysicalDatasetSubPath(dataset)).To(Equal(expected))
			},
			Entry("returns nested subpaths",
				&datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://ns-a/ns-b/sub-c/sub-d"}}}},
				[]string{"sub-c/sub-d"},
			),
			Entry("returns an empty subpath when path ends with slash",
				&datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://ns-a/ns-b/"}}}},
				[]string{""},
			),
			Entry("returns nil when no subpath exists",
				&datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://ns-a/ns-b"}}}},
				nil,
			),
		)
	})

	Describe("GetOwnerDatasetUIDFromRuntimeMeta", func() {
		DescribeTable("extracts the owner dataset UID from runtime metadata",
			func(meta metav1.ObjectMeta, expectedUID types.UID, expectedError string) {
				uid, err := GetOwnerDatasetUIDFromRuntimeMeta(meta)

				Expect(uid).To(Equal(expectedUID))
				if expectedError == "" {
					Expect(err).NotTo(HaveOccurred())
				} else {
					Expect(err).To(MatchError(ContainSubstring(expectedError)))
				}
			},
			Entry("returns the UID for a matching dataset owner",
				metav1.ObjectMeta{
					Name: "spark",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: datav1alpha1.Datasetkind,
						Name: "spark",
						UID:  types.UID("dataset-uid"),
					}},
				},
				types.UID("dataset-uid"),
				"",
			),
			Entry("returns empty UID when no dataset owner exists",
				metav1.ObjectMeta{
					Name: "spark",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "Job",
						Name: "spark",
						UID:  types.UID("job-uid"),
					}},
				},
				types.UID(""),
				"",
			),
			Entry("rejects multiple dataset owners",
				metav1.ObjectMeta{
					Name: "spark",
					OwnerReferences: []metav1.OwnerReference{
						{Kind: datav1alpha1.Datasetkind, Name: "spark", UID: types.UID("dataset-uid")},
						{Kind: datav1alpha1.Datasetkind, Name: "spark", UID: types.UID("dataset-uid-2")},
					},
				},
				types.UID(""),
				"found multiple Dataset owners",
			),
			Entry("rejects dataset owners whose name differs from the runtime",
				metav1.ObjectMeta{
					Name: "spark",
					OwnerReferences: []metav1.OwnerReference{{
						Kind: datav1alpha1.Datasetkind,
						Name: "other-dataset",
						UID:  types.UID("dataset-uid"),
					}},
				},
				types.UID(""),
				"expected to be same",
			),
		)
	})
})
