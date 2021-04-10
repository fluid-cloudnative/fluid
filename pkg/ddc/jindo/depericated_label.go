/*

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

package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/common/deprecated"
)

func (e *JindoEngine) getDeprecatedStorageLabelname(read common.ReadType, storage common.StorageType) string {
	return deprecated.LabelAnnotationStorageCapacityPrefix +
		string(read) +
		e.runtimeType +
		"-" +
		string(storage) +
		e.namespace +
		"-" +
		e.name
}

func (e *JindoEngine) getDeprecatedCommonLabelname() string {
	return deprecated.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name
}

func (e *JindoEngine) getDeprecatedRuntimeLabelname() string {
	return deprecated.LabelAnnotationStorageCapacityPrefix + e.runtimeType + "-" + e.namespace + "-" + e.name
}

func (e *JindoEngine) HasDeprecatedCommonLabelname() (deprecated bool, err error) {

	// return deprecated.LabelAnnotationStorageCapacityPrefix + e.namespace + "-" + e.name

	var (
		workerName string = e.getWorkerDaemonsetName()
		namespace  string = e.namespace
	)

	// runtime, err := e.getRuntime()
	// if err != nil {
	// 	return
	// }

	workers, err := e.getDaemonset(workerName, namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get worker", "workerName", workerName)
		return deprecated, err
	}

	nodeSelectors := workers.Spec.Template.Spec.NodeSelector
	e.Log.Info("The current node selectors for worker", "workerName", workerName, "nodeSelector", nodeSelectors)

	if _, deprecated = nodeSelectors[e.getDeprecatedCommonLabelname()]; deprecated {
		//
		e.Log.Info("the deprecated node selector exists", "nodeselector", e.getDeprecatedCommonLabelname())
	} else {
		e.Log.Info("The deprecated node selector doesn't exist", "nodeselector", e.getDeprecatedCommonLabelname())
	}

	return
}
