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

package lifecycle

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"
	"sync"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	// "github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// TODO: move this to some Scheduler-like struct
// schedulerMutex is a mutex to protect the scheduling process from race condition
var schedulerMutex = sync.Mutex{}

// AssignDatasetToNodes schedules datasets by assigning runtime pod to nodes.
// This is a thread-safe function to make sure there is only one dataset in scheduling at any time.
func AssignDatasetToNodes(runtimeInfo base.RuntimeInfoInterface,
	dataset *datav1alpha1.Dataset,
	runtimeClient client.Client,
	desiredNum int32) (currentScheduleNum int32, err error) {

	// Only one worker can enter this area and the reconciling runtime CR can be scheduled
	schedulerMutex.Lock()
	defer schedulerMutex.Unlock()

	var (
		nodeList              *corev1.NodeList = &corev1.NodeList{}
		alreadySchedNodeList  *corev1.NodeList = &corev1.NodeList{}
		currentScheduledNodes                  = map[string]corev1.Node{}
		newScheduledNodes                      = []corev1.Node{}
		newScheduleNum        int32
		log                   = rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())
	)

	// 1. Get snapshots of n lists for scheduling
	err = runtimeClient.List(context.TODO(), nodeList, &client.ListOptions{})
	if err != nil {
		return
	}

	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", runtimeInfo.GetCommonLabelname()))
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

	// 2. When in fuse global mode, sort nodes by number of running workloads on it.
	fuseGlobal, nodeSelector := runtimeInfo.GetFuseDeployMode()
	var nodes []corev1.Node

	if fuseGlobal {
		pvcMountNodesMap, err := kubeclient.GetPvcMountNodes(runtimeClient, dataset.Name, dataset.Namespace)
		if err != nil {
			log.Error(err, "Failed to get PVC Mount Nodes, will treat every n as no PVC mount Pods")
		}
		nodes = sortNodesToBeScheduled(nodeList.Items, pvcMountNodesMap, nodeSelector)
	} else {
		nodes = nodeList.Items
	}

	// 3. Pick nodes that is already scheduled with this runtime worker pod
	for _, node := range alreadySchedNodeList.Items {
		currentScheduledNodes[node.Name] = node
		log.Info("Node is already assigned", "n", node.Name, "dataset", dataset.Name)
	}

	// 4. Filter and label nodes until desired num of replicas is reached
	for _, n := range nodes {

		if int32(len(currentScheduledNodes)) == desiredNum {
			break
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			nodeName := n.Name
			node, err := kubeclient.GetNode(runtimeClient, nodeName)
			if err != nil {
				log.Error(err, "GetNode in AssignDatasetToNodes", "nodeName", nodeName)
				return err
			}

			// No need to label the node if it fails the node filter
			if !filterNode(*node, currentScheduledNodes, dataset, runtimeInfo, log) {
				return nil
			}

			if err = LabelCacheNode(*node, runtimeInfo, runtimeClient); err != nil {
				return err
			}

			// The candidate node is successfully labeled
			log.Info("New Node to schedule",
				"dataset", runtimeInfo.GetName(),
				"node", n.Name)
			newScheduledNodes = append(newScheduledNodes, n)
			currentScheduledNodes[n.Name] = n
			return nil
		})

		if err != nil {
			log.Error(err, "Scheduling node in AssignDatasetToNodes", "nodeName", n.Name)
			return int32(len(currentScheduledNodes)), err
		}
	}

	currentScheduleNum = int32(len(currentScheduledNodes))
	newScheduleNum = int32(len(newScheduledNodes))
	log.Info("node scheduled for dataset",
		"dataset", runtimeInfo.GetName(),
		"currentScheduleNum", currentScheduleNum,
		"newScheduleNum", newScheduleNum)

	return
}

// filterNode checks if the given runtime can be scheduled on the given node
func filterNode(node corev1.Node,
	currentScheduledNodes map[string]corev1.Node,
	dataset *datav1alpha1.Dataset,
	runtimeInfo base.RuntimeInfoInterface,
	log logr.Logger) bool {

	if _, found := currentScheduledNodes[node.Name]; found {
		log.Info("Node is filtered out because it is already assigned", "node", node.Name)
		return false
	}

	// if runtime.Spec.Placement.All().NodeAffinity != nil {
	// 	terms := runtime.Spec.Placement.All().NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	// 	if !v1helper.MatchNodeSelectorTerms(terms, labels.Set(node.Labels), nil) {
	// 		log.Info("Node is skipped because it can't meet node selector terms", "node", node.Name)
	// 		continue
	// 	}
	// }
	if dataset.Spec.NodeAffinity != nil {
		if dataset.Spec.NodeAffinity.Required != nil {
			terms := dataset.Spec.NodeAffinity.Required.NodeSelectorTerms
			if !v1helper.MatchNodeSelectorTerms(terms, labels.Set(node.Labels), nil) {
				log.Info("Node is filtered out because it can't meet node selector terms", "node", node.Name)
				return false
			}
		}
	}

	if !kubeclient.IsReady(node) {
		log.Info("Node is filtered out because it is not ready", "node", node.Name)
		return false
	}

	if node.Spec.Unschedulable {
		log.Info("Node is filtered out because it is unschedulable", "node", node.Name)
		return false
	}

	if len(node.Spec.Taints) > 0 {
		tolerateEffect := toleratesTaints(node.Spec.Taints, dataset.Spec.Tolerations)
		if tolerateEffect {
			log.Info("The tainted node also can be scheduled because of toleration effects",
				"node", node.Name,
				"taints", node.Spec.Taints,
				"tolerations", dataset.Spec.Tolerations)
		} else {
			log.Info("Node is filtered out because it's tainted", "node", node.Name)
			return false
		}
	}

	if !AlreadyAssigned(runtimeInfo, node) {
		if !CanbeAssigned(runtimeInfo, node) {
			log.Info("Node is filtered out because it is not assigned and also can't be assigned", "node", node.Name)
			return false
		}
	} else {
		log.Info("Node is filtered out because it's already scheduled for dataset",
			"dataset", runtimeInfo.GetName(),
			"node", node.Name)
		return false
	}

	//newScheduledNodes = append(newScheduledNodes, node)
	log.Info("New Node to schedule",
		"dataset", runtimeInfo.GetName(),
		"node", node.Name)
	return true
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
