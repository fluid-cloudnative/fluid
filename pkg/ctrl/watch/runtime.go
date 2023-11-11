/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
