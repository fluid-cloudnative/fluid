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
