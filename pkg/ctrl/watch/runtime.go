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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type runtimeEventHandler struct {
}

func (handler *runtimeEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) bool {
		log.V(1).Info("enter runtimeEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		_, ok := e.Object.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onCreateFunc Skip", "object", e.Object)
			return false
		}

		log.V(1).Info("exit runtimeEventHandler.onCreateFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}

func (handler *runtimeEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		log.V(1).Info("enter runtimeEventHandler.onUpdateFunc", "newObj.name", e.ObjectNew.GetName(), "newObj.namespace", e.ObjectNew.GetNamespace())
		runtimeNew, ok := e.ObjectNew.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		runtimeOld, ok := e.ObjectOld.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if runtimeNew.GetResourceVersion() == runtimeOld.GetResourceVersion() {
			log.V(1).Info("runtime.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("exit runtimeEventHandler.onUpdateFunc", "newObj.name", e.ObjectNew.GetName(), "newObj.namespace", e.ObjectNew.GetNamespace())
		return true
	}
}

func (handler *runtimeEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		log.V(1).Info("enter runtimeEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		_, ok := e.Object.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onDeleteFunc Skip", "object", e.Object)
			return false
		}

		log.V(1).Info("exit runtimeEventHandler.onDeleteFunc", "name", e.Object.GetName(), "namespace", e.Object.GetNamespace())
		return true
	}
}
