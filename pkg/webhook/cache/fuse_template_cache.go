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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

func toFuseTemplateKey(key types.NamespacedName, option common.FuseSidecarInjectOption) string {
	return key.String() + string(types.Separator) + option.String()
}

func GetFuseTemplateByKey(key types.NamespacedName, option common.FuseSidecarInjectOption) (info *common.FuseInjectionTemplate, found bool) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "fuseTemplateCache.GetFuseTemplateByKey",
			"pvc.name", key.Name, "pvc.namespace", key.Namespace)
	}
	if fuseTemplateCache == nil {
		log.V(1).Info("Runtime Info Cache is disabled.")
		return
	}

	var entry interface{}
	entry, found = fuseTemplateCache.Get(toFuseTemplateKey(key, option))
	if found {
		info = entry.(*common.FuseInjectionTemplate)
	}

	return
}

func AddFuseTemplateByKey(k types.NamespacedName, option common.FuseSidecarInjectOption, info *common.FuseInjectionTemplate) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "fuseTemplateCache.AddFuseTemplateByKey",
			"pvc.name", k.Name, "pvc.namespace", k.Namespace)
	}
	if fuseTemplateCache == nil {
		log.V(1).Info("FuseTemplate Info Cache is disabled.")
		return
	}

	key := toFuseTemplateKey(k, option)
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
