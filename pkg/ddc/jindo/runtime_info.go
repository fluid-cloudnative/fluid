package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// getRuntimeInfo gets runtime info
func (e *JindoEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}
		e.runtimeInfo = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, runtime.Spec.Tieredstore)
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Info("Failed to get dataset when getruntimeInfo")
		return e.runtimeInfo, err
	}

	e.runtimeInfo.SetupWithDataset(dataset)

	return e.runtimeInfo, nil
}
