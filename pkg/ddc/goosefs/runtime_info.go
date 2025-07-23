/*
Copyright 2022 The Fluid Authors.

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

package goosefs

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
func (e *GooseFSEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}
		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}

		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, opts...)
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

		}
	}

	if testutil.IsUnitTest() {
		return e.runtimeInfo, nil
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(e.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
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

			e.runtimeInfo.SetOwnerDatasetUID(firstOwner.UID)
		}
	}

	exclusiveModePtr := e.runtimeInfo.IsExclusive()
	if exclusiveModePtr == nil {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			e.runtimeInfo.SetupWithDataset(dataset)
		}
	}

	return e.runtimeInfo, nil
}
