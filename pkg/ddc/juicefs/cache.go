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
	"fmt"
	"strconv"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// queryCacheStatus checks the cache status
func (j *JuiceFSEngine) queryCacheStatus() (states cacheStates, err error) {
	dsName := j.getFuseDaemonsetName()
	pods, err := j.GetRunningPodsOfDaemonset(dsName, j.namespace)
	if err != nil || len(pods) == 0 {
		return
	}
	podMetrics := []fuseMetrics{}
	for _, pod := range pods {
		podMetricStr, err := j.GetPodMetrics(pod.Name)
		if err != nil {
			return states, err
		}
		podMetric := j.parseMetric(podMetricStr)
		podMetrics = append(podMetrics, podMetric)
	}

	var totalSpace int64
	if len(podMetrics) != 0 {
		totalSpace = podMetrics[0].usedSpace
	}
	var totalCache, totalCacheHits, totalCacheMiss, totalCacheHitThroughput, totalCacheMissThroughput int64
	for _, p := range podMetrics {
		totalCache += p.blockCacheBytes
		totalCacheHits += p.blockCacheHits
		totalCacheMiss += p.blockCacheMiss
		totalCacheHitThroughput += p.blockCacheHitsBytes
		totalCacheMissThroughput += p.blockCacheMissBytes
	}

	// caches = total cache / fuse pod num
	states.cached = utils.BytesSize(float64(totalCache) / float64(len(podMetrics)))

	// cachePercent = cached / total space
	if totalSpace != 0 {
		states.cachedPercentage = fmt.Sprintf("%.1f%%", float64(totalCache)*100.0/float64(int64(len(podMetrics))*totalSpace))
	} else {
		states.cachedPercentage = "0.0%"
	}

	// cacheHitRatio = total cache hits / (total cache hits + total cache miss)
	totalCacheCounts := totalCacheHits + totalCacheMiss
	if totalCacheCounts != 0 {
		states.cacheHitRatio = fmt.Sprintf("%.1f%%", float64(totalCacheHits)*100.0/float64(totalCacheCounts))
	} else {
		states.cacheHitRatio = "0.0%"
	}

	// cacheHitRatio = total cache hits / (total cache hits + total cache miss)
	totalCacheThroughput := totalCacheHitThroughput + totalCacheMissThroughput
	if totalCacheThroughput != 0 {
		states.cacheThroughputRatio = fmt.Sprintf("%.1f%%", float64(totalCacheHitThroughput)*100.0/float64(totalCacheThroughput))
	} else {
		states.cacheThroughputRatio = "0.0%"
	}

	if len(j.runtime.Spec.TieredStore.Levels) != 0 {
		//cacheCapacity = cachesize * numberworker
		cachesize, e := strconv.ParseUint(j.runtime.Spec.TieredStore.Levels[0].Quota.String(), 10, 64)
		if e != nil {
			return
		}
		states.cacheCapacity = utils.BytesSize(float64(1024 * 1024 * cachesize * uint64(len(pods))))
	}

	return
}
