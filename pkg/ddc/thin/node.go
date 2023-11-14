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
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t ThinEngine) AssignNodesToCache(desiredNum int32) (currentNum int32, err error) {
	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return currentNum, err
	}

	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return
	}

	t.Log.Info("AssignNodesToCache", "dataset", dataset)
	return datasetSchedule.AssignDatasetToNodes(runtimeInfo,
		dataset,
		t.Client,
		desiredNum)
}

func (t ThinEngine) SyncScheduleInfoToCacheNodes() (err error) {
	err = t.syncScheduleInfoToCacheNodes()
	if err != nil {
		return
	}

	updated, err := t.UpdateRuntimeSetConfigIfNeeded()
	t.Log.V(1).Info("UpdateRuntimeSetConfigIfNeeded", "updated", updated)
	return
}

func (t ThinEngine) syncScheduleInfoToCacheNodes() (err error) {
	defer utils.TimeTrack(time.Now(), "syncScheduleInfoToCacheNodes", "name", t.name, "namespace", t.namespace)

	var (
		currentCacheNodenames  []string
		previousCacheNodenames []string
	)

	workers, err := ctrl.GetWorkersAsStatefulset(t.Client,
		types.NamespacedName{Namespace: t.namespace, Name: t.getWorkerName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			t.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			return nil
		}
		return err
	}

	workerSelector, err := labels.Parse(fmt.Sprintf("fluid.io/dataset=%s-%s,app=%s,role=%s", t.namespace, t.name, common.ThinRuntime, workerPodRole))
	if err != nil {
		return err
	}

	workerPods, err := kubeclient.GetPodsForStatefulSet(t.Client, workers, workerSelector)
	if err != nil {
		return err
	}

	// find the nodes which should have the runtime label
	for _, pod := range workerPods {
		nodeName := pod.Spec.NodeName
		node := &v1.Node{}
		if err := t.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node); err != nil {
			return err
		}
		// nodesShouldHaveLabel = append(nodesShouldHaveLabel, node)
		currentCacheNodenames = append(currentCacheNodenames, nodeName)
	}

	// find the nodes which already have the runtime label
	previousCacheNodenames, err = t.getAssignedNodes()
	if err != nil {
		return err
	}

	currentCacheNodenames = utils.RemoveDuplicateStr(currentCacheNodenames)
	previousCacheNodenames = utils.RemoveDuplicateStr(previousCacheNodenames)

	addedCacheNodenames := utils.SubtractString(currentCacheNodenames, previousCacheNodenames)
	removedCacheNodenames := utils.SubtractString(previousCacheNodenames, currentCacheNodenames)

	if len(addedCacheNodenames) > 0 {

		for _, nodeName := range addedCacheNodenames {
			node := v1.Node{}
			err = t.Get(context.TODO(), types.NamespacedName{
				Name: nodeName,
			}, &node)
			if err != nil {
				t.Log.Error(err, "Failed to find new cache node", "node", nodeName)
				return err
			}
			if !datasetSchedule.CheckIfRuntimeInNode(node, t.runtimeInfo) {
				err = datasetSchedule.LabelCacheNode(node, t.runtimeInfo, t.Client)
				if err != nil {
					t.Log.Error(err, "Failed to label new cache node", "node", nodeName)
					return err
				}
			} else {
				t.Log.Info("The node is already added to cache", "node", nodeName)
			}
		}
	}

	if len(removedCacheNodenames) > 0 {
		for _, nodeName := range removedCacheNodenames {
			node := v1.Node{}
			err = t.Get(context.TODO(), types.NamespacedName{
				Name: nodeName,
			}, &node)
			if utils.IgnoreNotFound(err) != nil {
				t.Log.Error(err, "Failed to find new cache node", "node", nodeName)
				return err
			}
			if datasetSchedule.CheckIfRuntimeInNode(node, t.runtimeInfo) {
				err = datasetSchedule.UnlabelCacheNode(node, t.runtimeInfo, t.Client)
				if err != nil {
					t.Log.Error(err, "Failed to unlabel cache node", "node", nodeName)
					return err
				}
			} else {
				t.Log.Info("The node is already removed from cache", "node", nodeName)
			}

		}
	}

	return err
}

// getAssignedNodes gets the node which is already
func (t *ThinEngine) getAssignedNodes() (nodeNames []string, err error) {
	var (
		nodeList     = &v1.NodeList{}
		runtimeLabel = t.runtimeInfo.GetRuntimeLabelName()
	)

	nodeNames = []string{}
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", runtimeLabel))
	if err != nil {
		return
	}

	err = t.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})
	if err != nil {
		return
	}

	for _, node := range nodeList.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	return
}
