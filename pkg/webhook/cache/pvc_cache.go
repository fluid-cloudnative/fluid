package cache

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	pvcsCache *InjectionCache
	log       = ctrl.Log.WithName("webhook.cache")
)

func init() {
	pvcsCache = &InjectionCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	pvcsCache.cache.OnEvicted(func(key string, item interface{}) {
		log.V(1).Info("cache is evicted", "key", key, "info", item)
	})
}

type InjectionCache struct {
	// mu    sync.RWMutex
	cache *cache.Cache
}

func GetInjectionCacheForPVCs() *InjectionCache {
	return pvcsCache
}

func (c *InjectionCache) FindCachedInfoByPvc(pvc *corev1.PersistentVolumeClaim) (info *PersistentVolumeClaimCachedInfo, found bool) {
	if info == nil {
		log.V(1).Info("the input pvc to search is nil")
		return
	}

	name := pvc.GetName()
	namespace := pvc.GetNamespace()
	if len(namespace) == 0 {
		namespace = corev1.NamespaceDefault
	}
	namespacedKey := namespace + "/" + name

	// 1. find the pvc in cached by namespacedKey
	item, found := c.cache.Get(namespacedKey)
	if !found {
		return
	}

	switch v := item.(type) {
	case *PersistentVolumeClaimCachedInfo:
		info = v
	default:
		log.Info("No supported PersistentVolumeClaimCachedInfo Type", "v", v)
		return
	}

	// 2. check the uid
	if info.cachedPVC == nil {
		log.Info("The cached pvc is not found, skip.")
		c.cache.Delete(namespacedKey)
		return nil, found
	}

	if info.cachedPVC.GetUID() != pvc.UID {
		log.Info("The pvc is found, but uid not match. So the pvc is outdated. Now it's abandoned.",
			"inputPVC", pvc,
			"oldUid", info.cachedPVC)
		c.cache.Delete(namespacedKey)
		return nil, found
	}

	found = true
	return
}
