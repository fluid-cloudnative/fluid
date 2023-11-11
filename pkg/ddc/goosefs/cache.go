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

package goosefs

import (
	"errors"
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

var (
	cachedPercentageFormat = "%.1f%%"
)

// queryCacheStatus checks the cache status
func (e *GooseFSEngine) queryCacheStatus() (states cacheStates, err error) {
	// get goosefs fsadmin report summary
	summary, err := e.GetReportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get GooseFS summary when query cache status")
		return states, err
	}

	if len(summary) == 0 {
		return states, errors.New("GooseFS summary is empty")
	}

	// parse goosefs fsadmin report summary
	states = e.ParseReportSummary(summary)

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get dataset when query cache status")
		return states, err
	}

	e.patchDatasetStatus(dataset, &states)

	states.cacheHitStates = e.GetCacheHitStates()

	return states, nil

}

func (e GooseFSEngine) patchDatasetStatus(dataset *v1alpha1.Dataset, states *cacheStates) {
	// skip when `dataset.Status.UfsTotal` is empty
	if dataset.Status.UfsTotal == "" {
		return
	}
	// skip when `dataset.Status.UfsTotal` is "[Calculating]"
	if dataset.Status.UfsTotal == METADATA_SYNC_NOT_DONE_MSG {
		return
	}

	usedInBytes, _ := utils.FromHumanSize(states.cached)
	ufsTotalInBytes, _ := utils.FromHumanSize(dataset.Status.UfsTotal)

	states.cachedPercentage = fmt.Sprintf(cachedPercentageFormat, float64(usedInBytes)/float64(ufsTotalInBytes)*100.0)

}

// GetCacheHitStates gets cache hit related info by parsing GooseFS metrics
func (e *GooseFSEngine) GetCacheHitStates() (cacheHitStates cacheHitStates) {
	// get cache hit states every 1 minute(CACHE_HIT_QUERY_INTERVAL_MIN * 20s)
	cacheHitStates.timestamp = time.Now()
	if e.lastCacheHitStates != nil && cacheHitStates.timestamp.Sub(e.lastCacheHitStates.timestamp).Minutes() < CACHE_HIT_QUERY_INTERVAL_MIN {
		return *e.lastCacheHitStates
	}

	metrics, err := e.GetReportMetrics()
	if err != nil {
		e.Log.Error(err, "Failed to get GooseFS metrics when get cache hit states")
		if e.lastCacheHitStates != nil {
			return *e.lastCacheHitStates
		}
		return
	}

	// refresh last cache hit states
	e.ParseReportMetric(metrics, &cacheHitStates, e.lastCacheHitStates)

	e.lastCacheHitStates = &cacheHitStates
	return
}

// clean cache
func (e *GooseFSEngine) invokeCleanCache(path string) (err error) {
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
	fileUitls := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUitls.CleanCache(path)

}
