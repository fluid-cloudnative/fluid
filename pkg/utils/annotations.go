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
	"fmt"
	stdlog "log"
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// DEPRECATED: the label key for Fluid webhook to determine serverless platform.
	// It's replaced by commmon.AnnotationServerlessPlatform.
	DeprecatedServerlessPlatformKey string = ""
	disableApplicationController    string = ""
)

func init() {
	if envVal, exists := os.LookupEnv(common.DeprecatedEnvServerlessPlatformKey); exists {
		DeprecatedServerlessPlatformKey = envVal
		stdlog.Printf("Found %s value %s, using it as ServerlessPlatformLabelKey", common.DeprecatedEnvServerlessPlatformKey, envVal)
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
	return KeyValueMatched(infos, common.InjectSidecarPostStart, common.False)
}

func AppContainerPostStartInjectEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectAppPostStart)
}

// ---- Utils functions to decide serverless platform ----
const (
	PlatformDefault      = "Default"
	PlatformUnprivileged = "Unprivileged"
)

func GetServerlessPlatform(metaObj metav1.ObjectMeta) (platform string, err error) {
	metaLabels := metaObj.Labels
	metaAnnotations := metaObj.Annotations

	// Setting both DeprecatedServerlessPlatformKey and common.InjectServerless is not allowed
	if KeyMatched(metaLabels, DeprecatedServerlessPlatformKey) && enabled(metaLabels, common.InjectServerless) {
		err = fmt.Errorf("\"%s\" and \"%s\" is not allowed to set together, remove \"%s\" and retry", DeprecatedServerlessPlatformKey, common.InjectServerless, DeprecatedServerlessPlatformKey)
		return
	}

	// handle deprecated serverless platform key.
	if KeyMatched(metaLabels, DeprecatedServerlessPlatformKey) {
		platform = metaLabels[DeprecatedServerlessPlatformKey]
		return
	}

	// handle deprecated common.InjectFuseSidecar. In this case,
	// only two platforms are supported: PlatformDefault and PlatformUnprivileged
	if enabled(metaLabels, common.InjectFuseSidecar) {
		if enabled(metaLabels, common.InjectUnprivilegedFuseSidecar) {
			platform = PlatformUnprivileged
		} else {
			platform = PlatformDefault
		}
		return
	}

	if enabled(metaLabels, common.InjectServerless) {
		if enabled(metaLabels, common.InjectUnprivilegedFuseSidecar) {
			platform = PlatformUnprivileged
			return
		}

		// Setting common.InjectServerless in labels and common.AnnotationServerlessPlatform in annotations
		// together to indicate the serverless platform
		if KeyMatched(metaAnnotations, common.AnnotationServerlessPlatform) {
			platform = metaAnnotations[common.AnnotationServerlessPlatform]
			return
		}

		platform = PlatformDefault
		return
	}

	// default to an empty platform, meaning no platform is found
	return "", fmt.Errorf("no serverless platform can be found from Pod's metadata")
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

// FuseSidecarPrivileged decides if the injected fuse sidecar should be privileged, only used when fuse sidecar should be injected
// - sidecar is privileged only when setting serverless.fluid.io/inject=true without unprivileged.sidecar.fluid.io/inject=true
func FuseSidecarPrivileged(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectServerless) && !(enabled(infos, common.InjectUnprivilegedFuseSidecar))
}

func InjectSidecarDone(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectSidecarDone)
}

func AppControllerDisabled(info map[string]string) (match bool) {
	return KeyMatched(info, disableApplicationController)
}

func serverlessPlatformMatched(infos map[string]string) (match bool) {
	if len(DeprecatedServerlessPlatformKey) == 0 {
		return
	}

	return KeyMatched(infos, DeprecatedServerlessPlatformKey)
}

func SkipPrecheckEnable(infos map[string]string) (match bool) {
	return enabled(infos, common.SkipPrecheckAnnotationKey)
}

// enabled checks if the given name has a value of "true"
func enabled(infos map[string]string, name string) (match bool) {
	return KeyValueMatched(infos, name, common.True)
}
