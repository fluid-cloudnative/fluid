package lifecycle

import (
	"context"
	"fmt"
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

func AssignDatasetToNodes(runtimeInfo base.RuntimeInfoInterface,
	dataset *datav1alpha1.Dataset,
	runtimeClient client.Client,
	desiredNum int32) (currentScheduleNum int32, err error) {
	var (
		nodeList              *corev1.NodeList = &corev1.NodeList{}
		alreadySchedNodeList  *corev1.NodeList = &corev1.NodeList{}
		currentScheduledNodes                  = map[string]corev1.Node{}
		newScheduledNodes                      = []corev1.Node{}
		newScheduleNum        int32
		log                   = rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())
	)

	err = runtimeClient.List(context.TODO(), nodeList, &client.ListOptions{})
	if err != nil {
		return
	}

	// datasetLabels := labels.SelectorFromSet(labels.Set(map[string]string{e.getCommonLabelname(): "true"}))
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

	for _, node := range alreadySchedNodeList.Items {
		currentScheduledNodes[node.Name] = node
		log.Info("Node is already assigned", "node", node.Name, "dataset", dataset.Name)
	}

	fuseGlobal, nodeSelector := runtimeInfo.GetFuseDeployMode()

	pvcMountNodesMap, err := kubeclient.GetPvcMountNodes(runtimeClient, dataset.Name, dataset.Namespace)
	if err != nil {
		log.Error(err, "Failed to get PVC Mount Nodes, will treat every node as with no PVC mount Pods")
	}

	var nodes []corev1.Node

	if fuseGlobal {
		nodes = sortNodesToBeScheduled(nodeList.Items, pvcMountNodesMap, nodeSelector)
	} else {
		nodes = nodeList.Items
	}

	// storageMap := tieredstore.GetLevelStorageMap(runtime)
	for _, node := range nodes {

		if int32(len(currentScheduledNodes)) == desiredNum {
			break
		}

		if _, found := currentScheduledNodes[node.Name]; found {
			log.Info("Node is skipped because it is already assigned", "node", node.Name)
			continue
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
	// 2.Add label to the selected node

	for _, node := range newScheduledNodes {
		err = LabelCacheNode(node, runtimeInfo, runtimeClient)
		if err != nil {
			return
		}
	}

	return
}

// sortNodesToBeScheduled will sort nodes to be scheduled when scale up
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
