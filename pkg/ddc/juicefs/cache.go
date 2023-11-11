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

package juicefs

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// queryCacheStatus checks the cache status
func (j *JuiceFSEngine) queryCacheStatus() (states cacheStates, err error) {
	edition := j.GetEdition()

	var cachesize uint64
	if len(j.runtime.Spec.TieredStore.Levels) != 0 {
		cachesize, err = strconv.ParseUint(strconv.FormatInt(j.runtime.Spec.TieredStore.Levels[0].Quota.Value(), 10), 10, 64)
		if err != nil {
			return
		}
	}
	// if cacheSize is overwritten in worker options, deprecated
	if cacheSizeStr := j.runtime.Spec.Worker.Options["cache-size"]; cacheSizeStr != "" {
		var cacheSizeMB uint64
		cacheSizeMB, err = strconv.ParseUint(cacheSizeStr, 10, 64)
		if err != nil {
			return
		}

		// cacheSize is in MiB
		cachesize = cacheSizeMB * 1024 * 1024
	}
	if cachesize != 0 {
		states.cacheCapacity = utils.BytesSize(float64(cachesize * uint64(j.runtime.Status.WorkerNumberReady)))
	}

	var containerName string
	var pods []corev1.Pod
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
