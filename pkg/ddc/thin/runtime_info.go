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

package thin

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (t *ThinEngine) CheckRuntimeReady() (ready bool) {
	//TODO implement me
	return true
}

// getRuntimeInfo gets runtime info
func (t *ThinEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if t.runtimeInfo == nil {
		runtime, err := t.getRuntime()
		if err != nil {
			return t.runtimeInfo, err
		}

		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}

		t.runtimeInfo, err = base.BuildRuntimeInfo(t.name, t.namespace, t.runtimeType, opts...)
		if err != nil {
			return t.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		t.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)

		if !t.UnitTest {
			// Check if the runtime is using deprecated naming style for PersistentVolumes
			isPVNameDeprecated, err := volume.HasDeprecatedPersistentVolumeName(t.Client, t.runtimeInfo, t.Log)
			if err != nil {
				return t.runtimeInfo, err
			}
			t.runtimeInfo.SetDeprecatedPVName(isPVNameDeprecated)

			t.Log.Info("Deprecation check finished", "isLabelDeprecated", t.runtimeInfo.IsDeprecatedNodeLabel(), "isPVNameDeprecated", t.runtimeInfo.IsDeprecatedPVName())

			// Setup with Dataset Info
			dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
			if err != nil {
				if len(runtime.GetOwnerReferences()) > 0 {
					t.runtimeInfo.SetOwnerDatasetUID(runtime.GetOwnerReferences()[0].UID)
				}
				if utils.IgnoreNotFound(err) == nil {
					t.Log.Info("Dataset is notfound", "name", t.name, "namespace", t.namespace)
					return t.runtimeInfo, nil
				}

				t.Log.Info("Failed to get dataset when getruntimeInfo")
				return t.runtimeInfo, err
			}

			t.runtimeInfo.SetupWithDataset(dataset)
			t.runtimeInfo.SetOwnerDatasetUID(dataset.GetUID())

			t.Log.Info("Setup with dataset done", "exclusive", t.runtimeInfo.IsExclusive())
		}
	}

	if t.UnitTest {
		return t.runtimeInfo, nil
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(t.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := t.getRuntime()
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

			t.runtimeInfo.SetOwnerDatasetUID(firstOwner.UID)
		}
	}

	exclusiveModePtr := t.runtimeInfo.IsExclusive()
	if exclusiveModePtr == nil {
		dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			t.runtimeInfo.SetupWithDataset(dataset)
		}
	}

	return t.runtimeInfo, nil
}
