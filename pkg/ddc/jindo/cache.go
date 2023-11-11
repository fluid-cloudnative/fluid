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

package jindo

import (
	"fmt"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// queryCacheStatus checks the cache status
func (e *JindoEngine) queryCacheStatus() (states cacheStates, err error) {
	defer utils.TimeTrack(time.Now(), "JindoEngine.queryCacheStatus", "name", e.name, "namespace", e.namespace)
	summary, err := e.GetReportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get Jindo summary when query cache status")
		return states, err
	}
	strs := strings.Split(summary, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, SUMMARY_PREFIX_TOTAL_CAPACITY) {
			totalCacheCapacityJindo, _ := utils.FromHumanSize(strings.TrimPrefix(str, SUMMARY_PREFIX_TOTAL_CAPACITY))
			// Convert JindoFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cacheCapacity = utils.BytesSize(float64(totalCacheCapacityJindo))
		}
		if strings.HasPrefix(str, SUMMARY_PREFIX_USED_CAPACITY) {
			usedCacheCapacityJindo, _ := utils.FromHumanSize(strings.TrimPrefix(str, SUMMARY_PREFIX_USED_CAPACITY))
			// Convert JindoFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cached = utils.BytesSize(float64(usedCacheCapacityJindo))
		}
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Info("Failed to get dataset when query cache status")
		return states, err
	}

	// `dataset.Status.UfsTotal` probably haven't summed, in which case we won't compute cache percentage
	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != METADATA_SYNC_NOT_DONE_MSG {
		usedInBytes, _ := utils.FromHumanSize(states.cached)
		ufsTotalInBytes, _ := utils.FromHumanSize(dataset.Status.UfsTotal)
		// jindofs calculate cached storage bytesize with block sum, so percentage will be over 100% if totally cached
		percentTage := 0.0
		if ufsTotalInBytes != 0 {
			percentTage = float64(usedInBytes) / float64(ufsTotalInBytes)
		}
		// avoid jindo blocksize greater than ufssize
		if percentTage > 1 {
			percentTage = 1.0
		}
		states.cachedPercentage = fmt.Sprintf("%.1f%%", percentTage*100.0)
	}

	return states, nil
}

// clean cache
func (e *JindoEngine) invokeCleanCache() (err error) {
	// 1. Check if master is ready, if not, just return
	masterName := e.getMasterName()
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("Failed to get master", "err", err.Error())
			return nil
		}
		// other error
		return err
	}
	if master.Status.ReadyReplicas == 0 {
		e.Log.Info("The master is not ready, just skip clean cache.", "master", masterName)
		return nil
	} else {
		e.Log.Info("The master is ready, so start cleaning cache", "master", masterName)
	}

	// 2. run clean action
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	e.Log.Info("cleaning cache and wait for a while")
	return fileUitls.CleanCache()
}
