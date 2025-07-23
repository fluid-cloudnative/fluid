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

package thin

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type validateFn func(ctx cruntime.ReconcileRequestContext) error

var checks []validateFn = []validateFn{
	validateDuplicateDatasetMounts,
}

func validateDuplicateDatasetMounts(ctx cruntime.ReconcileRequestContext) error {
	if ctx.Dataset == nil {
		return nil
	}

	if len(ctx.Dataset.Spec.Mounts) == 0 {
		return nil
	}

	fieldErrorList := field.ErrorList{}

	existedMountNames := map[string]int{}
	existedMountPath := map[string]int{}
	for idx, mount := range ctx.Dataset.Spec.Mounts {
		path := mount.Path
		if len(path) == 0 {
			path = fmt.Sprintf("/%s", mount.Name)
		}

		if _, exists := existedMountNames[mount.Name]; exists {
			fieldErrorList = append(fieldErrorList, field.Duplicate(field.NewPath("Dataset").Child("spec", "mounts").Index(idx).Child("name"), mount.Name))
			continue // Skip to next iteration because collided mount names imply collided mount paths.
		}

		if _, exists := existedMountPath[path]; exists {
			fieldErrorList = append(fieldErrorList, field.Duplicate(field.NewPath("Dataset").Child("spec", "mounts").Index(idx).Child("path"), path))
			continue
		}

		existedMountNames[mount.Name] = idx
		existedMountPath[path] = idx
	}

	return fieldErrorList.ToAggregate()
}

func (t *ThinEngine) Validate(ctx cruntime.ReconcileRequestContext) (err error) {
	// XXXEngine.runtimeInfo must have full information about the bound dataset for further reconcilation.
	// getRuntimeInfo() here is a refresh to make sure the information is correctly set
	runtimeInfo, err := t.getRuntimeInfo()
	if err != nil {
		return err
	}

	err = base.ValidateRuntimeInfo(runtimeInfo)
	if err != nil {
		return err
	}

	for _, checkFn := range checks {
		if err := checkFn(ctx); err != nil {
			return err
		}
	}
	return nil
}
