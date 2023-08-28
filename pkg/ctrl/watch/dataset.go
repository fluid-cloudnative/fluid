/*
Copyright 2023 The Fluid Authors.

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

package watch

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type datasetEventHandler struct {
}

func (handler *datasetEventHandler) onUpdateFunc(r Controller, datasetType string) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		log.V(1).Info("enter datasetEventHandler.onUpdateFunc", "newObj.name", e.ObjectNew.GetName(), "newObj.namespace", e.ObjectNew.GetNamespace())
		datasetNew, ok := e.ObjectNew.(*datav1alpha1.Dataset)
		if !ok {
			log.Info("dataset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}
		if len(datasetNew.Status.Runtimes) > 0 && datasetNew.Status.Runtimes[0].Type != datasetType {
			log.V(1).Info("dataset.onUpdateFunc Skip due to runtime type not match", "wantType", datasetType, "datasetType", datasetNew.Status.Runtimes[0].Type)
			return needUpdate
		}

		datasetOld, ok := e.ObjectOld.(*datav1alpha1.Dataset)
		if !ok {
			log.Info("dataset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if datasetNew.GetResourceVersion() == datasetOld.GetResourceVersion() {
			log.V(1).Info("dataset.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("exit datasetEventHandler.onUpdateFunc", "newObj.name", e.ObjectNew.GetName(), "newObj.namespace", e.ObjectNew.GetNamespace())
		return true
	}
}
