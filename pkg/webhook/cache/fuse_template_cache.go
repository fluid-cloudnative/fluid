/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
