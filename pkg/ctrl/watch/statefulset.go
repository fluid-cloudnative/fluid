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
