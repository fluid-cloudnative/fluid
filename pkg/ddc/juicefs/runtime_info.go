/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/testutil"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// getRuntimeInfo gets runtime info
func (j *JuiceFSEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if j.runtimeInfo == nil {
		runtime, err := j.getRuntime()
		if err != nil {
			return j.runtimeInfo, err
		}
		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}

		j.runtimeInfo, err = base.BuildRuntimeInfo(j.name, j.namespace, j.runtimeType, opts...)
		if err != nil {
			return j.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		j.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)

		j.runtimeInfo.SetFuseName(j.getFuseName())

		if !j.UnitTest {
			// Check if the runtime is using deprecated labels
			isLabelDeprecated, err := j.HasDeprecatedCommonLabelName()
			if err != nil {
				return j.runtimeInfo, err
			}
			j.runtimeInfo.SetDeprecatedNodeLabel(isLabelDeprecated)

			// Check if the runtime is using deprecated naming style for PersistentVolumes
			isPVNameDeprecated, err := volume.HasDeprecatedPersistentVolumeName(j.Client, j.runtimeInfo, j.Log)
			if err != nil {
				return j.runtimeInfo, err
			}
			j.runtimeInfo.SetDeprecatedPVName(isPVNameDeprecated)

			j.Log.Info("Deprecation check finished", "isLabelDeprecated", j.runtimeInfo.IsDeprecatedNodeLabel(), "isPVNameDeprecated", j.runtimeInfo.IsDeprecatedPVName())
		}
	}

	if testutil.IsUnitTest() {
		return j.runtimeInfo, nil
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(j.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := j.getRuntime()
		if err != nil {
			return nil, err
		}

		owners := runtime.GetOwnerReferences()
		if len(owners) > 0 {
			firstOwner := owners[0]
			firstOwnerPath := field.NewPath("metadata").Child("ownerReferences").Index(0)
			if firstOwner.Kind != datav1alpha1.Datasetkind {
				return nil, fmt.Errorf("first owner of the runtime (%s) has invalid Kind \"%s\", expected to be %s ", firstOwnerPath.String(), firstOwner.Kind, datav1alpha1.Datasetkind)
			}

			if firstOwner.Name != runtime.GetName() {
				return nil, fmt.Errorf("first owner of the runtime (%s) has different name with runtime, expected to be same", firstOwnerPath.String())
			}

			j.runtimeInfo.SetOwnerDatasetUID(firstOwner.UID)
		}
	}

	exclusiveModePtr := j.runtimeInfo.IsExclusive()
	if exclusiveModePtr == nil {
		dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			j.runtimeInfo.SetupWithDataset(dataset)
		}
	}

	return j.runtimeInfo, nil
}
