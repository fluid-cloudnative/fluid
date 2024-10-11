package alluxio

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
)

// getRuntimeInfo gets runtime info
func (e *AlluxioEngine) GetRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}

		e.runtimeInfo, err = base.BuildRuntimeInfo(e.Name, e.Namespace, e.runtimeType, runtime.Spec.TieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
		if err != nil {
			return e.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		e.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)

		if !e.UnitTest {
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
			dataset, err := utils.GetDataset(e.Client, e.Name, e.Namespace)
			if err != nil {
				if utils.IgnoreNotFound(err) == nil {
					e.Log.Info("Dataset is notfound", "name", e.Name, "namespace", e.Namespace)
					return e.runtimeInfo, nil
				}

				e.Log.Info("Failed to get dataset when getruntimeInfo")
				return e.runtimeInfo, err
			}

			e.runtimeInfo.SetupWithDataset(dataset)

			e.Log.Info("Setup with dataset done", "exclusive", e.runtimeInfo.IsExclusive())
		}
	}

	return e.runtimeInfo, nil
}
