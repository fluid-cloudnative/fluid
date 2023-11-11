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

package utils

import (
	stdlog "log"
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

var (
	ServerlessPlatformKey        string = ""
	ServerlessPlatformVal        string = ""
	disableApplicationController string = ""
)

func init() {
	if envVal, exists := os.LookupEnv(common.EnvServerlessPlatformKey); exists {
		ServerlessPlatformKey = envVal
		stdlog.Printf("Found %s value %s, using it as ServerlessPlatformLabelKey", common.EnvServerlessPlatformKey, envVal)
	}
	if envVal, exists := os.LookupEnv(common.EnvServerlessPlatformVal); exists {
		ServerlessPlatformVal = envVal
		stdlog.Printf("Found %s value %s, using it as ServerlessPlatformLabelValue", common.EnvServerlessPlatformVal, envVal)
	}
	if envVal, exists := os.LookupEnv(common.EnvDisableApplicationController); exists {
		disableApplicationController = envVal
		stdlog.Printf("Found %s value %s, using it as disableApplicationController", common.EnvDisableApplicationController, envVal)
	}
}

//
// ---- Util functions to control pod's mutation behaviors using CSI
//

// ServerfulFuseEnabled decides if FUSE CSI related optimization should be injected, e.g. HostToContainer mountPropagation for FUSE Recovery feature.
func ServerfulFuseEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectServerfulFuse)
}

//
// ---- Util functions to control pod's fuse sidecar mutation behaviors ----
//

func InjectCacheDirEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectCacheDir)
}

func SkipSidecarPostStartInject(infos map[string]string) (match bool) {
	return matchedValue(infos, common.InjectSidecarPostStart, common.False)
}

func AppContainerPostStartInjectEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectAppPostStart)
}

// ---- Utils functions to decide serverless platform ----
const (
	PlatformDefault      = "Default"
	PlatformUnprivileged = "Unprivileged"
)

func GetServerlessPlatfrom(infos map[string]string) (platform string) {
	if matchedKey(infos, ServerlessPlatformKey) {
		return infos[ServerlessPlatformKey]
	}

	if enabled(infos, common.InjectServerless) || enabled(infos, common.InjectFuseSidecar) {
		if enabled(infos, common.InjectUnprivilegedFuseSidecar) {
			return PlatformUnprivileged
		} else {
			return PlatformDefault
		}
	}

	// default to an empty platform, meaning no platform is found
	return ""
}

// ServerlessEnabled decides if fuse sidecar should be injected, whether privileged or unprivileged
// - serverlessPlatform implies injecting unprivileged fuse sidecar
// - serverless.fluid.io/inject=true implies injecting (privileged/unprivileged) fuse sidecar,
// - [deprecated] fuse.sidecar.fluid.io/inject=true is the deprecated version of serverless.fluid.io/inject=true
func ServerlessEnabled(infos map[string]string) (match bool) {
	return serverlessPlatformMatched(infos) || enabled(infos, common.InjectServerless) || enabled(infos, common.InjectFuseSidecar)
}

// FuseSidecarUnprivileged decides if the injected fuse sidecar should be unprivileged, only used when fuse sidecar should be injected
// - serverlessPlatform implies injecting unprivileged fuse sidecar
// - serverless.fluid.io/inject=true + unprivileged.sidecar.fluid.io/inject=true implies injecting unprivileged fuse sidecar,
func FuseSidecarUnprivileged(infos map[string]string) (match bool) {
	return serverlessPlatformMatched(infos) || (ServerlessEnabled(infos) && enabled(infos, common.InjectUnprivilegedFuseSidecar))
}

func InjectSidecarDone(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectSidecarDone)
}

func AppControllerDisabled(info map[string]string) (match bool) {
	return matchedKey(info, disableApplicationController)
}

func serverlessPlatformMatched(infos map[string]string) (match bool) {
	if len(ServerlessPlatformKey) == 0 {
		return
	}

	return matchedKey(infos, ServerlessPlatformKey)
}

// enabled checks if the given name has a value of "true"
func enabled(infos map[string]string, name string) (match bool) {
	return matchedValue(infos, name, common.True)
}

// matchedValue checks if the given name has the expected value
func matchedValue(infos map[string]string, name string, val string) (match bool) {
	for key, value := range infos {
		if key == name && value == val {
			match = true
			break
		}
	}
	return
}

// matchedKey checks if the given name exists
func matchedKey(infos map[string]string, name string) (match bool) {
	for key := range infos {
		if key == name {
			match = true
			break
		}
	}
	return
}
