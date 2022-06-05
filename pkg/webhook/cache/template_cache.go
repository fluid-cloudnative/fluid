package cache

import (
	"strconv"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

const (
	Separator = '/'
)

func toFuseTemplateKey(key types.NamespacedName, enableCacheDir bool) string {
	return key.String() + string(Separator) + strconv.FormatBool(enableCacheDir)
}

func GetFuseTemplateByKey(key types.NamespacedName, enableCacheDir bool) (info *common.FuseInjectionTemplate, found bool) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "fuseTemplateCache.GetFuseTemplateByKey",
			"pvc.name", key.Name, "pvc.namespace", key.Namespace)
	}
	if fuseTemplateCache == nil {
		log.V(1).Info("Runtime Info Cache is disabled.")
		return
	}

	var entry interface{}
	entry, found = fuseTemplateCache.Get(toFuseTemplateKey(key, enableCacheDir))
	if found {
		info = entry.(*common.FuseInjectionTemplate)
	}

	return
}

func AddFuseTemplateByKey(k types.NamespacedName, enableCacheDir bool, info *common.FuseInjectionTemplate) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "fuseTemplateCache.AddFuseTemplateByKey",
			"pvc.name", k.Name, "pvc.namespace", k.Namespace)
	}
	if fuseTemplateCache == nil {
		log.V(1).Info("FuseTemplate Info Cache is disabled.")
		return
	}

	key := toFuseTemplateKey(k, enableCacheDir)
	result, found := fuseTemplateCache.Get(key)
	if !found {
		fuseTemplateCache.Add(key, info, timeToLive)
		log.V(1).Info("add to fuseTemplateCache",
			"key", key, "info", info)
	} else {
		log.V(1).Info("skip adding to fuseTemplateCache, because it's already there",
			"key", key, "info", result)
	}
}
