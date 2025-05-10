/*
Copyright 2024 The Fluid Author.

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

package validation

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func IsValidDataset(dataset v1alpha1.Dataset, enableMountValidation bool) error {
	if errs := validation.IsDNS1035Label(dataset.ObjectMeta.Name); len(dataset.ObjectMeta.Name) > 0 && len(errs) > 0 {
		return field.Invalid(field.NewPath("metadata").Child("name"), dataset.ObjectMeta.Name, strings.Join(errs, ","))
	}

	// 0.1 Validate the mount name and mount path
	// Users can set the environment variable to 'false' to disable this validation
	// Default is true
	if !enableMountValidation {
		return nil
	}
	for _, mount := range dataset.Spec.Mounts {
		// The field mount.Name and mount.Path is optional
		// Empty name or path is allowed
		if len(mount.Name) != 0 {
			// If users set the mount.Name, it should comply with the DNS1035 rule.
			if errs := validation.IsDNS1035Label(mount.Name); len(errs) > 0 {
				return field.Invalid(field.NewPath("spec").Child("mounts").Child("name"), mount.Name, strings.Join(errs, ","))
			}
		}
		if len(mount.Path) != 0 {
			// If users set the mount.Path, check it.
			if err := IsValidMountPath(mount.Path); err != nil {
				return field.Invalid(field.NewPath("spec").Child("mounts").Child("path"), mount.Path, err.Error())
			}
		}
	}
	return nil
}
