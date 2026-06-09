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

package watch

import (
	"github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type advanceStatefulsetEventHandler struct {
}

func (handler *advanceStatefulsetEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		log.V(1).Info("enter advanceStatefulsetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		advanceStatefulset, ok := e.Object.(*v1alpha1.AdvancedStatefulSet)
		if !ok {
			log.Info("statefulset.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if advanceStatefulset.DeletionTimestamp != nil {
			return false
		}

		if !isObjectInManaged(advanceStatefulset, r) {
			return false
		}

		log.V(1).Info("exit advanceStatefulsetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}

func (handler *advanceStatefulsetEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		log.V(1).Info("enter advanceStatefulsetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		advanceStatefulset, ok := e.ObjectNew.(*v1alpha1.AdvancedStatefulSet)
		if !ok {
			log.Info("statefulset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if !isObjectInManaged(advanceStatefulset, r) {
			return needUpdate
		}

		statefulsetOld, ok := e.ObjectOld.(*v1alpha1.AdvancedStatefulSet)
		if !ok {
			log.Info("statefulset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if advanceStatefulset.GetResourceVersion() == statefulsetOld.GetResourceVersion() {
			log.V(1).Info("statefulset.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("exit advanceStatefulsetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		return true
	}
}

func (handler *advanceStatefulsetEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		log.V(1).Info("enter advanceStatefulsetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		advanceStatefulset, ok := e.Object.(*v1alpha1.AdvancedStatefulSet)
		if !ok {
			log.Info("statefulset.onDeleteFunc Skip", "object", e.Object)
			return false
		}

		if !isObjectInManaged(advanceStatefulset, r) {
			return false
		}

		log.V(1).Info("exit advanceStatefulsetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}
