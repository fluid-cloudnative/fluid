/*
Copyright 2020 The Fluid Authors.

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
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// AssignNodesToCache finds nodes to place the cache engine
func (e *AlluxioEngine) AssignNodesToCache(desiredNum int32) (currentScheduleNum int32, err error) {

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return currentScheduleNum, err
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return
	}

	e.Log.Info("AssignNodesToCache", "dataset", dataset)
	return datasetSchedule.AssignDatasetToNodes(runtimeInfo,
		dataset,
		e.Client,
		desiredNum)

}

// SyncScheduleInfoToCacheNodes syncs schedule info to nodes
func (e *AlluxioEngine) SyncScheduleInfoToCacheNodes() (err error) {
	defer utils.TimeTrack(time.Now(), "SyncScheduleInfoToCacheNodes", "name", e.name, "namespace", e.namespace)

	var (
		currentCacheNodenames  []string
		previousCacheNodenames []string
	)

	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			e.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			return nil
		}
		return err
	}

	workerSelector, err := labels.Parse(fmt.Sprintf("fluid.io/dataset=%s-%s,app=alluxio,role=alluxio-worker", e.namespace, e.name))
	if err != nil {
		return err
	}

	workerPods, err := kubeclient.GetPodsForStatefulSet(e.Client, workers, workerSelector)
	if err != nil {
		return err
	}

	// find the nodes which should have the runtime label
	for _, pod := range workerPods {
		nodeName := pod.Spec.NodeName
		node := &v1.Node{}
		if err := e.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node); err != nil {
			return err
		}
		// nodesShouldHaveLabel = append(nodesShouldHaveLabel, node)
		currentCacheNodenames = append(currentCacheNodenames, nodeName)
	}

	// find the nodes which already have the runtime label
	previousCacheNodenames, err = e.getAssignedNodes()
	if err != nil {
		return err
	}

	// runtimeLabel indicates the specific runtime pod is on the node
	// e.g. fluid.io/s-alluxio-default-hbase=true
	// runtimeLabel := e.runtimeInfo.GetRuntimeLabelName()
	// runtimeLabel := e.runtimeInfo.GetRuntimeLabelName()

	currentCacheNodenames = utils.RemoveDuplicateStr(currentCacheNodenames)
	previousCacheNodenames = utils.RemoveDuplicateStr(previousCacheNodenames)

	addedCacheNodenames := utils.SubtractString(currentCacheNodenames, previousCacheNodenames)
	removedCacheNodenames := utils.SubtractString(previousCacheNodenames, currentCacheNodenames)

	if len(addedCacheNodenames) > 0 {

		for _, nodeName := range addedCacheNodenames {
			node := v1.Node{}
			err = e.Get(context.TODO(), types.NamespacedName{
				Name: nodeName,
			}, &node)
			if err != nil {
				e.Log.Error(err, "Failed to find new cache node", "node", nodeName)
				return err
			}
			if !datasetSchedule.CheckIfRuntimeInNode(node, e.runtimeInfo) {
				err = datasetSchedule.LabelCacheNode(node, e.runtimeInfo, e.Client)
				if err != nil {
					e.Log.Error(err, "Failed to label new cache node", "node", nodeName)
					return err
				}
			} else {
				e.Log.Info("The node is already added to cache", "node", nodeName)
			}
		}
	}

	if len(removedCacheNodenames) > 0 {
		for _, nodeName := range removedCacheNodenames {
			node := v1.Node{}
			err = e.Get(context.TODO(), types.NamespacedName{
				Name: nodeName,
			}, &node)
			if utils.IgnoreNotFound(err) != nil {
				e.Log.Error(err, "Failed to find new cache node", "node", nodeName)
				return err
			}
			if datasetSchedule.CheckIfRuntimeInNode(node, e.runtimeInfo) {
				err = datasetSchedule.UnlabelCacheNode(node, e.runtimeInfo, e.Client)
				if err != nil {
					e.Log.Error(err, "Failed to unlabel cache node", "node", nodeName)
					return err
				}
			} else {
				e.Log.Info("The node is already removed from cache", "node", nodeName)
			}

		}
	}

	return err
}

// getAssignedNodes gets the node which is already
func (e *AlluxioEngine) getAssignedNodes() (nodeNames []string, err error) {
	var (
		nodeList     = &v1.NodeList{}
		runtimeLabel = e.runtimeInfo.GetRuntimeLabelName()
	)

	nodeNames = []string{}
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", runtimeLabel))
	if err != nil {
		return
	}

	err = e.List(context.TODO(), nodeList, &client.ListOptions{
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
