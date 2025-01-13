/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t ThinEngine) Shutdown() (err error) {
	if t.retryShutdown < t.gracefulShutdownLimits {
		err = t.cleanupCache()
		if err != nil {
			t.retryShutdown = t.retryShutdown + 1
			t.Log.Info("clean cache failed",
				"retry times", t.retryShutdown)
			return
		}
	}

	_, err = t.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = t.destroyMaster()
	if err != nil {
		return
	}

	err = t.cleanAll()
	return
}

// destroyMaster Destroy the master
func (t *ThinEngine) destroyMaster() (err error) {
	var found bool
	found, err = helm.CheckRelease(t.name, t.namespace)
	if err != nil {
		return err
	}

	if found {
		err = helm.DeleteRelease(t.name, t.namespace)
		if err != nil {
			return
		}
	} else {
		// When upgrade Fluid to v1.0.0+ from a lower version, there may be some orphaned configmaps when deleting a ThinRuntime if it's created before the upgradation.
		// Detect such orphaned configmaps and clean them up.
		err = t.cleanUpOrphanedResources()
		if err != nil {
			t.Log.Info("WARNING: failed to delete orphaned resource, some resources may not be cleaned up in the cluster", "err", err)
			err = nil
		}
	}

	return
}

// cleanupCache cleans up the cache
func (t *ThinEngine) cleanupCache() (err error) {
	// todo
	return
}

// destroyWorkers attempts to delete the workers until worker num reaches the given expectedWorkers, if expectedWorkers is -1, it means all the workers should be deleted
// This func returns currentWorkers representing how many workers are left after this process.
func (t *ThinEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := t.getRuntimeInfo()
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
	t.Log.Info("check node labels", "labelNames", labelNames)

	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", labelCommonName))
	if err != nil {
		return currentWorkers, err
	}

	err = t.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})

	if err != nil {
		return currentWorkers, err
	}

	currentWorkers = int32(len(nodeList.Items))
	if expectedWorkers >= currentWorkers {
		t.Log.Info("No need to scale in. Skip.")
		return currentWorkers, nil
	}

	var nodes []corev1.Node
	if expectedWorkers >= 0 {
		t.Log.Info("Scale in thinfs workers", "expectedWorkers", expectedWorkers)
		// This is a scale in operation
		nodes, err = t.sortNodesToShutdown(nodeList.Items)
		if err != nil {
			return currentWorkers, err
		}

	} else {
		// Destroy all workers. This is a subprocess during deletion of ThinRuntime
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
			node, err := kubeclient.GetNode(t.Client, nodeName)
			if err != nil {
				t.Log.Error(err, "Fail to get node", "nodename", nodeName)
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
			modifiedLabels, err := utils.ChangeNodeLabelWithPatchMode(t.Client, toUpdate, labelsToModify)
			if err != nil {
				return err
			}
			t.Log.Info("Destroy worker", "Dataset", t.name, "deleted worker node", node.Name, "removed or updated labels", modifiedLabels)
			return nil
		})

		if err != nil {
			return currentWorkers, err
		}

		currentWorkers--
	}

	return currentWorkers, nil
}

func (t *ThinEngine) sortNodesToShutdown(candidateNodes []corev1.Node) (nodes []corev1.Node, err error) {
	// If fuses are deployed in global mode. Scaling in workers has nothing to do with fuses.
	// All nodes with related label can be candidate nodes.
	nodes = candidateNodes

	// Prefer to choose nodes with less data cache
	//Todo

	return nodes, nil
}

func (t *ThinEngine) cleanAll() (err error) {
	count, err := t.Helper.CleanUpFuse()
	if err != nil {
		t.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	t.Log.Info("clean up fuse count", "n", count)

	var (
		valueConfigmapName = t.getHelmValuesConfigMapName()
		configmapName      = t.name + "-config"
		namespace          = t.namespace
	)

	cms := []string{valueConfigmapName, configmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(t.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

func (t *ThinEngine) cleanUpOrphanedResources() (err error) {
	orphanedConfigMapName := fmt.Sprintf("%s-runtimeset", t.name)
	cm, err := kubeclient.GetConfigmapByName(t.Client, orphanedConfigMapName, t.namespace)
	if err != nil {
		return err
	}

	if cm != nil {
		if err = kubeclient.DeleteConfigMap(t.Client, orphanedConfigMapName, t.namespace); err != nil && utils.IgnoreNotFound(err) != nil {
			return err
		}
		t.Log.Info("Found orphaned configmap, successfully deleted it", "configmap", orphanedConfigMapName)
	}

	return nil
}
