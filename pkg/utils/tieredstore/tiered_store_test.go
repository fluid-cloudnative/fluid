/*
Copyright 2023 The Fluid Author.

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

package tieredstore

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("TieredStore", func() {

	Describe("sortMediumType", func() {
		Context("Len", func() {
			It("should return the correct length", func() {
				testCases := []struct {
					name       string
					sortMedium sortMediumType
				}{
					{
						name: "5 elements",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
							common.SSD,
							common.HDD,
						},
					},
					{
						name: "2 elements",
						sortMedium: []common.MediumType{
							common.SSD,
							common.HDD,
						},
					},
					{
						name: "3 elements",
						sortMedium: []common.MediumType{
							common.HDD,
							common.SSD,
							common.HDD,
						},
					},
				}

				for _, tc := range testCases {
					Expect(tc.sortMedium.Len()).To(Equal(len(tc.sortMedium)), "test case: %s", tc.name)
				}
			})
		})

		Context("Swap", func() {
			It("should correctly swap elements", func() {
				testCases := []struct {
					name       string
					sortMedium sortMediumType
					i          int
					j          int
				}{
					{
						name: "swap indices 2 and 1",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
						},
						i: 2,
						j: 1,
					},
					{
						name: "swap indices 1 and 3",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
							common.SSD,
						},
						i: 1,
						j: 3,
					},
					{
						name: "swap indices 4 and 2",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
							common.SSD,
							common.HDD,
						},
						i: 4,
						j: 2,
					},
				}

				for _, tc := range testCases {
					temp := make([]common.MediumType, len(tc.sortMedium))
					copy(temp, tc.sortMedium)

					tc.sortMedium.Swap(tc.i, tc.j)

					Expect(tc.sortMedium[tc.i]).To(Equal(temp[tc.j]), "test case: %s", tc.name)
					Expect(tc.sortMedium[tc.j]).To(Equal(temp[tc.i]), "test case: %s", tc.name)
				}
			})
		})

		Context("Less", func() {
			It("should correctly compare elements", func() {
				testCases := []struct {
					name       string
					sortMedium sortMediumType
					i          int
					j          int
					want       bool
				}{
					{
						name: "HDD vs SSD",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
						},
						i:    2,
						j:    1,
						want: false,
					},
					{
						name: "SSD vs SSD",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
							common.SSD,
						},
						i:    1,
						j:    3,
						want: false,
					},
					{
						name: "Memory vs HDD",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.HDD,
							common.SSD,
							common.Memory,
						},
						i:    4,
						j:    2,
						want: true,
					},
					{
						name: "Memory vs SSD",
						sortMedium: []common.MediumType{
							common.Memory,
							common.SSD,
							common.SSD,
							common.Memory,
						},
						i:    3,
						j:    2,
						want: true,
					},
				}

				for _, tc := range testCases {
					result := tc.sortMedium.Less(tc.i, tc.j)
					Expect(result).To(Equal(tc.want), "test case: %s", tc.name)
				}
			})
		})
	})

	Describe("makeMediumTypeSorted", func() {
		It("should return sorted and deduplicated medium types", func() {
			testCases := []struct {
				name       string
				sortMedium sortMediumType
				expected   []common.MediumType
			}{
				{
					name: "already sorted",
					sortMedium: []common.MediumType{
						common.Memory,
						common.SSD,
						common.HDD,
					},
					expected: []common.MediumType{
						common.Memory,
						common.SSD,
						common.HDD,
					},
				},
				{
					name: "with duplicates",
					sortMedium: []common.MediumType{
						common.Memory,
						common.SSD,
						common.HDD,
						common.SSD,
						common.HDD,
					},
					expected: []common.MediumType{
						common.Memory,
						common.SSD,
						common.HDD,
					},
				},
				{
					name: "unsorted with no duplicates",
					sortMedium: []common.MediumType{
						common.HDD,
						common.Memory,
						common.SSD,
					},
					expected: []common.MediumType{
						common.Memory,
						common.SSD,
						common.HDD,
					},
				},
				{
					name: "unsorted with duplicates",
					sortMedium: []common.MediumType{
						common.HDD,
						common.SSD,
						common.Memory,
						common.SSD,
						common.HDD,
						common.Memory,
					},
					expected: []common.MediumType{
						common.Memory,
						common.SSD,
						common.HDD,
					},
				},
			}

			for _, tc := range testCases {
				newMediumTypes := makeMediumTypeSorted(tc.sortMedium)

				// Check that the result has the correct length
				Expect(newMediumTypes).To(HaveLen(len(tc.expected)),
					"test case %s: incorrect length", tc.name)

				// Verify the result contains the correct set of unique elements
				for i, expectedType := range tc.expected {
					Expect(newMediumTypes[i]).To(Equal(expectedType),
						"test case %s: element at index %d should be %v, got %v",
						tc.name, i, expectedType, newMediumTypes[i])
				}

				if len(newMediumTypes) >= 2 {
					for index := 1; index < len(newMediumTypes); index++ {
						// Check for no duplicates
						Expect(newMediumTypes[index]).NotTo(Equal(newMediumTypes[index-1]),
							"test case %s: found duplicate MediumTypes", tc.name)

						// Check for correct sort order (should be <= to handle invalid types with same order)
						Expect(common.GetDefaultTieredStoreOrder(newMediumTypes[index-1])).To(
							BeNumerically("<=", common.GetDefaultTieredStoreOrder(newMediumTypes[index])),
							"test case %s: incorrect sort order", tc.name)
					}
				}
			}
		})
	})

	Describe("GetLevelStorageMap", func() {
		It("should correctly aggregate storage by cache store type", func() {
			testCases := []struct {
				name        string
				tieredStore datav1alpha1.TieredStore
				want        map[common.CacheStoreType]int64
			}{
				{
					name:        "empty tiered store",
					tieredStore: datav1alpha1.TieredStore{},
					want:        map[common.CacheStoreType]int64{},
				},
				{
					name: "single memory level",
					tieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache1/,/path/to/cache2/",
								Quota:      resource.NewQuantity(124, resource.BinarySI),
							},
						},
					},
					want: map[common.CacheStoreType]int64{
						common.MemoryCacheStore: 124,
					},
				},
				{
					name: "multiple memory levels",
					tieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache1/,/path/to/cache2/",
								Quota:      resource.NewQuantity(124, resource.BinarySI),
							},
							{
								Path:  "/path/to/cache2/,/path/to/cache3/",
								Quota: resource.NewQuantity(125, resource.BinarySI),
							},
						},
					},
					want: map[common.CacheStoreType]int64{
						common.MemoryCacheStore: 248,
					},
				},
				{
					name: "mixed storage types",
					tieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache1/,/path/to/cache2/",
								Quota:      resource.NewQuantity(124, resource.BinarySI),
							},
							{
								MediumType: common.HDD,
								Path:       "/path/to/cache3/,/path/to/cache4/",
								Quota:      resource.NewQuantity(256, resource.BinarySI),
							},
							{
								MediumType: common.SSD,
								Path:       "/path/to/cache5/,/path/to/cache6/",
								Quota:      resource.NewQuantity(256, resource.BinarySI),
							},
						},
					},
					want: map[common.CacheStoreType]int64{
						common.MemoryCacheStore: 124,
						common.DiskCacheStore:   512,
					},
				},
			}

			for _, tc := range testCases {
				runtimeInfo, err := base.BuildRuntimeInfo(
					"name",
					"namespace",
					"runtimeType",
					base.WithTieredStore(tc.tieredStore),
				)
				Expect(err).NotTo(HaveOccurred(), "test case %s: failed to build runtimeInfo", tc.name)

				result := GetLevelStorageMap(runtimeInfo)
				Expect(result).To(HaveLen(len(tc.want)), "test case %s: incorrect number of storage types", tc.name)

				for storeType, expectedQuota := range tc.want {
					actualQuantity, found := result[storeType]
					Expect(found).To(BeTrue(), "test case %s: storage type %v not found", tc.name, storeType)

					actualQuota, ok := actualQuantity.AsInt64()
					Expect(ok).To(BeTrue(), "test case %s: failed to convert quantity to int64", tc.name)
					Expect(actualQuota).To(Equal(expectedQuota), "test case %s: incorrect quota for %v", tc.name, storeType)
				}
			}
		})
	})

	Describe("GetTieredLevel", func() {
		var mockQuota *resource.Quantity

		BeforeEach(func() {
			mockQuota = resource.NewQuantity(124, resource.BinarySI)
		})

		It("should return the correct tier level index", func() {
			testCases := []struct {
				name        string
				tieredStore datav1alpha1.TieredStore
				search      common.MediumType
				want        int
			}{
				{
					name:        "empty tiered store",
					tieredStore: datav1alpha1.TieredStore{},
					search:      common.Memory,
					want:        -1,
				},
				{
					name: "single level - found",
					tieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache1/,/path/to/cache2/",
								Quota:      mockQuota,
							},
						},
					},
					search: common.Memory,
					want:   0,
				},
				{
					name: "multiple levels - found at index 1",
					tieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache1/,/path/to/cache2/",
								Quota:      mockQuota,
							},
							{
								MediumType: common.SSD,
								Path:       "/path/to/cache3/,/path/to/cache4/",
								Quota:      mockQuota,
							},
						},
					},
					search: common.SSD,
					want:   1,
				},
				{
					name: "duplicate medium types - returns sorted index",
					tieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache1/,/path/to/cache2/",
								Quota:      mockQuota,
							},
							{
								MediumType: common.Memory,
								Path:       "/path/to/cache3/,/path/to/cache4/",
								Quota:      mockQuota,
							},
							{
								MediumType: common.SSD,
								Path:       "/path/to/cache5/,/path/to/cache6/",
								Quota:      mockQuota,
							},
							{
								MediumType: common.HDD,
								Path:       "/path/to/cache7/,/path/to/cache8/",
								Quota:      mockQuota,
							},
						},
					},
					search: common.HDD,
					want:   2,
				},
			}

			for _, tc := range testCases {
				runtimeInfo, err := base.BuildRuntimeInfo(
					"name",
					"namespace",
					"runtimeType",
					base.WithTieredStore(tc.tieredStore),
				)
				Expect(err).NotTo(HaveOccurred(), "test case %s: failed to build runtimeInfo", tc.name)

				result := GetTieredLevel(runtimeInfo, tc.search)
				Expect(result).To(Equal(tc.want), "test case %s: incorrect tier level", tc.name)
			}
		})
	})
})
