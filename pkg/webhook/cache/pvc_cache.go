package cache

import (
	"sync"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cache "github.com/patrickmn/go-cache"
)

type PVCInfoCache struct {
	// The uuid is used to check if the pvcCache is validate.
	// It's pvc's uuid
	uuid string

	// Check if the pvc belongs to the fluid dataset
	isDataset bool

	// The runtime Info to cache
	runtimeInfo base.RuntimeInfoInterface

	// The fuse template to inject
	fuseTemplateToInject *common.FuseInjectionTemplate
}

var pvcsCache *InjectionCache

func init() {
	pvcsCache = &InjectionCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

type InjectionCache struct {
	mu    sync.RWMutex
	cache *cache.Cache
}

func GetInjectionCacheForPVCs() *InjectionCache {
	return pvcsCache
}
