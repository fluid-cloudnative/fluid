/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"regexp"
	"strings"
)

// getPodMetrics get juicefs pod metrics
func (j JuiceFSEngine) getPodMetrics(podName string) (metrics string, err error) {
	fileUtils := operations.NewJuiceFileUtils(podName, common.JuiceFSFuseContainer, j.namespace, j.Log)
	metrics, err = fileUtils.GetMetric()
	if err != nil {
		return "", err
	}
	return
}

// parseMetric parse juicefs report metric to cache
func (j JuiceFSEngine) parseMetric(metrics string) (podMetric fuseMetrics) {
	strs := strings.Split(metrics, "\n")
	for _, str := range strs {
		if strings.HasPrefix(str, "#") {
			continue
		}
		counterPattern := regexp.MustCompile(`}\s(.*)`)

		if strings.HasPrefix(str, BlockCacheBytes) {
			podMetric.blockCacheBytes, _ = parseInt64Size(counterPattern.FindStringSubmatch(str)[1])
		}
		if strings.HasPrefix(str, BlockCacheHits) {
			podMetric.blockCacheHits, _ = parseInt64Size(counterPattern.FindStringSubmatch(str)[1])
		}
		if strings.HasPrefix(str, BlockCacheMiss) {
			podMetric.blockCacheMiss, _ = parseInt64Size(counterPattern.FindStringSubmatch(str)[1])
		}
		if strings.HasPrefix(str, BlockCacheHitBytes) {
			podMetric.blockCacheHitsBytes, _ = parseInt64Size(counterPattern.FindStringSubmatch(str)[1])
		}
		if strings.HasPrefix(str, BlockCacheMissBytes) {
			podMetric.blockCacheMissBytes, _ = parseInt64Size(counterPattern.FindStringSubmatch(str)[1])
		}
		if strings.HasPrefix(str, UsedSpace) {
			podMetric.usedSpace, _ = parseInt64Size(counterPattern.FindStringSubmatch(str)[1])
		}
	}
	return
}
