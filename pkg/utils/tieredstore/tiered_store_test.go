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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("sortMediumType", func() {
	Describe("Len", func() {
		It("should return correct length for empty slice", func() {
			medium := sortMediumType{}
			Expect(medium.Len()).To(Equal(0))
		})

		It("should return correct length for single element", func() {
			medium := sortMediumType{common.Memory}
			Expect(medium.Len()).To(Equal(1))
		})

		It("should return correct length for multiple elements", func() {
			medium := sortMediumType{
				common.Memory,
				common.SSD,
				common.HDD,
				common.SSD,
				common.HDD,
			}
			Expect(medium.Len()).To(Equal(5))
		})

		It("should return correct length for different combinations", func() {
			testCases := []struct {
				name     string
				medium   sortMediumType
				expected int
			}{
				{
					name:     "two elements",
					medium:   sortMediumType{common.SSD, common.HDD},
					expected: 2,
				},
				{
					name:     "three elements",
					medium:   sortMediumType{common.HDD, common.SSD, common.HDD},
					expected: 3,
				},
			}

			for _, tc := range testCases {
				Expect(tc.medium.Len()).To(Equal(tc.expected), "test case: "+tc.name)
			}
		})
	})

	Describe("Swap", func() {
		It("should swap elements at given indices", func() {
			medium := sortMediumType{common.Memory, common.SSD, common.HDD}
			medium.Swap(0, 2)
			Expect(medium[0]).To(Equal(common.HDD))
			Expect(medium[2]).To(Equal(common.Memory))
			Expect(medium[1]).To(Equal(common.SSD))
		})

		It("should swap adjacent elements", func() {
			medium := sortMediumType{common.Memory, common.SSD, common.HDD, common.SSD}
			medium.Swap(1, 2)
			Expect(medium[1]).To(Equal(common.HDD))
			Expect(medium[2]).To(Equal(common.SSD))
		})

		It("should swap non-adjacent elements", func() {
			medium := sortMediumType{
				common.Memory,
				common.SSD,
				common.HDD,
				common.SSD,
				common.HDD,
			}
			medium.Swap(0, 4)
			Expect(medium[0]).To(Equal(common.HDD))
			Expect(medium[4]).To(Equal(common.Memory))
		})

		It("should handle swap with same index", func() {
			medium := sortMediumType{common.Memory, common.SSD}
			original := medium[0]
			medium.Swap(0, 0)
			Expect(medium[0]).To(Equal(original))
		})
	})

	Describe("Less", func() {
		It("should return true when first element has lower priority", func() {
			medium := sortMediumType{common.HDD, common.SSD}
			Expect(medium.Less(0, 1)).To(BeFalse())
		})

		It("should return false when first element has higher priority", func() {
			medium := sortMediumType{common.Memory, common.SSD, common.HDD}
			Expect(medium.Less(0, 2)).To(BeTrue())
		})

		It("should compare Memory and SSD correctly", func() {
			medium := sortMediumType{common.Memory, common.SSD}
			Expect(medium.Less(0, 1)).To(BeTrue())
		})

		It("should compare SSD and HDD correctly", func() {
			medium := sortMediumType{common.SSD, common.HDD}
			Expect(medium.Less(0, 1)).To(BeTrue())
		})

		It("should handle same medium types", func() {
			medium := sortMediumType{common.Memory, common.SSD, common.SSD, common.Memory}
			Expect(medium.Less(2, 3)).To(BeFalse())
			Expect(medium.Less(3, 2)).To(BeTrue())
		})

		It("should handle complex ordering scenarios", func() {
			medium := sortMediumType{
				common.Memory,
				common.SSD,
				common.HDD,
				common.SSD,
				common.Memory,
			}

			// Memory (index 4) < HDD (index 2)
			Expect(medium.Less(4, 2)).To(BeTrue())

			// SSD (index 1) > Memory (index 0)
			Expect(medium.Less(1, 0)).To(BeFalse())

			// HDD (index 2) > SSD (index 3)
			Expect(medium.Less(2, 3)).To(BeFalse())
		})
	})
})

