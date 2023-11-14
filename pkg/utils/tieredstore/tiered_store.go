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
