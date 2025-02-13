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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (j *JuiceFSEngine) HasDeprecatedCommonLabelName() (deprecated bool, err error) {
	// return deprecated.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name

	var (
		fuseName  string = j.getFuseName()
		namespace string = j.namespace
	)

	fuses, err := j.getDaemonset(fuseName, namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			j.Log.Info("Fuses with deprecated label not found")
			deprecated = false
			err = nil
			return
		}
		j.Log.Error(err, "Failed to get fuse", "fuseName", fuseName)
		return deprecated, err
	}

	nodeSelectors := fuses.Spec.Template.Spec.NodeSelector
	j.Log.Info("The current node selectors for worker", "fuseName", fuseName, "nodeSelector", nodeSelectors)

	deprecatedCommonLabelName := utils.GetCommonLabelName(true, j.namespace, j.name, j.runtimeInfo.GetOwnerDatasetUID())
	if _, deprecated = nodeSelectors[deprecatedCommonLabelName]; deprecated {
		//
		j.Log.Info("the deprecated node selector exists", "nodeselector", deprecatedCommonLabelName)
	} else {
		j.Log.Info("The deprecated node selector doesn't exist", "nodeselector", deprecatedCommonLabelName)
	}

	return
}
