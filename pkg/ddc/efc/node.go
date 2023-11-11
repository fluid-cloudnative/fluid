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

package efc

import (
	"context"
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AssignNodesToCache finds nodes to place the cache engine
func (e *EFCEngine) AssignNodesToCache(desiredNum int32) (currentScheduleNum int32, err error) {
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
func (e *EFCEngine) SyncScheduleInfoToCacheNodes() (err error) {
	defer utils.TimeTrack(time.Now(), "SyncScheduleInfoToCacheNodes", "name", e.name, "namespace", e.namespace)

	var (
		currentCacheNodenames  []string
		previousCacheNodenames []string
	)

	workerPods, err := e.getWorkerRunningPods()
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
		currentCacheNodenames = append(currentCacheNodenames, nodeName)
	}

	// find the nodes which already have the runtime label
	previousCacheNodenames, err = e.getAssignedNodes()
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
func (e *EFCEngine) getAssignedNodes() (nodeNames []string, err error) {
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
