package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
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

		if runtime.Spec.Fuse.Global {
			e.runtimeInfo.SetupFuseDeployMode(runtime.Spec.Fuse.Global, runtime.Spec.Fuse.NodeSelector)
			e.Log.Info("Enable global mode for fuse")
		} else {
			e.Log.Info("Disable global mode for fuse")
		}

		// Check if the runtime is using deprecated labels
		isLabelDeprecated, err := e.HasDeprecatedCommonLabelname()
		if err != nil {
			return e.runtimeInfo, err
		}
		e.runtimeInfo.SetDeprecatedNodeLabel(isLabelDeprecated)

		// Check if the runtime is using deprecated naming style for PersistentVolumes
		isPVNameDeprecated, err := volume.HasDeprecatedPersistentVolumeName(e.Client, e.runtimeInfo, e.Log)
		if err != nil {
			return e.runtimeInfo, err
		}
		e.runtimeInfo.SetDeprecatedPVName(isPVNameDeprecated)

		e.Log.Info("Deprecation check finished", "isLabelDeprecated", e.runtimeInfo.IsDeprecatedNodeLabel(), "isPVNameDeprecated", e.runtimeInfo.IsDeprecatedPVName())

		// Setup with Dataset Info
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

		e.Log.Info("Setup with dataset done", "exclusive", e.runtimeInfo.IsExclusive())
	}

	return e.runtimeInfo, nil
}
