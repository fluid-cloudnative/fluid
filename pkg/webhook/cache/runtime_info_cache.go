package cache

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	corev1 "k8s.io/api/core/v1"
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

func GetRuntimeInfoByPVC(pvc *corev1.PersistentVolumeClaim) (info base.RuntimeInfoInterface, found bool) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "runtimeInfoCache.GetRuntimeInfoByPVC",
			"pvc.name", pvc.GetName(), "pvc.namespace", pvc.GetNamespace())
	}
	if runtimeInfoCache == nil {
		log.V(1).Info("Runtime Info Cache is disabled.")
		return
	}
	if pvc == nil || len(pvc.GetUID()) == 0 {
		log.V(1).Info("PVC key is nil.")
		return
	}

	var entry interface{}
	entry, found = runtimeInfoCache.Get(pvc.GetUID())
	if found {
		info = entry.(base.RuntimeInfoInterface)
	}

	return
}

func AddRuntimeInfoByPVC(pvc *corev1.PersistentVolumeClaim, info base.RuntimeInfoInterface) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "runtimeInfoCache.AddRuntimeInfoByPVC",
			"pvc.name", pvc.GetName(), "pvc.namespace", pvc.GetNamespace())
	}
	if runtimeInfoCache == nil {
		log.V(1).Info("Runtime Info Cache is disabled.")
		return
	}
	if pvc == nil || len(pvc.GetUID()) == 0 {
		log.V(1).Info("PVC key is nil.")
		return
	}

	_, found := runtimeInfoCache.Get(pvc.GetUID())
	if !found {
		runtimeInfoCache.Add(pvc.UID, info, timeToLive)
		log.V(1).Info("add to runtimeInfoCache",
			"pvc", pvc, "info", info)
	} else {
		log.V(1).Info("skip adding to runtimeInfoCache, because it's already there",
			"pvc", pvc, "info", info)
	}
}
