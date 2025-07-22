/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (e *VineyardEngine) CheckRuntimeReady() (ready bool) {
	//TODO implement me
	return true
}

// getRuntimeInfo gets runtime info
func (e *VineyardEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}

		// Add the symlink method to the vineyard runtime metadata
		if runtime.ObjectMeta.Annotations == nil {
			runtime.ObjectMeta.Annotations = make(map[string]string)
		}
		runtime.ObjectMeta.Annotations["data.fluid.io/metadataList"] = `[{"Labels": {"fluid.io/node-publish-method": "symlink"}, "selector": { "kind": "PersistentVolume"}}]`
		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}
		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, opts...)
		if err != nil {
			return e.runtimeInfo, err
		}
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(e.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := e.getRuntime()
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

			e.runtimeInfo.SetOwnerDatasetUID(firstOwner.UID)
		}
	}

	return e.runtimeInfo, nil
}
