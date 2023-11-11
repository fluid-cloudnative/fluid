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
	"regexp"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
)

// GetPodMetrics get juicefs pod metrics
func (j *JuiceFSEngine) GetPodMetrics(podName, containerName string) (metrics string, err error) {
	fileUtils := operations.NewJuiceFileUtils(podName, containerName, j.namespace, j.Log)
	metrics, err = fileUtils.GetMetric(j.getMountPoint())
	if err != nil {
		return "", err
	}
	return
}

// parseMetric parse juicefs report metric to cache
func (j JuiceFSEngine) parseMetric(metrics, edition string) (podMetric fuseMetrics) {
	var blockCacheBytes, blockCacheHits, blockCacheMiss, blockCacheHitBytes, blockCacheMissBytes string

	if edition == EnterpriseEdition {
		blockCacheBytes = BlockCacheBytesOfEnterprise
		blockCacheHits = BlockCacheHitsOfEnterprise
		blockCacheMiss = BlockCacheMissOfEnterprise
		blockCacheHitBytes = BlockCacheHitBytesOfEnterprise
		blockCacheMissBytes = BlockCacheMissBytesOfEnterprise
	} else {
		blockCacheBytes = BlockCacheBytesOfCommunity
		blockCacheHits = BlockCacheHitsOfCommunity
		blockCacheMiss = BlockCacheMissOfCommunity
		blockCacheHitBytes = BlockCacheHitBytesOfCommunity
		blockCacheMissBytes = BlockCacheMissBytesOfCommunity
	}

	counterPattern := regexp.MustCompile(`([^:\s]*):?\s?(.*)`)
	strs := strings.Split(metrics, "\n")
	for _, str := range strs {

		result := counterPattern.FindStringSubmatch(str)
		if len(result) != 3 {
			continue
		}

		switch result[1] {
		case blockCacheBytes:
			podMetric.blockCacheBytes, _ = parseInt64Size(result[2])
		case blockCacheHits:
			podMetric.blockCacheHits, _ = parseInt64Size(result[2])
		case blockCacheMiss:
			podMetric.blockCacheMiss, _ = parseInt64Size(result[2])
		case blockCacheHitBytes:
			podMetric.blockCacheHitsBytes, _ = parseInt64Size(result[2])
		case blockCacheMissBytes:
			podMetric.blockCacheMissBytes, _ = parseInt64Size(result[2])
		default:

		}
	}
	return
}
