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
	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type cacheworkersetEventHandler struct {
}

func (handler *cacheworkersetEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		log.V(1).Info("enter cacheworkersetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		cacheworkerset, ok := e.Object.(*cacheworkerset.CacheWorkerSet)
		if !ok {
			log.Info("cacheworkerset.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if cacheworkerset.DeletionTimestamp != nil {
			return false
		}

		if !isObjectInManaged(cacheworkerset, r) {
			return false
		}

		log.V(1).Info("exit cacheworkersetEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}

func (handler *cacheworkersetEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		log.V(1).Info("enter cacheworkersetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		cacheworkersetNew, ok := e.ObjectNew.(*cacheworkerset.CacheWorkerSet)
		if !ok {
			log.Info("cacheworkerset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if !isObjectInManaged(cacheworkersetNew, r) {
			return needUpdate
		}

		cacheworkersetOld, ok := e.ObjectOld.(*cacheworkerset.CacheWorkerSet)
		if !ok {
			log.Info("cacheworkerset.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if cacheworkersetNew.GetResourceVersion() == cacheworkersetOld.GetResourceVersion() {
			log.V(1).Info("cacheworkerset.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("exit cacheworkersetEventHandler.onUpdateFunc", "name", e.ObjectNew.GetName(), "namespace", e.ObjectNew.GetNamespace())
		return true
	}
}

func (handler *cacheworkersetEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		log.V(1).Info("enter cacheworkersetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		cacheworkerset, ok := e.Object.(*cacheworkerset.CacheWorkerSet)
		if !ok {
			log.Info("cacheworkerset.onDeleteFunc Skip", "object", e.Object)
			return false
		}

		if !isObjectInManaged(cacheworkerset, r) {
			return false
		}

		log.V(1).Info("exit cacheworkersetEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}