var _ = Describe("makeMediumTypeSorted", func() {
	It("should sort and deduplicate empty slice", func() {
		input := []common.MediumType{}
		result := makeMediumTypeSorted(input)
		Expect(result).To(BeEmpty())
	})

	It("should sort single element", func() {
		input := []common.MediumType{common.Memory}
		result := makeMediumTypeSorted(input)
		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal(common.Memory))
	})

	It("should maintain sorted order for already sorted slice", func() {
		input := []common.MediumType{common.Memory, common.SSD, common.HDD}
		result := makeMediumTypeSorted(input)
		Expect(result).To(HaveLen(3))
		Expect(result[0]).To(Equal(common.Memory))
		Expect(result[1]).To(Equal(common.SSD))
		Expect(result[2]).To(Equal(common.HDD))
	})

	It("should sort unsorted slice correctly", func() {
		input := []common.MediumType{common.HDD, common.Memory, common.SSD}
		result := makeMediumTypeSorted(input)
		Expect(result).To(HaveLen(3))
		Expect(result[0]).To(Equal(common.Memory))
		Expect(result[1]).To(Equal(common.SSD))
		Expect(result[2]).To(Equal(common.HDD))
	})

	It("should remove duplicates and sort", func() {
		input := []common.MediumType{
			common.SSD,
			common.HDD,
			common.SSD,
			common.Memory,
		}
		result := makeMediumTypeSorted(input)
		Expect(result).To(HaveLen(3))
		Expect(result[0]).To(Equal(common.Memory))
		Expect(result[1]).To(Equal(common.SSD))
		Expect(result[2]).To(Equal(common.HDD))
	})

	It("should handle multiple duplicates", func() {
		input := []common.MediumType{
			common.Memory,
			common.SSD,
			common.HDD,
			common.SSD,
			common.HDD,
		}
		result := makeMediumTypeSorted(input)
		Expect(result).To(HaveLen(3))
		Expect(result[0]).To(Equal(common.Memory))
		Expect(result[1]).To(Equal(common.SSD))
		Expect(result[2]).To(Equal(common.HDD))
	})

	It("should verify no duplicates in result", func() {
		input := []common.MediumType{
			common.Memory,
			common.Memory,
			common.SSD,
			common.SSD,
			common.HDD,
		}
		result := makeMediumTypeSorted(input)

		seen := make(map[common.MediumType]bool)
		for _, mt := range result {
			Expect(seen[mt]).To(BeFalse(), "duplicate found: "+string(mt))
			seen[mt] = true
		}
	})
})

var _ = Describe("GetLevelStorageMap", func() {
	It("should return empty map for empty tiered store", func() {
		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(datav1alpha1.TieredStore{}),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetLevelStorageMap(runtimeInfo)
		Expect(result).To(BeEmpty())
	})

	It("should calculate memory cache store correctly", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.Memory,
					Path:       "/path/to/cache1/,/path/to/cache2/",
					Quota:      resource.NewQuantity(124, resource.BinarySI),
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetLevelStorageMap(runtimeInfo)
		Expect(result).To(HaveLen(1))

		memoryStore, exists := result[common.MemoryCacheStore]
		Expect(exists).To(BeTrue())
		value, ok := memoryStore.AsInt64()
		Expect(ok).To(BeTrue())
		Expect(value).To(Equal(int64(124)))
	})

	It("should aggregate multiple memory levels", func() {
		tieredStore := datav1alpha1.TieredStore{
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
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetLevelStorageMap(runtimeInfo)
		Expect(result).To(HaveLen(1))

		memoryStore, exists := result[common.MemoryCacheStore]
		Expect(exists).To(BeTrue())
		value, ok := memoryStore.AsInt64()
		Expect(ok).To(BeTrue())
		Expect(value).To(Equal(int64(248)))
	})

	It("should separate memory and disk cache stores", func() {
		tieredStore := datav1alpha1.TieredStore{
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
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetLevelStorageMap(runtimeInfo)
		Expect(result).To(HaveLen(2))

		memoryStore, exists := result[common.MemoryCacheStore]
		Expect(exists).To(BeTrue())
		memValue, ok := memoryStore.AsInt64()
		Expect(ok).To(BeTrue())
		Expect(memValue).To(Equal(int64(124)))

		diskStore, exists := result[common.DiskCacheStore]
		Expect(exists).To(BeTrue())
		diskValue, ok := diskStore.AsInt64()
		Expect(ok).To(BeTrue())
		Expect(diskValue).To(Equal(int64(512)))
	})

	It("should handle only SSD storage", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.SSD,
					Path:       "/path/to/cache1/",
					Quota:      resource.NewQuantity(1000, resource.BinarySI),
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetLevelStorageMap(runtimeInfo)
		Expect(result).To(HaveLen(1))

		diskStore, exists := result[common.DiskCacheStore]
		Expect(exists).To(BeTrue())
		value, ok := diskStore.AsInt64()
		Expect(ok).To(BeTrue())
		Expect(value).To(Equal(int64(1000)))
	})

	It("should handle only HDD storage", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.HDD,
					Path:       "/path/to/cache1/",
					Quota:      resource.NewQuantity(2000, resource.BinarySI),
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetLevelStorageMap(runtimeInfo)
		Expect(result).To(HaveLen(1))

		diskStore, exists := result[common.DiskCacheStore]
		Expect(exists).To(BeTrue())
		value, ok := diskStore.AsInt64()
		Expect(ok).To(BeTrue())
		Expect(value).To(Equal(int64(2000)))
	})
})

