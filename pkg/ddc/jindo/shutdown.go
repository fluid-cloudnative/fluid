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

package jindo

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/retry"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Shutdown shuts down the Jindo engine
func (e *JindoEngine) Shutdown() (err error) {

	if e.retryShutdown < e.gracefulShutdownLimits {
		err = e.invokeCleanCache()
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed",
				// "engine", e,
				"retry times", e.retryShutdown)
			return
		}
	}

	_, err = e.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = e.releasePorts()
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

// destroyMaster destroys the master
func (e *JindoEngine) destroyMaster() (err error) {
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

func (e *JindoEngine) releasePorts() (err error) {
	var valueConfigMapname = e.name + "-jindofs-config"

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		return errors.Wrap(err, "GetRuntimePortAllocator when releasePorts")
	}

	cm, err := kubeclient.GetConfigmapByName(e.Client, valueConfigMapname, e.namespace)
	if err != nil {
		return errors.Wrap(err, "GetConfigmapByName when releasePorts")
	}

	// The value configMap is not found
	if cm == nil {
		e.Log.Info("value configMap not found, there might be some unreleased ports", "valueConfigMapName", valueConfigMapname)
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
func (e *JindoEngine) cleanAll() (err error) {
	count, err := e.Helper.CleanUpFuse()
	if err != nil {
		e.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	e.Log.Info("clean up fuse count", "n", count)

	err = e.cleanConfigMap()
	if err != nil {
		e.Log.Error(err, "Err in cleaning configMap")
		return err
	}
	return
}

// cleanConfigmap cleans up the configmaps, such as:
// {dataset name}-jindo-values, {dataset name}-jindofs-client-config, {dataset name}-jindofs-config
func (e *JindoEngine) cleanConfigMap() (err error) {
	var (
		valueConfigmapName  = e.getHelmValuesConfigmapName()
		configmapName       = e.name + "-" + RuntimeFSType + "-config"
		clientConfigmapName = e.name + "-" + RuntimeFSType + "-client-config"
		namespace           = e.namespace
	)

	cms := []string{valueConfigmapName, configmapName, clientConfigmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(e.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// destroyWorkers will delete the workers by number of the workers, if workers is -1, it means all the workers are deleted
func (e *JindoEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
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
		return
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
		e.Log.Info("Scale in Jindo workers", "expectedWorkers", expectedWorkers)
		// This is a scale in operation
		nodes, err = e.sortNodesToShutdown(nodeList.Items)
		if err != nil {
			return currentWorkers, err
		}

	} else {
		// Destroy all workers. This is a subprocess during deletion of JindoRuntime
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
				e.Log.Error(err, "Fail to get node", "nodename", nodeName)
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

func (e *JindoEngine) sortNodesToShutdown(candidateNodes []corev1.Node) (nodes []corev1.Node, err error) {
	// If fuses are deployed in global mode. Scaling in workers has nothing to do with fuses.
	// All nodes with related label can be candidate nodes.
	nodes = candidateNodes
	// TODO support jindo calculate node usedCapacity
	return nodes, nil
}
