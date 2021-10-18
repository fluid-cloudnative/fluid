package jindo

import (
	"context"
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (e *JindoEngine) AssignNodesToCache(desiredNum int32) (currentScheduleNum int32, err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return currentScheduleNum, err
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.Info("AssignNodesToCache", "dataset", dataset)
	if err != nil {
		return
	}

	return datasetSchedule.AssignDatasetToNodes(runtimeInfo,
		dataset,
		e.Client,
		desiredNum)
}

// SyncScheduleInfoToCacheNodes syncs the cache info of the nodes by labeling the nodes
// And the Application pod can leverage such info for scheduling
func (e *JindoEngine) SyncScheduleInfoToCacheNodes() (err error) {
	defer utils.TimeTrack(time.Now(), "SyncScheduleInfoToCacheNodes")

	var (
		currentCacheNodenames  []string
		previousCacheNodenames []string
	)

	workers, err := e.getStatefulset(e.getWorkertName(), e.namespace)
	if err != nil {
		return err
	}

	workerSelector, err := labels.Parse(fmt.Sprintf("fluid.io/dataset=%s-%s,app=jindofs,role=jindofs-worker", e.namespace, e.name))
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

			err = datasetSchedule.LabelCacheNode(node, e.runtimeInfo, e.Client)
			if err != nil {
				e.Log.Error(err, "Failed to label new cache node", "node", nodeName)
				return err
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
		}
	}

	return err
}

// getAssignedNodes gets the node which is already
func (e *JindoEngine) getAssignedNodes() (nodeNames []string, err error) {
	var (
		nodeList     = &corev1.NodeList{}
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
