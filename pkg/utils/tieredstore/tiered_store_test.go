/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package tieredstore

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestLen(t *testing.T) {
	testCases := map[string]sortMediumType{
		"test case 1": []common.MediumType{
			common.Memory,
			common.SSD,
			common.HDD,
			common.SSD,
			common.HDD,
		},
		"test case 2": []common.MediumType{
			common.SSD,
			common.HDD,
		},
		"test case 3": []common.MediumType{
			common.HDD,
			common.SSD,
			common.HDD,
		},
	}
	for k, item := range testCases {
		if item.Len() != len(item) {
			t.Errorf("%s check failure,want:%v,got:%v", k, len(item), item.Len())
		}
	}
}

func TestSwap(t *testing.T) {
	testCases := map[string]struct {
		sortMedium sortMediumType
		i          int
		j          int
	}{
		"test case 1": {
			sortMedium: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
			},
			i: 2,
			j: 1,
		},
		"test case 2": {
			sortMedium: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
				common.SSD,
			},
			i: 1,
			j: 3,
		},
		"test case 3": {
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
	for k, item := range testCases {
		var temp = make([]common.MediumType, len(item.sortMedium))
		_ = copy(temp, item.sortMedium)
		if item.i < item.sortMedium.Len() && item.j < item.sortMedium.Len() {
			item.sortMedium.Swap(item.i, item.j)
			if temp[item.i] != item.sortMedium[item.j] || temp[item.j] != item.sortMedium[item.i] {
				t.Errorf("%s check failure", k)
			}
		} else {
			t.Errorf("%s is not suitable", k)
		}
	}
}

func TestLess(t *testing.T) {
	testCases := map[string]struct {
		sortMedium sortMediumType
		i          int
		j          int
		want       bool
	}{
		"test case 1": {
			sortMedium: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
			},
			i:    2,
			j:    1,
			want: false,
		},
		"test case 2": {
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
		"test case 3": {
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
		"test case 4": {
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
	for k, item := range testCases {
		if item.i < item.sortMedium.Len() && item.j < item.sortMedium.Len() {
			result := item.sortMedium.Less(item.i, item.j)
			if result != item.want {
				t.Errorf("%s check failure,want:%t,got:%t", k, item.want, result)
			}
		} else {
			t.Errorf("%s is not suitable", k)
		}

	}
}

func TestMakeMediumTypeSorted(t *testing.T) {
	testCases := map[string]struct {
		sortMedium sortMediumType
		want       sortMediumType
	}{
		"test case 1": {
			sortMedium: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
			},
			want: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
			},
		},
		"test case 2": {
			sortMedium: []common.MediumType{
				common.SSD,
				common.HDD,
				common.SSD,
				"apple",
				"baba",
				common.Memory,
			},
			want: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
			},
		},
		"test case 3": {
			sortMedium: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
				common.SSD,
				common.HDD,
			},
			want: []common.MediumType{
				common.Memory,
				common.SSD,
				common.HDD,
			},
		},
	}
	for k, item := range testCases {
		newMediumTypes := makeMediumTypeSorted(item.sortMedium)
		if len(newMediumTypes) >= 2 {
			for index := 1; index < len(newMediumTypes); index++ {
				if newMediumTypes[index-1] == newMediumTypes[index] {
					t.Errorf("%s cannot paas, because of repeat MediumTypes", k)
				}
				if common.GetDefaultTieredStoreOrder(newMediumTypes[index-1]) > common.GetDefaultTieredStoreOrder(newMediumTypes[index]) {
					t.Errorf("%s cannot paas, because of wrong sort result", k)
				}
			}
		}
	}
}

func TestGetLevelStorageMap(t *testing.T) {
	testCases := map[string]struct {
		tieredStore datav1alpha1.TieredStore
		want        map[common.CacheStoreType]int64
	}{
		"test case 1": {
			tieredStore: datav1alpha1.TieredStore{},
			want:        map[common.CacheStoreType]int64{},
		},
		"test case 2": {
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
		"test case 3": {
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
		"test case 4": {
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
	for k, item := range testCases {
		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			item.tieredStore,
		)
		if err != nil {
			t.Errorf("%s cannot build the runtimeInfo", k)
		}
		result := GetLevelStorageMap(runtimeInfo)
		if len(result) != len(item.want) {
			t.Errorf("%s cannot paas, want %v types, get %v types", k, len(item.want), len(result))
		} else {
			for index, value := range result {
				int64Result, _ := value.AsInt64()
				if item.want[index] != int64Result {
					t.Errorf("%s cannot paas, want %v, get %v", k, item.want[index], int64Result)
				}
			}
		}
	}
}

func TestGetTieredLevel(t *testing.T) {
	var mockQuota = resource.NewQuantity(124, resource.BinarySI)
	testCases := map[string]struct {
		tieredStore datav1alpha1.TieredStore
		search      common.MediumType
		want        int
	}{
		"test case 1": {
			tieredStore: datav1alpha1.TieredStore{},
			search:      common.Memory,
			want:        -1,
		},
		"test case 2": {
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
		"test case 3": {
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
		"test case 4": {
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

	for k, item := range testCases {
		runtimeInfo, err := base.BuildRuntimeInfo(
			"name",
			"namespace",
			"runtimeType",
			item.tieredStore,
		)
		if err != nil {
			t.Errorf("%s cannot build the runtimeInfo", k)
		}
		result := GetTieredLevel(runtimeInfo, item.search)
		if result != item.want {
			t.Errorf("%s cannot paas, want %v, get %v", k, item.want, result)
		}
	}

}