var _ = Describe("GetTieredLevel", func() {
	var mockQuota *resource.Quantity

	BeforeEach(func() {
		mockQuota = resource.NewQuantity(124, resource.BinarySI)
	})

	It("should return -1 for empty tiered store", func() {
		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(datav1alpha1.TieredStore{}),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetTieredLevel(runtimeInfo, common.Memory)
		Expect(result).To(Equal(-1))
	})

	It("should return -1 for non-existent medium type", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.Memory,
					Path:       "/path/to/cache1/",
					Quota:      mockQuota,
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetTieredLevel(runtimeInfo, common.SSD)
		Expect(result).To(Equal(-1))
	})

	It("should return 0 for single Memory level", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.Memory,
					Path:       "/path/to/cache1/,/path/to/cache2/",
					Quota:      mockQuota,
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetTieredLevel(runtimeInfo, common.Memory)
		Expect(result).To(Equal(0))
	})

	It("should return correct index for SSD in two-level store", func() {
		tieredStore := datav1alpha1.TieredStore{
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
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetTieredLevel(runtimeInfo, common.SSD)
		Expect(result).To(Equal(1))
	})

	It("should handle duplicate medium types and return correct sorted index", func() {
		tieredStore := datav1alpha1.TieredStore{
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
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		// After deduplication and sorting: Memory(0), SSD(1), HDD(2)
		result := GetTieredLevel(runtimeInfo, common.HDD)
		Expect(result).To(Equal(2))

		result = GetTieredLevel(runtimeInfo, common.Memory)
		Expect(result).To(Equal(0))

		result = GetTieredLevel(runtimeInfo, common.SSD)
		Expect(result).To(Equal(1))
	})

	It("should handle unsorted levels correctly", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.HDD,
					Path:       "/path/to/cache1/",
					Quota:      mockQuota,
				},
				{
					MediumType: common.Memory,
					Path:       "/path/to/cache2/",
					Quota:      mockQuota,
				},
				{
					MediumType: common.SSD,
					Path:       "/path/to/cache3/",
					Quota:      mockQuota,
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		// After sorting: Memory(0), SSD(1), HDD(2)
		Expect(GetTieredLevel(runtimeInfo, common.Memory)).To(Equal(0))
		Expect(GetTieredLevel(runtimeInfo, common.SSD)).To(Equal(1))
		Expect(GetTieredLevel(runtimeInfo, common.HDD)).To(Equal(2))
	})

	It("should return 0 for only HDD", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.HDD,
					Path:       "/path/to/cache1/",
					Quota:      mockQuota,
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetTieredLevel(runtimeInfo, common.HDD)
		Expect(result).To(Equal(0))
	})

	It("should return 0 for only SSD", func() {
		tieredStore := datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{
				{
					MediumType: common.SSD,
					Path:       "/path/to/cache1/",
					Quota:      mockQuota,
				},
			},
		}

		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			base.WithTieredStore(tieredStore),
		)
		Expect(err).NotTo(HaveOccurred())

		result := GetTieredLevel(runtimeInfo, common.SSD)
		Expect(result).To(Equal(0))
	})
})
