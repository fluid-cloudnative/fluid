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
	runtimeInfoCache *utilcache.LRUExpireCache
	cacheSize        = 64
	log              = ctrl.Log.WithName("webhook.cache")
	timeToLive       = 10 * time.Minute
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
		enabled bool = utils.GetBoolValueFormEnv(common.EnvEnableRuntimeInfoCache, false)
		found   bool
	)
	cacheSize, found = utils.GetIntValueFormEnv(common.EnvRuntimeInfoCacheSize)
	if !found {
		cacheSize = defaultCacheSize
	}

	if cacheSize > 0 && enabled {
		runtimeInfoCache = utilcache.NewLRUExpireCache(cacheSize)
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
