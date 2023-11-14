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

package cache

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"k8s.io/apimachinery/pkg/types"
	utilcache "k8s.io/apimachinery/pkg/util/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	defaultCacheSize = 64

	defaultTimeToLive = 10 * time.Minute
)

var (
	runtimeInfoCache  *utilcache.LRUExpireCache
	fuseTemplateCache *utilcache.LRUExpireCache
	cacheSize         = 64
	log               = ctrl.Log.WithName("webhook.cache")
	timeToLive        = 5 * time.Minute
)

// PersistentVolumeClaimInfoCache represents runtime info and whether it belongs dataset
type PersistentVolumeClaimInfoCache struct {
	info      base.RuntimeInfoInterface
	isDataset bool
}

func (c PersistentVolumeClaimInfoCache) GetRuntimeInfo() base.RuntimeInfoInterface {
	return c.info
}

func (c PersistentVolumeClaimInfoCache) IsBelongToDataset() bool {
	return c.isDataset
}

// By default, cache is disabled
func init() {
	var (
		enabled bool = utils.GetBoolValueFromEnv(common.EnvEnableRuntimeInfoCache, false)
		found   bool
	)
	cacheSize, found = utils.GetIntValueFromEnv(common.EnvRuntimeInfoCacheSize)
	if !found {
		cacheSize = defaultCacheSize
	}

	if cacheSize > 0 && enabled {
		runtimeInfoCache = utilcache.NewLRUExpireCache(cacheSize)
		fuseTemplateCache = utilcache.NewLRUExpireCache(cacheSize)
		timeToLive = utils.GetDurationValueFromEnv(common.EnvRuntimeInfoCacheTTL, defaultTimeToLive)
	}
}

func GetRuntimeInfoByKey(key types.NamespacedName) (info *PersistentVolumeClaimInfoCache, found bool) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "runtimeInfoCache.GetRuntimeInfoByKey",
			"pvc.name", key.Name, "pvc.namespace", key.Namespace)
	}
	if runtimeInfoCache == nil {
		log.V(1).Info("Runtime Info Cache is disabled.")
		return
	}

	var entry interface{}
	entry, found = runtimeInfoCache.Get(key.String())
	if found {
		info = entry.(*PersistentVolumeClaimInfoCache)
	}

	return
}

func AddRuntimeInfoByKey(key types.NamespacedName, runtimeInfo base.RuntimeInfoInterface, isDataset bool) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "runtimeInfoCache.AddRuntimeInfoByKey",
			"pvc.name", key.Name, "pvc.namespace", key.Namespace)
	}
	if runtimeInfoCache == nil {
		log.V(1).Info("Runtime Info Cache is disabled.")
		return
	}

	result, found := runtimeInfoCache.Get(key.String())
	if !found {
		info := &PersistentVolumeClaimInfoCache{
			isDataset: isDataset,
			info:      runtimeInfo,
		}
		runtimeInfoCache.Add(key.String(), info, timeToLive)
		log.V(1).Info("add to runtimeInfoCache",
			"key", key, "info", info)
	} else {
		log.V(1).Info("skip adding to runtimeInfoCache, because it's already there",
			"key", key, "info", result)
	}
}
