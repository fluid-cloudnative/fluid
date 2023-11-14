/*
Copyright 2023 The Fluid Authors.

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
