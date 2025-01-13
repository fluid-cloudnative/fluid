/*
Copyright 2020 The Fluid Author.

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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"

	"sort"
	"strings"

	"k8s.io/client-go/util/retry"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// shut down the Alluxio engine
func (e *AlluxioEngine) Shutdown() (err error) {
	gracefulShutdownLimits, err := e.getGracefulShutdownLimits()
	if err != nil {
		return
	}
	if e.retryShutdown < gracefulShutdownLimits {
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
		base.SafeClose(e.MetadataSyncDoneCh)
	}

	_, err = e.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = e.destroyMaster()
	if err != nil {
		return
	}

	// There is no need to release the ports in container network mode
	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}
	if datav1alpha1.IsHostNetwork(runtime.Spec.Master.NetworkMode) ||
		datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode) {
		e.Log.Info("releasePorts for hostnetwork mode")
		err = e.releasePorts()
		if err != nil {
			return
		}
	} else {
		e.Log.Info("skip releasePorts for container network mode")
	}

	return e.cleanAll()
}

// destroyMaster Destroys the master
func (e *AlluxioEngine) destroyMaster() (err error) {
	var found bool
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
	if utils.IgnoreNotFound(err) != nil {
		return err
	}
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

	// time.Sleep(time.Duration(5 * time.Second))
	return fmt.Errorf("to make sure if the remaining cache is cleaned up, check again")
}

func (e *AlluxioEngine) releasePorts() (err error) {
	var valueConfigMapName = e.getHelmValuesConfigMapName()

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		return errors.Wrap(err, "GetRuntimePortAllocator when releasePorts")
	}

	cm, err := kubeclient.GetConfigmapByName(e.Client, valueConfigMapName, e.namespace)
	if err != nil {
		return errors.Wrap(err, "GetConfigmapByName when releasePorts")
	}

	// The value configMap is not found
	if cm == nil {
		e.Log.Info("value configMap not found, there might be some unreleased ports", "valueConfigMapName", valueConfigMapName)
		return nil
	}

	portsToRelease, err := parsePortsFromConfigMap(cm)
	if err != nil {
		return errors.Wrap(err, "parsePortsFromConfigMap when releasePorts")
	}

	allocator.ReleaseReservedPorts(portsToRelease)
	return nil
}

// cleanAll cleans up the all
func (e *AlluxioEngine) cleanAll() (err error) {
	count, err := e.Helper.CleanUpFuse()
	if err != nil {
		e.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	e.Log.Info("clean up fuse count", "n", count)

	var (
		valueConfigmapName = e.getHelmValuesConfigMapName()
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
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return currentWorkers, err
	}

	var (
		nodeList           = &corev1.NodeList{}
		labelExclusiveName = utils.GetExclusiveKey()
		labelName          = runtimeInfo.GetRuntimeLabelName()
		labelCommonName    = runtimeInfo.GetCommonLabelName()
		labelMemoryName    = runtimeInfo.GetLabelNameForMemory()
		labelDiskName      = runtimeInfo.GetLabelNameForDisk()
		labelTotalName     = runtimeInfo.GetLabelNameForTotal()
	)

	labelNames := []string{labelName, labelTotalName, labelDiskName, labelMemoryName, labelCommonName}
	e.Log.Info("check node labels", "labelNames", labelNames)

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
		e.Log.Info("Scale in Alluxio workers", "expectedWorkers", expectedWorkers)
		// This is a scale in operation
		nodes, err = e.sortNodesToShutdown(nodeList.Items)
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

		if len(node.Labels) == 0 {
			continue
		}

		nodeName := node.Name
		var labelsToModify common.LabelsToModify
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			node, err := kubeclient.GetNode(e.Client, nodeName)
			if err != nil {
				e.Log.Error(err, "Fail to get node", "nodeName", nodeName)
				return err
			}

			toUpdate := node.DeepCopy()
			for _, label := range labelNames {
				labelsToModify.Delete(label)
			}

			exclusiveLabelValue := runtimeInfo.GetExclusiveLabelValue()
			if val, exist := toUpdate.Labels[labelExclusiveName]; exist && val == exclusiveLabelValue {
				labelsToModify.Delete(labelExclusiveName)
			}

			err = lifecycle.DecreaseDatasetNum(toUpdate, runtimeInfo, &labelsToModify)
			if err != nil {
				return err
			}
			// Update the toUpdate in UPDATE mode
			// modifiedLabels, err := utils.ChangeNodeLabelWithUpdateMode(e.Client, toUpdate, labelToModify)
			// Update the toUpdate in PATCH mode
			modifiedLabels, err := utils.ChangeNodeLabelWithPatchMode(e.Client, toUpdate, labelsToModify)
			if err != nil {
				return err
			}
			e.Log.Info("Destroy worker", "Dataset", e.name, "deleted worker node", node.Name, "removed or updated labels", modifiedLabels)
			return nil
		})

		if err != nil {
			return currentWorkers, err
		}

		currentWorkers--
	}

	return currentWorkers, nil
}

func (e *AlluxioEngine) sortNodesToShutdown(candidateNodes []corev1.Node) (nodes []corev1.Node, err error) {
	// If fuses are deployed in global mode. Scaling in workers has nothing to do with fuses.
	// All nodes with related label can be candidate nodes.
	nodes = candidateNodes

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
