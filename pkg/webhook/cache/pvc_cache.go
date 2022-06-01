package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	cache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	pvcsCache     *InjectionCache
	log           = ctrl.Log.WithName("webhook.cache")
	timeToLive    = 10 * time.Minute
	cleanInterval = 5 * time.Minute
)

func init() {
	pvcsCache = &InjectionCache{
		cache: cache.New(timeToLive, cleanInterval),
	}

	pvcsCache.cache.OnEvicted(func(key string, item interface{}) {
		log.V(1).Info("cache is evicted", "key", key, "info", item)
	})
}

type InjectionCache struct {
	mu    sync.RWMutex
	cache *cache.Cache
}

func GetOrCreateCachedInfo(pvc *corev1.PersistentVolumeClaim) (info *PersistentVolumeClaimCachedInfo, err error) {
	if pvcsCache == nil {
		return
	}
	return pvcsCache.GetOrCreateInfo(pvc)
}

func AddCachedInfoForPersistentVolumeClaim(info *PersistentVolumeClaimCachedInfo) (err error) {
	if pvcsCache == nil {
		return
	}

	return pvcsCache.AddInfo(info)
}

func (c *InjectionCache) AddInfo(info *PersistentVolumeClaimCachedInfo) (err error) {
	if info == nil {
		return fmt.Errorf("info to add is nil")
	}

	if info.cachedPVC == nil {
		return fmt.Errorf("cached pvc of info to add is nil")
	}

	return c.cache.Add(utils.GetNamespaceKey(info.cachedPVC), info, timeToLive)
}

func (c *InjectionCache) GetOrCreateInfo(pvc *corev1.PersistentVolumeClaim) (info *PersistentVolumeClaimCachedInfo, err error) {
	defer c.mu.Unlock()
	c.mu.Lock()
	info, found := c.Get(pvc)
	if !found {
		info := &PersistentVolumeClaimCachedInfo{
			cachedPVC: pvc,
		}
		err = c.cache.Add(utils.GetNamespaceKey(info.cachedPVC), info, timeToLive)
	}

	return info, err
}

func (c *InjectionCache) Get(pvc *corev1.PersistentVolumeClaim) (info *PersistentVolumeClaimCachedInfo, found bool) {
	if info == nil {
		log.V(1).Info("the input pvc to search is nil")
		return
	}

	// 1. find the pvc in cached by namespacedKey
	namespacedKey := utils.GetNamespaceKey(pvc)
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
