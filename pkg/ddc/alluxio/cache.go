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
	units "github.com/docker/go-units"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// queryCacheStatus checks the cache status
func (e *AlluxioEngine) queryCacheStatus() (states cacheStates, err error) {
	cacheCapacity, err := e.getCurrentCachedCapacity()
	if err != nil {
		e.Log.Error(err, "Failed to sync the cache")
		return states, err
	}

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

	_, cached, cachedPercentage, err := e.du()
	if err != nil {
		return states, err
	}

	return cacheStates{
		cacheCapacity:    units.BytesSize(float64(cacheCapacity)),
		cached:           units.BytesSize(float64(cached)),
		cachedPercentage: cachedPercentage,
	}, nil

}

// getCachedCapacityOfNode cacluates the node
func (e *AlluxioEngine) getCurrentCachedCapacity() (totalCapacity int64, err error) {
	workerName := e.getWorkerDaemonsetName()
	pods, err := e.getRunningPodsOfDaemonset(workerName, e.namespace)
	if err != nil {
		return totalCapacity, err
	}

	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			e.Log.Info("The node is skipped due to its node name is null", "node", pod.Spec.NodeName,
				"pod", pod.Name, "namespace", e.namespace)
			continue
		}

		capacity, err := e.getCurrentCacheCapacityOfNode(nodeName)
		if err != nil {
			return totalCapacity, err
		}
		totalCapacity += capacity
	}

	return

}

// getCurrentCacheCapacityOfNode cacluates the node
func (e *AlluxioEngine) getCurrentCacheCapacityOfNode(nodeName string) (capacity int64, err error) {
	labelName := e.getStoragetLabelname(humanReadType, totalStorageType)

	node, err := kubeclient.GetNode(e.Client, nodeName)
	if err != nil {
		return capacity, err
	}

	if !kubeclient.IsReady(*node) {
		e.Log.Info("Skip the not ready node", "node", node.Name)
		return 0, nil
	}

	for k, v := range node.Labels {
		if k == labelName {
			value := "0"
			if v != "" {
				value = v
			}
			// capacity = units.BytesSize(float64(value))
			capacity, err = units.RAMInBytes(value)
			if err != nil {
				return capacity, err
			}
			e.Log.V(1).Info("getCurrentCacheCapacityOfNode", k, value)
			e.Log.V(1).Info("getCurrentCacheCapacityOfNode byteSize", k, capacity)
		}
	}

	return

}

// get the value of cached
// func (e *AlluxioEngine) cachedState() (int64, error) {
// 	podName, containerName := e.getMasterPodInfo()
// 	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
// 	cached, err := fileUitls.CachedState()

// 	return int64(cached), err

// }

// clean cache
func (e *AlluxioEngine) invokeCleanCache(path string) (err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUitls.CleanCache(path)

}
