/*

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
	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/api/resource"
)

// GetLevelStorageMap gets the level storage map
func GetLevelStorageMap(runtime *datav1alpha1.AlluxioRuntime) (storage map[common.CacheStoreType]*resource.Quantity) {
	storage = map[common.CacheStoreType]*resource.Quantity{}

	for _, level := range runtime.Spec.Tieredstore.Levels {
		storageType := common.MemoryCacheStore
		if level.MediumType == common.SSD {
			storageType = common.DiskCacheStore
		}

		if capacity, found := storage[storageType]; found {
			capacity.Add(*level.Quota)
			storage[storageType] = capacity
		} else {
			storage[storageType] = level.Quota
		}
	}

	return storage

}
