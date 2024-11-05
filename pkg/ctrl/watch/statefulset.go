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
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type statefulsetEventHandler struct {
}

func (handler *statefulsetEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		log.V(1).Info("enter statefulsetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		statefulset, ok := e.Object.(*appsv1.StatefulSet)
		if !ok {
			log.Info("statefulset.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if statefulset.DeletionTimestamp != nil {
			return false
		}

		if !isObjectInManaged(statefulset, r) {
			return false
		}

		log.V(1).Info("exit statefulsetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}

func (handler *statefulsetEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		log.V(1).Info("enter statefulsetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		statefulsetNew, ok := e.ObjectNew.(*appsv1.StatefulSet)
		if !ok {
			log.Info("statefulset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if !isObjectInManaged(statefulsetNew, r) {
			return needUpdate
		}

		statefulsetOld, ok := e.ObjectOld.(*appsv1.StatefulSet)
		if !ok {
			log.Info("statefulset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if statefulsetNew.GetResourceVersion() == statefulsetOld.GetResourceVersion() {
			log.V(1).Info("statefulset.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("exit statefulsetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		return true
	}
}

func (handler *statefulsetEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		log.V(1).Info("enter statefulsetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		statefulset, ok := e.Object.(*appsv1.StatefulSet)
		if !ok {
			log.Info("statefulset.onDeleteFunc Skip", "object", e.Object)
			return false
		}

		if !isObjectInManaged(statefulset, r) {
			return false
		}

		log.V(1).Info("exit statefulsetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}
