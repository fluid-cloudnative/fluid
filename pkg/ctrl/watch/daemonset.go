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

type daemonsetEventHandler struct {
}

func (handler *daemonsetEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		log.V(1).Info("enter daemonsetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		daemonset, ok := e.Object.(*appsv1.DaemonSet)
		if !ok {
			log.Info("daemonset.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if !isObjectInManaged(daemonset, r) {
			return false
		}

		log.V(1).Info("exit daemonsetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}

func (handler *daemonsetEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		log.V(1).Info("enter daemonsetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		daemonsetNew, ok := e.ObjectNew.(*appsv1.DaemonSet)
		if !ok {
			log.Info("daemonset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if !isObjectInManaged(daemonsetNew, r) {
			return false
		}

		daemonsetOld, ok := e.ObjectOld.(*appsv1.DaemonSet)
		if !ok {
			log.Info("daemonset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if daemonsetNew.GetResourceVersion() == daemonsetOld.GetResourceVersion() {
			log.V(1).Info("daemonset.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("exit daemonsetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		return true
	}
}

func (handler *daemonsetEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		log.V(1).Info("enter daemonsetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		daemonset, ok := e.Object.(*appsv1.DaemonSet)
		if !ok {
			log.Info("daemonset.onDeleteFunc Skip", "object", e.Object)
			return false
		}

		if !isObjectInManaged(daemonset, r) {
			return false
		}

		log.V(1).Info("exit daemonsetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}
