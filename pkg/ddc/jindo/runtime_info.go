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
		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, runtime.Spec.Tieredstore)
		if err != nil {
			return e.runtimeInfo, err
		}
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("Dataset is notfound", "name", e.name, "namespace", e.namespace)
			return e.runtimeInfo, nil
		}

		e.Log.Info("Failed to get dataset when getruntimeInfo")
		return e.runtimeInfo, err
	}

	e.runtimeInfo.SetupWithDataset(dataset)

	return e.runtimeInfo, nil
}
