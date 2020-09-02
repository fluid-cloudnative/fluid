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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

// AssignNodesToCache finds nodes
func (e *AlluxioEngine) AssignNodesToCache(desiredNum int32) (currentScheduleNum int32, err error) {
	var (
		nodeList              *corev1.NodeList = &corev1.NodeList{}
		alreadySchedNodeList  *corev1.NodeList = &corev1.NodeList{}
		currentScheduledNodes                  = map[string]corev1.Node{}
		newScheduledNodes                      = []corev1.Node{}
		newScheduleNum        int32
		dataset               *datav1alpha1.Dataset
	)

	err = e.List(context.TODO(), nodeList, &client.ListOptions{})
	if err != nil {
		return
	}

	runtime, err := e.getRuntime()
	if err != nil {
		return currentScheduleNum, err
	}

	dataset, err = utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.Info("AssignNodesToCache", "dataset", dataset)
	if err != nil {
		return
	}

	// datasetLabels := labels.SelectorFromSet(labels.Set(map[string]string{e.getCommonLabelname(): "true"}))
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", e.getCommonLabelname()))
	if err != nil {
		return currentScheduleNum, err
	}
	err = e.List(context.TODO(), alreadySchedNodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})
	if err != nil {
		if !errors.IsNotFound(err) {
			return
		}
	}

	for _, node := range alreadySchedNodeList.Items {
		currentScheduledNodes[node.Name] = node
		e.Log.Info("Node is already assigned", "node", node.Name, "dataset", dataset.Name)
	}

	// storageMap := tieredstore.GetLevelStorageMap(runtime)
	for _, node := range nodeList.Items {

		if int32(len(currentScheduledNodes)) == desiredNum {
			break
		}

		if _, found := currentScheduledNodes[node.Name]; found {
			e.Log.Info("Node is skipped because it is already assigned", "node", node.Name)
			continue
		}

		// if runtime.Spec.Placement.All().NodeAffinity != nil {
		// 	terms := runtime.Spec.Placement.All().NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		// 	if !v1helper.MatchNodeSelectorTerms(terms, labels.Set(node.Labels), nil) {
		// 		e.Log.Info("Node is skipped because it can't meet node selector terms", "node", node.Name)
		// 		continue
		// 	}
		// }
		if dataset.Spec.NodeAffinity != nil {
			if dataset.Spec.NodeAffinity.Required != nil {
				terms := dataset.Spec.NodeAffinity.Required.NodeSelectorTerms
				if !v1helper.MatchNodeSelectorTerms(terms, labels.Set(node.Labels), nil) {
					e.Log.Info("Node is skipped because it can't meet node selector terms", "node", node.Name)
					continue
				}
			}
		}

		if !kubeclient.IsReady(node) {
			e.Log.Info("Node is skipped because it is not ready", "node", node.Name)
			continue
		}

		if len(node.Spec.Taints) > 0 {
			e.Log.Info("Skip the node because it's tainted", "node", node.Name)
			continue
		}

		if !e.alreadyAssigned(runtime, node) {
			if !e.canbeAssigned(runtime, node) {
				e.Log.Info("Node is skipped because it is not assigned and also can't be assigned", "node", node.Name)
				continue
			} else {
				newScheduledNodes = append(newScheduledNodes, node)
				e.Log.Info("New Node to schedule",
					"dataset", e.name,
					"node", node.Name)
			}
		} else {
			e.Log.Info("Node is already scheduled for dataset",
				"dataset", e.name,
				"node", node.Name)
		}

		currentScheduledNodes[node.Name] = node
	}

	currentScheduleNum = int32(len(currentScheduledNodes))
	newScheduleNum = int32(len(newScheduledNodes))
	e.Log.Info("Find node to schedule or scheduled for dataset",
		"dataset", e.name,
		"currentScheduleNum", currentScheduleNum,
		"newScheduleNum", newScheduleNum)
	// 2.Add label to the selected node

	for _, node := range newScheduledNodes {
		err = e.labelCacheNode(node, runtime)
		if err != nil {
			return
		}
	}

	return

}

