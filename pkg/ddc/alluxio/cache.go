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

package alluxio

import (
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
)

var (
	cachedPercentageFormat = "%.1f%%"
)

// queryCacheStatus checks the cache status
func (e *AlluxioEngine) queryCacheStatus() (states cacheStates, err error) {
	// get alluxio fsadmin report summary
	summary, err := e.GetReportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get Alluxio summary when query cache status")
		return states, err
	}

	if len(summary) == 0 {
		return states, errors.New("Alluxio summary is empty")
	}

	// parse alluxio fsadmin report summary
	states = e.ParseReportSummary(summary)

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) != nil {
			e.Log.Error(err, "Failed to get dataset when query cache status")
		}
		return states, err
	}

	e.patchDatasetStatus(dataset, &states)

	states.cacheHitStates = e.GetCacheHitStates()

	return states, nil

}

func (e AlluxioEngine) patchDatasetStatus(dataset *v1alpha1.Dataset, states *cacheStates) {
	// skip when `dataset.Status.UfsTotal` is empty
	if dataset.Status.UfsTotal == "" {
		return
	}
	// skip when `dataset.Status.UfsTotal` is "[Calculating]"
	if dataset.Status.UfsTotal == metadataSyncNotDoneMsg {
		return
	}

	usedInBytes, _ := utils.FromHumanSize(states.cached)
	ufsTotalInBytes, _ := utils.FromHumanSize(dataset.Status.UfsTotal)

	states.cachedPercentage = fmt.Sprintf(cachedPercentageFormat, float64(usedInBytes)/float64(ufsTotalInBytes)*100.0)

}

// getCacheHitStates gets cache hit related info by parsing Alluxio metrics
func (e *AlluxioEngine) GetCacheHitStates() (cacheHitStates cacheHitStates) {
	// get cache hit states every 1 minute(cacheHitQueryIntervalMin * 20s)
	cacheHitStates.timestamp = time.Now()
	if e.lastCacheHitStates != nil && cacheHitStates.timestamp.Sub(e.lastCacheHitStates.timestamp).Minutes() < cacheHitQueryIntervalMin {
		return *e.lastCacheHitStates
	}

	metrics, err := e.GetReportMetrics()
	if err != nil {
		e.Log.Error(err, "Failed to get Alluxio metrics when get cache hit states")
		if e.lastCacheHitStates != nil {
			return *e.lastCacheHitStates
		}
		return
	}

	e.ParseReportMetric(metrics, &cacheHitStates, e.lastCacheHitStates)

	// refresh last cache hit states
	e.lastCacheHitStates = &cacheHitStates
	return
}

// getCachedCapacityOfNode cacluates the node
//func (e *AlluxioEngine) getCurrentCachedCapacity() (totalCapacity int64, err error) {
//	workerName := e.getWorkerDaemonsetName()
//	pods, err := e.getRunningPodsOfDaemonset(workerName, e.namespace)
//	if err != nil {
//		return totalCapacity, err
//	}
//
//	for _, pod := range pods {
//		nodeName := pod.Spec.NodeName
//		if nodeName == "" {
//			e.Log.Info("The node is skipped due to its node name is null", "node", pod.Spec.NodeName,
//				"pod", pod.Name, "namespace", e.namespace)
//			continue
//		}
//
//		capacity, err := e.getCurrentCacheCapacityOfNode(nodeName)
//		if err != nil {
//			return totalCapacity, err
//		}
//		totalCapacity += capacity
//	}
//
//	return
//
//}

// getCurrentCacheCapacityOfNode cacluates the node
//func (e *AlluxioEngine) getCurrentCacheCapacityOfNode(nodeName string) (capacity int64, err error) {
//	labelName := e.getStoragetLabelname(humanReadType, totalStorageType)
//
//	node, err := kubeclient.GetNode(e.Client, nodeName)
//	if err != nil {
//		return capacity, err
//	}
//
//	if !kubeclient.IsReady(*node) {
//		e.Log.Info("Skip the not ready node", "node", node.Name)
//		return 0, nil
//	}
//
//	for k, v := range node.Labels {
//		if k == labelName {
//			value := "0"
//			if v != "" {
//				value = v
//			}
//			// capacity = units.BytesSize(float64(value))
//			capacity, err = units.RAMInBytes(value)
//			if err != nil {
//				return capacity, err
//			}
//			e.Log.V(1).Info("getCurrentCacheCapacityOfNode", k, value)
//			e.Log.V(1).Info("getCurrentCacheCapacityOfNode byteSize", k, capacity)
//		}
//	}
//
//	return
//
//}

// get the value of cached
// func (e *AlluxioEngine) cachedState() (int64, error) {
// 	podName, containerName := e.getMasterPodInfo()
// 	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
// 	cached, err := fileUitls.CachedState()

// 	return int64(cached), err

// }

// clean cache
func (e *AlluxioEngine) invokeCleanCache(path string) (err error) {
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
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	cleanCacheGracePeriodSeconds, err := e.getCleanCacheGracePeriodSeconds()
	if err != nil {
		return err
	}
	return fileUitls.CleanCache(path, cleanCacheGracePeriodSeconds)

}

func (e *AlluxioEngine) getGracefulShutdownLimits() (gracefulShutdownLimits int32, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	if runtime.Spec.RuntimeManagement.CleanCachePolicy.MaxRetryAttempts != nil {
		gracefulShutdownLimits = *runtime.Spec.RuntimeManagement.CleanCachePolicy.MaxRetryAttempts
	} else {
		gracefulShutdownLimits = defaultGracefulShutdownLimits
	}

	return
}

func (e *AlluxioEngine) getCleanCacheGracePeriodSeconds() (cleanCacheGracePeriodSeconds int32, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	if runtime.Spec.RuntimeManagement.CleanCachePolicy.GracePeriodSeconds != nil {
		cleanCacheGracePeriodSeconds = *runtime.Spec.RuntimeManagement.CleanCachePolicy.GracePeriodSeconds
	} else {
		cleanCacheGracePeriodSeconds = defaultCleanCacheGracePeriodSeconds
	}

	return
}
