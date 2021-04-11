package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	datasetSchedule "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
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

	deprecated, err := e.HasDeprecatedCommonLabelname()
	if err != nil {
		return
	}

	e.runtimeInfo.SetDeprecatedNodeLabel(deprecated)
	return datasetSchedule.AssignDatasetToNodes(runtimeInfo,
		dataset,
		e.Client,
		desiredNum)
}