// alreadyAssigned checks if the node is already assigned the runtime engine
func (e *AlluxioEngine) alreadyAssigned(runtime *datav1alpha1.AlluxioRuntime, node corev1.Node) (assigned bool) {
	label := e.getCommonLabelname()

	if len(node.Labels) > 0 {
		_, assigned = node.Labels[label]
	}

	e.Log.Info("Check alreadyAssigned", "node", node.Name, "label", label, "assigned", assigned)

	return

}

// alreadyAssignedByFluid checks if the node is occupied by other dataset
func (e *AlluxioEngine) alreadyAssignedByFluid(node corev1.Node) (assigned bool) {
	labels := node.Labels
	if len(labels) > 0 {
		for label := range labels {
			if strings.HasPrefix(label, common.LabelAnnotationStorageCapacityPrefix) {
				assigned = true
				e.Log.Info("alreadyAssignedByFluid find the node is already used by dataset.",
					"node", node.Name,
					"label", label)
				break
			}
		}
	}

	return
}

// canbeAssigned checks if the node is already assigned the runtime engine
func (e *AlluxioEngine) canbeAssigned(runtime *datav1alpha1.AlluxioRuntime, node corev1.Node) bool {
	// TODO(cheyang): the different dataset can be put in the same node, but it has to handle port conflict
	if e.alreadyAssignedByFluid(node) {
		return false
	}

	storageMap := tieredstore.GetLevelStorageMap(runtime)

	for key, requirement := range storageMap {
		if key == common.MemoryCacheStore {
			nodeMemoryCapacity := *node.Status.Allocatable.Memory()
			if requirement.Cmp(nodeMemoryCapacity) <= 0 {
				e.Log.Info("requirement is less than node memory capacity", "requirement", requirement,
					"nodeMemoryCapacity", nodeMemoryCapacity)
			} else {
				e.Log.Info("requirement is more than node memory capacity", "requirement", requirement,
					"nodeMemoryCapacity", nodeMemoryCapacity)
				return false
			}
		}

		// } else {
		// 	nodeDiskCapacity := *node.Status.Allocatable.StorageEphemeral()
		// 	if requirement.Cmp(nodeDiskCapacity) <= 0 {
		// 		e.Log.Info("requirement is less than node disk capacity", "requirement", requirement,
		// 			"nodeDiskCapacity", nodeDiskCapacity)
		// 	} else {
		// 		e.Log.Info("requirement is more than node disk capacity", "requirement", requirement,
		// 			"nodeDiskCapacity", nodeDiskCapacity)
		// 		return false
		// 	}
		// }
	}

	return true

}

func (e *AlluxioEngine) labelCacheNode(nodeToLabel corev1.Node, runtime *datav1alpha1.AlluxioRuntime) (err error) {
	var (
		labelName       = e.getRuntimeLabelname()
		labelCommonName = e.getCommonLabelname()
	)

	storageMap := tieredstore.GetLevelStorageMap(runtime)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		nodeName := nodeToLabel.Name
		node, err := kubeclient.GetNode(e.Client, nodeName)
		toUpdate := node.DeepCopy()
		if toUpdate.Labels == nil {
			toUpdate.Labels = make(map[string]string)
		}

		toUpdate.Labels[labelName] = "true"
		toUpdate.Labels[labelCommonName] = "true"
		totalRequirement, err := resource.ParseQuantity("0Gi")
		if err != nil {
			e.Log.Error(err, "Failed to parse the total requirement")
		}
		for key, requirement := range storageMap {
			value := tranformQuantityToUnits(requirement)
			if key == common.MemoryCacheStore {
				toUpdate.Labels[e.getStoragetLabelname(humanReadType, memoryStorageType)] = value
			} else {
				toUpdate.Labels[e.getStoragetLabelname(humanReadType, diskStorageType)] = value
			}
			totalRequirement.Add(*requirement)
		}
		totalValue := tranformQuantityToUnits(&totalRequirement)
		toUpdate.Labels[e.getStoragetLabelname(humanReadType, totalStorageType)] = totalValue

		// toUpdate.Labels[labelNameToAdd] = "true"
		err = e.Client.Update(context.TODO(), toUpdate)
		if err != nil {
			e.Log.Error(err, "LabelCachedNodes")
			return err
		}
		return nil
	})

	if err != nil {
		e.Log.Error(err, "LabelCacheNode")
		return err
	}

	return nil
}
