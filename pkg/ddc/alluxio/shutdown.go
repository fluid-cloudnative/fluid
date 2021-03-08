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
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// shut down the Alluxio engine
func (e *AlluxioEngine) Shutdown() (err error) {
	if e.retryShutdown < e.gracefulShutdownLimits {
		err = e.cleanupCache()
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed",
				// "engine", e,
				"retry times", e.retryShutdown)
			return
		}
	}

	if e.MetadataSyncDoneCh != nil {
		close(e.MetadataSyncDoneCh)
	}

	_, err = e.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = e.destroyMaster()
	if err != nil {
		return
	}

	err = e.cleanAll()
	return err
}

// destroyMaster Destroies the master
func (e *AlluxioEngine) destroyMaster() (err error) {
	found := false
	found, err = helm.CheckRelease(e.name, e.namespace)
	if err != nil {
		return err
	}

	if found {
		err = helm.DeleteRelease(e.name, e.namespace)
		if err != nil {
			return
		}
	}
	return
}

// // Destroy the workers
// func (e *AlluxioEngine) destroyWorkers() error {
// 	return nil
// }

// cleanupCache cleans up the cache
func (e *AlluxioEngine) cleanupCache() (err error) {
	// TODO(cheyang): clean up the cache
	cacheStates, err := e.queryCacheStatus()
	if cacheStates.cached == "" {
		return
	}

	e.Log.Info("The cache before cleanup",
		"cached", cacheStates.cached,
		"cachedPercentage", cacheStates.cachedPercentage)

	cached, err := utils.FromHumanSize(cacheStates.cached)
	if err != nil {
		return err
	}

	if cached == 0 {
		e.Log.Info("No need to clean cache",
			"cached", cacheStates.cached,
			"cachedPercentage", cacheStates.cachedPercentage)
		return nil
	}

	err = e.invokeCleanCache("/")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		} else if strings.Contains(err.Error(), "does not have a host assigned") {
			return nil
		}
		return err
	} else {
		e.Log.Info("Clean up the cache successfully")
	}

	time.Sleep(time.Duration(10 * time.Second))

	// ufs, cached, cachedPercentage, err = e.du()
	// if err != nil {
	// 	return
	// }

	// e.Log.Info("The cache after cleanup", "ufs", ufs,
	// 	"cached", cached,
	// 	"cachedPercentage", cachedPercentage)

	// if cached > 0 {
	// 	return fmt.Errorf("The remaining cached is not cleaned up, it still has %d", cached)
	// }

	return fmt.Errorf("the remaining cached is not cleaned up, check again")
}

// cleanAll cleans up the all
func (e *AlluxioEngine) cleanAll() (err error) {
	var (
		valueConfigmapName = e.name + "-" + e.runtimeType + "-values"
		configmapName      = e.name + "-config"
		namespace          = e.namespace
	)

	cms := []string{valueConfigmapName, configmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(e.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// destroyWorkers attempts to delete the workers until worker num reaches the given expectedWorkers, if expectedWorkers is -1, it means all the workers should be deleted
// This func returns currentWorkers representing how many workers are left after this process.
func (e *AlluxioEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
	var (
		nodeList = &corev1.NodeList{}

		labelName          = e.getRuntimeLabelname()
		labelCommonName    = e.getCommonLabelname()
		labelMemoryName    = e.getStoragetLabelname(humanReadType, memoryStorageType)
		labelDiskName      = e.getStoragetLabelname(humanReadType, diskStorageType)
		labelTotalname     = e.getStoragetLabelname(humanReadType, totalStorageType)
		labelExclusiveName = utils.GetExclusiveKey()
	)

	labelNames := []string{labelName, labelTotalname, labelDiskName, labelMemoryName, labelCommonName}
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", labelCommonName))
	if err != nil {
		return currentWorkers, err
	}

	err = e.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})

	if err != nil {
		return currentWorkers, err
	}

	currentWorkers = int32(len(nodeList.Items))
	if expectedWorkers >= currentWorkers {
		e.Log.Info("No need to scale in. Skip.")
		return currentWorkers, nil
	}

	var nodes []corev1.Node
	if expectedWorkers >= 0 {
		e.Log.V(1).Info("Scale in Alluxio workers", "expectedWorkers", expectedWorkers)
		// This is a scale in operation
		runtimeInfo, err := e.getRuntimeInfo()
		if err != nil {
			e.Log.Error(err, "getRuntimeInfo when scaling in")
			return currentWorkers, err
		}

		fuseGlobal, _ := runtimeInfo.GetFuseDeployMode()
		nodes, err = e.sortNodesToShutdown(nodeList.Items, fuseGlobal)
		if err != nil {
			return currentWorkers, err
		}

	} else {
		// Destroy all workers. This is a subprocess during deletion of AlluxioRuntime
		nodes = nodeList.Items
	}

	// 1.select the nodes
	for _, node := range nodes {
		if expectedWorkers == currentWorkers {
			break
		}
		// nodes = append(nodes, &node)
		toUpdate := node.DeepCopy()
		if len(toUpdate.Labels) == 0 {
			continue
		}

		for _, label := range labelNames {
			delete(toUpdate.Labels, label)
		}

		exclusiveLabelValue := utils.GetExclusiveValue(e.namespace, e.name)
		if val, exist := toUpdate.Labels[labelExclusiveName]; exist && val == exclusiveLabelValue {
			delete(toUpdate.Labels, labelExclusiveName)
			labelNames = append(labelNames, labelExclusiveName)
		}

		if len(toUpdate.Labels) < len(node.Labels) {
			err := e.Client.Update(context.TODO(), toUpdate)
			if err != nil {
				return expectedWorkers, err
			}
			e.Log.Info("Destroy worker", "Dataset", e.name, "deleted worker node", node.Name, "removed labels", labelNames)
		}

		currentWorkers--
	}

	return currentWorkers, nil
}

func (e *AlluxioEngine) sortNodesToShutdown(candidateNodes []corev1.Node, fuseGlobal bool) (nodes []corev1.Node, err error) {
	if !fuseGlobal {
		// If fuses are deployed in non-global mode, workers and fuses will be scaled in together.
		// It can be dangerous if we scale in nodes where there are pods using the related pvc.
		// So firstly we filter out such nodes
		pvcMountNodes, err := kubeclient.GetPvcMountNodes(e.Client, e.name, e.namespace)
		if err != nil {
			e.Log.Error(err, "GetPvcMountNodes when scaling in")
			return nil, err
		}

		for _, node := range candidateNodes {
			if _, found := pvcMountNodes[node.Name]; !found {
				nodes = append(nodes, node)
			}
		}
	} else {
		// If fuses are deployed in global mode. Scaling in workers has nothing to do with fuses.
		// All nodes with related label can be candidate nodes.
		nodes = candidateNodes
	}

	// Prefer to choose nodes with less data cache.
	// Since this is just a preference, anything unexpected will be ignored.
	worker2UsedCapacityMap, err := e.GetWorkerUsedCapacity()
	if err != nil {
		e.Log.Info("GetWorkerUsedCapacity when sorting nodes to be shutdowned. Got err: %v. Ignore it", err)
	}

	if worker2UsedCapacityMap != nil && len(nodes) >= 2 {
		// Sort candidate nodes by used capacity in ascending order
		sort.Slice(nodes, func(i, j int) bool {
			usageNodeA := lookUpUsedCapacity(nodes[i], worker2UsedCapacityMap)
			usageNodeB := lookUpUsedCapacity(nodes[j], worker2UsedCapacityMap)
			return usageNodeA < usageNodeB
		})
	}

	return nodes, nil
}
