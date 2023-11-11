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

package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	v1helper "k8s.io/component-helpers/scheduling/corev1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	// "github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// TODO: move this to some Scheduler-like struct
// SchedulerMutex is a mutex to protect the scheduling process from race condition
var SchedulerMutex = sync.Mutex{}

func AssignDatasetToNodes(runtimeInfo base.RuntimeInfoInterface,
	dataset *datav1alpha1.Dataset,
	runtimeClient client.Client,
	desiredNum int32) (currentScheduleNum int32, err error) {

	// Only one worker can enter this area and the reconciling runtime CR can be scheduled
	SchedulerMutex.Lock()
	defer SchedulerMutex.Unlock()
	defer utils.TimeTrack(time.Now(), "AssignDatasetToNodes", "runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())

	var (
		nodeList              *corev1.NodeList = &corev1.NodeList{}
		alreadySchedNodeList  *corev1.NodeList = &corev1.NodeList{}
		currentScheduledNodes                  = map[string]corev1.Node{}
		newScheduledNodes                      = []corev1.Node{}
		newScheduleNum        int32
		log                   = rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())
	)

	// 1. get all nodes in the cluster
	err = runtimeClient.List(context.TODO(), nodeList, &client.ListOptions{})
	if err != nil {
		return
	}

	// 2. filters scheduled nodes and build a map for future use
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", runtimeInfo.GetCommonLabelName()))
	if err != nil {
		return currentScheduleNum, err
	}
	err = runtimeClient.List(context.TODO(), alreadySchedNodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})
	if err != nil {
		if !errors.IsNotFound(err) {
			return
		}
	}

	for _, node := range alreadySchedNodeList.Items {
		currentScheduledNodes[node.Name] = node
		log.Info("Node is already assigned", "node", node.Name, "dataset", dataset.Name)
	}

	// 3. Sort nodes if in fuse global mode
	fuseGlobal, nodeSelector := runtimeInfo.GetFuseDeployMode()

	pvcMountNodesMap, err := kubeclient.GetPvcMountNodes(runtimeClient, dataset.Name, dataset.Namespace)
	if err != nil {
		log.Error(err, "Failed to get PVC Mount Nodes, will treat every node as no PVC mount Pods")
	}

	var nodes []corev1.Node

	if fuseGlobal {
		nodes = sortNodesToBeScheduled(nodeList.Items, pvcMountNodesMap, nodeSelector)
	} else {
		nodes = nodeList.Items
	}

	// 4. filter candidate nodes
	for _, node := range nodes {

		if int32(len(currentScheduledNodes)) == desiredNum {
			break
		}

		if _, found := currentScheduledNodes[node.Name]; found {
			log.Info("Node is skipped because it is already assigned", "node", node.Name)
			continue
		}

		if dataset.Spec.NodeAffinity != nil {
			if dataset.Spec.NodeAffinity.Required != nil {
				terms := dataset.Spec.NodeAffinity.Required.NodeSelectorTerms
				matched, err := v1helper.MatchNodeSelectorTerms(&node, &corev1.NodeSelector{NodeSelectorTerms: terms})
				if err != nil {
					log.Error(err, "Node is skipped because of error", "node", node.Name)
					continue
				}
				if !matched {
					log.Info("Node is skipped because it can't meet node selector terms", "node", node.Name)
					continue
				}
			}
		}

		if !kubeclient.IsReady(node) {
			log.Info("Node is skipped because it is not ready", "node", node.Name)
			continue
		}

		if node.Spec.Unschedulable {
			log.Info("Node is skipped because it is unschedulable", "node", node.Name)
			continue
		}

		if len(node.Spec.Taints) > 0 {
			tolerateEffect := toleratesTaints(node.Spec.Taints, dataset.Spec.Tolerations)
			if tolerateEffect {
				log.Info("The tainted node also can be scheduled because of toleration effects",
					"node", node.Name,
					"taints", node.Spec.Taints,
					"tolerations", dataset.Spec.Tolerations)
			} else {
				log.Info("Skip the node because it's tainted", "node", node.Name)
				continue
			}

		}

		if !AlreadyAssigned(runtimeInfo, node) {
			if !CanbeAssigned(runtimeInfo, node) {
				log.Info("Node is skipped because it is not assigned and also can't be assigned", "node", node.Name)
				continue
			} else {
				newScheduledNodes = append(newScheduledNodes, node)
				log.Info("New Node to schedule",
					"dataset", runtimeInfo.GetName(),
					"node", node.Name)
			}
		} else {
			log.Info("Node is already scheduled for dataset",
				"dataset", runtimeInfo.GetName(),
				"node", node.Name)
		}

		currentScheduledNodes[node.Name] = node
	}

	currentScheduleNum = int32(len(currentScheduledNodes))
	newScheduleNum = int32(len(newScheduledNodes))
	log.Info("Find node to schedule or scheduled for dataset",
		"dataset", runtimeInfo.GetName(),
		"currentScheduleNum", currentScheduleNum,
		"newScheduleNum", newScheduleNum)

	// 5. bind the dataset to selected nodes via adding corresponding labels on them
	for _, node := range newScheduledNodes {
		err = LabelCacheNode(node, runtimeInfo, runtimeClient)
		if err != nil {
			return
		}
	}

	return
}

// sortNodesToBeScheduled sorts nodes to be scheduled when scale up
func sortNodesToBeScheduled(nodes []corev1.Node, pvcMountNodesMap map[string]int64, nodeSelector map[string]string) []corev1.Node {
	var (
		// There are three slices which have different priorities
		// 1. nodes which have PVC mount Pods on it now
		pvcMountNodes []corev1.Node
		// 2. nodes without PVC mount Pods on it but are selected by fuse, perhaps will have them in future
		selectedNodes []corev1.Node
		// 3. nodes not selected by fuse, will never have PVC mount Pods
		notSelectedNodes []corev1.Node
	)

	for _, node := range nodes {
		if num, found := pvcMountNodesMap[node.Name]; found {
			if len(pvcMountNodes) == 0 {
				pvcMountNodes = append(pvcMountNodes, node)
			} else {
				// Binary Insertion
				low := 0
				high := len(pvcMountNodes) - 1
				for low <= high {
					middle := (low + high) / 2
					if num > pvcMountNodesMap[pvcMountNodes[middle].Name] {
						high = middle - 1
					} else {
						low = middle + 1
					}
				}
				k := len(pvcMountNodes) - 1
				pvcMountNodes = append(pvcMountNodes, pvcMountNodes[k])
				for k >= low {
					pvcMountNodes[k+1] = pvcMountNodes[k]
					k = k - 1
				}
				pvcMountNodes[low] = node
			}
		} else if utils.ContainsSelector(node.GetLabels(), nodeSelector) {
			selectedNodes = append(selectedNodes, node)
		} else {
			notSelectedNodes = append(notSelectedNodes, node)
		}
	}
	return append(append(pvcMountNodes, selectedNodes...), notSelectedNodes...)
}
