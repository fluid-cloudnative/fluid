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

package alluxio

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// queryCacheStatus checks the cache status
func (e *AlluxioEngine) queryCacheStatus() (states cacheStates, err error) {
	summary, err := e.reportSummary()
	if err != nil {
		e.Log.Error(err, "Failed to get Alluxio summary when query cache status")
		return states, err
	}
	strs := strings.Split(summary, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, SUMMARY_PREFIX_TOTAL_CAPACITY) {
			totalCacheCapacityAlluxio, _ := utils.FromHumanSize(strings.TrimPrefix(str, SUMMARY_PREFIX_TOTAL_CAPACITY))
			// Convert Alluxio's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cacheCapacity = utils.BytesSize(float64(totalCacheCapacityAlluxio))
		}
		if strings.HasPrefix(str, SUMMARY_PREFIX_USED_CAPACITY) {
			usedCacheCapacityAlluxio, _ := utils.FromHumanSize(strings.TrimPrefix(str, SUMMARY_PREFIX_USED_CAPACITY))
			// Convert Alluxio's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cached = utils.BytesSize(float64(usedCacheCapacityAlluxio))
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
		states.cachedPercentage = fmt.Sprintf("%.1f%%", float64(usedInBytes)/float64(ufsTotalInBytes)*100.0)
	}

	states.cacheHitStates = e.getCacheHitStates()

	return states, nil

	// dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	// if err != nil {
	// 	e.Log.Error(err, "Failed to sync the cache")
	// 	return states, err
	// }

	// totalToCache, err := units.RAMInBytes(totalToCacheStr)
	// if err != nil {
	// 	return states, err
	// }

	// check the cached
	// cached, err := e.cachedState()
	// if err != nil {
	// 	return states, err
	// }

	//_, cached, cachedPercentage, err := e.du()
	//if err != nil {
	//	return states, err
	//}
	//
	//return cacheStates{
	//	cacheCapacity:    units.BytesSize(float64(cacheCapacity)),
	//	cached:           units.BytesSize(float64(cached)),
	//	cachedPercentage: cachedPercentage,
	//}, nil

}

// getCacheHitStates gets cache hit related info by parsing Alluxio metrics
func (e *AlluxioEngine) getCacheHitStates() (cacheHitStates cacheHitStates) {
	// get cache hit states every 1 minute(CACHE_HIT_QUERY_INTERVAL_MIN * 20s)
	cacheHitStates.timestamp = time.Now()
	if e.lastCacheHitStates != nil && cacheHitStates.timestamp.Sub(e.lastCacheHitStates.timestamp).Minutes() < CACHE_HIT_QUERY_INTERVAL_MIN {
		return *e.lastCacheHitStates
	}

	metrics, err := e.reportMetrics()
	if err != nil {
		e.Log.Error(err, "Failed to get Alluxio metrics when get cache hit states")
		if e.lastCacheHitStates != nil {
			return *e.lastCacheHitStates
		}
		return
	}

	var localThroughput, remoteThroughput, ufsThroughput int64
	strs := strings.Split(metrics, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		counterPattern := regexp.MustCompile(`\(Type:\sCOUNTER,\sValue:\s(.*)\)`)
		gaugePattern := regexp.MustCompile(`\(Type:\sGAUGE,\sValue:\s(.*)/MIN\)`)
		if strings.HasPrefix(str, METRICS_PREFIX_BYTES_READ_LOCAL) {
			cacheHitStates.bytesReadLocal, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, METRICS_PREFIX_BYTES_READ_REMOTE) {
			cacheHitStates.bytesReadRemote, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, METRICS_PREFIX_BYTES_READ_UFS_ALL) {
			cacheHitStates.bytesReadUfsAll, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, METRICS_PREFIX_BYTES_READ_LOCAL_THROUGHPUT) {
			localThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, METRICS_PREFIX_BYTES_READ_REMOTE_THROUGHPUT) {
			remoteThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, METRICS_PREFIX_BYTES_READ_UFS_THROUGHPUT) {
			ufsThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
		}
	}

	if e.lastCacheHitStates == nil {
		e.lastCacheHitStates = &cacheHitStates
		return
	}

	// Summarize local/remote cache hit ratio
	deltaReadLocal := cacheHitStates.bytesReadLocal - e.lastCacheHitStates.bytesReadLocal
	deltaReadRemote := cacheHitStates.bytesReadRemote - e.lastCacheHitStates.bytesReadRemote
	deltaReadUfsAll := cacheHitStates.bytesReadUfsAll - e.lastCacheHitStates.bytesReadUfsAll
	deltaReadTotal := deltaReadLocal + deltaReadRemote + deltaReadUfsAll

	if deltaReadTotal != 0 {
		cacheHitStates.localHitRatio = fmt.Sprintf("%.1f%%", float64(deltaReadLocal)*100.0/float64(deltaReadTotal))
		cacheHitStates.remoteHitRatio = fmt.Sprintf("%.1f%%", float64(deltaReadRemote)*100.0/float64(deltaReadTotal))
		cacheHitStates.cacheHitRatio = fmt.Sprintf("%.1f%%", float64(deltaReadLocal+deltaReadRemote)*100.0/float64(deltaReadTotal))
	} else {
		// No data is requested in last minute
		cacheHitStates.localHitRatio = "0.0%"
		cacheHitStates.remoteHitRatio = "0.0%"
		cacheHitStates.cacheHitRatio = "0.0%"
	}

	// Summarize local/remote throughput ratio
	totalThroughput := localThroughput + remoteThroughput + ufsThroughput
	if totalThroughput != 0 {
		cacheHitStates.localThroughputRatio = fmt.Sprintf("%.1f%%", float64(localThroughput)*100.0/float64(totalThroughput))
		cacheHitStates.remoteThroughputRatio = fmt.Sprintf("%.1f%%", float64(remoteThroughput)*100.0/float64(totalThroughput))
		cacheHitStates.cacheThroughputRatio = fmt.Sprintf("%.1f%%", float64(localThroughput+remoteThroughput)*100.0/float64(totalThroughput))
	} else {
		cacheHitStates.localThroughputRatio = "0.0%"
		cacheHitStates.remoteThroughputRatio = "0.0%"
		cacheHitStates.cacheThroughputRatio = "0.0%"
	}

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
	masterName := e.getMasterStatefulsetName()
	master, err := e.getMasterStatefulset(masterName, e.namespace)
	if err != nil {
		e.Log.Info("Failed to get master", "err", err.Error())
		return
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
	return fileUitls.CleanCache(path)

}
