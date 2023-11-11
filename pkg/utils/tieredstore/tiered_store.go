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
	"sort"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("tieredStore")

type sortMediumType []common.MediumType

func (s sortMediumType) Len() int {
	return len(s)
}

func (s sortMediumType) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortMediumType) Less(i, j int) bool {
	return common.GetDefaultTieredStoreOrder(s[i]) < common.GetDefaultTieredStoreOrder(s[j])
}

// makeMediumTypeSorted get a newly sorted MediumTypes without repeating MediumType
func makeMediumTypeSorted(mediumTypes []common.MediumType) []common.MediumType {
	newMediumTypes := make(sortMediumType, 0, len(mediumTypes))
	knownMediumTypes := map[common.MediumType]bool{}
	for _, c := range mediumTypes {
		if _, found := knownMediumTypes[c]; found {
			continue
		}
		newMediumTypes = append(newMediumTypes, c)
		knownMediumTypes[c] = true
	}
	sort.Sort(newMediumTypes)
	return newMediumTypes
}

// GetLevelStorageMap gets the level storage map
func GetLevelStorageMap(runtimeInfo base.RuntimeInfoInterface) (storage map[common.CacheStoreType]*resource.Quantity) {
	storage = map[common.CacheStoreType]*resource.Quantity{}

	for _, level := range runtimeInfo.GetTieredStoreInfo().Levels {
		storageType := common.MemoryCacheStore
		if level.MediumType == common.SSD || level.MediumType == common.HDD {
			storageType = common.DiskCacheStore
		}

		totalQuota := resource.NewQuantity(0, resource.BinarySI)

		if capacity, found := storage[storageType]; found {
			totalQuota = capacity
		}
		for _, cachePath := range level.CachePaths {
			totalQuota.Add(*cachePath.Quota)
		}

		storage[storageType] = totalQuota
	}

	return storage

}

// GetTieredLevel returns index of the given mediumType
func GetTieredLevel(runtimeInfo base.RuntimeInfoInterface, mediumType common.MediumType) int {
	var levels []common.MediumType
	for _, level := range runtimeInfo.GetTieredStoreInfo().Levels {
		levels = append(levels, level.MediumType)
	}

	log.Info("GetTieredLevel", "levels", levels)
	orderedLevels := makeMediumTypeSorted(levels)
	log.Info("GetTieredLevel", "orderedLevels", orderedLevels)
	for index, value := range orderedLevels {
		if value == mediumType {
			return index
		}
	}

	return -1
}
