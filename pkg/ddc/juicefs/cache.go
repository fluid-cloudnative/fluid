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

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

// queryCacheStatus checks the cache status
func (j *JuiceFSEngine) queryCacheStatus() (states cacheStates, err error) {
	edition := j.GetEdition()

	if len(j.runtime.Spec.TieredStore.Levels) != 0 {
		cachesize, e := strconv.ParseUint(strconv.FormatInt(j.runtime.Spec.TieredStore.Levels[0].Quota.Value(), 10), 10, 64)
		if e != nil {
			err = e
			return
		}
		states.cacheCapacity = utils.BytesSize(float64(cachesize * uint64(j.runtime.Spec.Replicas)))
	}

	var containerName string
	var pods []v1.Pod
	// enterprise edition use cache of workers which form a cache group, while community edition use cache of fuse pod whose cache if no-sharing
	if edition == EnterpriseEdition {
		containerName = common.JuiceFSWorkerContainer
		stsName := j.getWorkerName()
		pods, err = j.GetRunningPodsOfStatefulSet(stsName, j.namespace)
		if err != nil || len(pods) == 0 {
			return
		}
	} else {
		containerName = common.JuiceFSFuseContainer
		dsName := j.getFuseDaemonsetName()
		pods, err = j.GetRunningPodsOfDaemonset(dsName, j.namespace)
		if err != nil || len(pods) == 0 {
			return
		}
	}

	podMetrics := []fuseMetrics{}
	for _, pod := range pods {
		podMetricStr, err := j.GetPodMetrics(pod.Name, containerName)
		if err != nil {
			return states, err
		}
		podMetric := j.parseMetric(podMetricStr, edition)
		podMetrics = append(podMetrics, podMetric)
	}

	var totalSpace int64
	if len(podMetrics) != 0 {
		totalSpace, _ = j.UsedStorageBytes()
	}
	var totalCache, totalCacheHits, totalCacheMiss, totalCacheHitThroughput, totalCacheMissThroughput int64
	for _, p := range podMetrics {
		totalCache += p.blockCacheBytes
		totalCacheHits += p.blockCacheHits
		totalCacheMiss += p.blockCacheMiss
		totalCacheHitThroughput += p.blockCacheHitsBytes
		totalCacheMissThroughput += p.blockCacheMissBytes
	}

	if edition == EnterpriseEdition {
		// caches = total cache of worker pod num
		states.cached = utils.BytesSize(float64(totalCache))
	} else {
		// caches = total cache / fuse pod num
		states.cached = utils.BytesSize(float64(totalCache) / float64(len(podMetrics)))
		totalCache = totalCache / int64(len(podMetrics))
	}

	// cachePercent = cached / total space
	if totalSpace != 0 {
		states.cachedPercentage = fmt.Sprintf("%.1f%%", float64(totalCache)*100.0/float64(totalSpace))
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
	return
}
