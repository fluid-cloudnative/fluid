package jindo

import (
	"context"
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

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
		nodesShouldHaveLabel  []*v1.Node
		nodesAlreadyHaveLabel []*v1.Node
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
		nodesShouldHaveLabel = append(nodesShouldHaveLabel, node)
	}

	// find the nodes which already have the runtime label

	// runtimeLabel indicates the specific runtime pod is on the node
	// e.g. fluid.io/s-alluxio-default-hbase=true
	// runtimeLabel := e.runtimeInfo.GetRuntimeLabelName()
	// runtimeLabel := e.runtimeInfo.GetRuntimeLabelName()

	return err
}
