/*

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

import "github.com/fluid-cloudnative/fluid/pkg/common"

func ServerlessEnabled(infos map[string]string) (match bool) {
	return matchedValue(infos, common.ServerlessPlatform, common.ASKPlatform) || enabled(infos, common.InjectServerless) || enabled(infos, common.InjectFuseSidecar)
}

func FuseSidecarEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectFuseSidecar)
}

func FuseSidecarUnprivileged(infos map[string]string) (match bool) {
	return matchedValue(infos, common.ServerlessPlatform, common.ASKPlatform) || (ServerlessEnabled(infos) && enabled(infos, common.InjectUnprivilegedFuseSidecar))
}

func WorkerSidecarEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectWorkerSidecar)
}

func InjectSidecarDone(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectSidecarDone)
}

func InjectCacheDirEnabled(infos map[string]string) (match bool) {
	return enabled(infos, common.InjectCacheDir)
}

func enabled(infos map[string]string, name string) (match bool) {
	for key, value := range infos {
		if key == name && value == common.True {
			match = true
			break
		}
	}
	return
}

func matchedValue(infos map[string]string, name string, val string) (match bool) {
	for key, value := range infos {
		if key == name && value == val {
			match = true
			break
		}
	}
	return
}
