package jindo

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	podList := &v1.PodList{}
	workerLabels, err := labels.Parse(fmt.Sprintf("fluid.io/dataset=%s-%s,app=jindofs,role=jindofs-worker", e.namespace, e.name))
	if err != nil {
		return err
	}

	err = e.List(context.TODO(), podList, &client.ListOptions{
		LabelSelector: workerLabels,
	})

	return err
}
