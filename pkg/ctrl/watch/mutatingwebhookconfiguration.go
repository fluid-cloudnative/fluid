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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type mutatingWebhookConfigurationEventHandler struct{}

func (handler *mutatingWebhookConfigurationEventHandler) onCreateFunc(webhookName string) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		mutatingWebhookConfiguration, ok := e.Object.(*admissionregistrationv1.MutatingWebhookConfiguration)
		if !ok {
			log.Info("mutatingWebhookConfiguration.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if mutatingWebhookConfiguration.GetName() != webhookName {
			log.V(1).Info("mutatingWebhookConfiguration.onUpdateFunc Skip", "object", e.Object)
			return false
		}

		log.V(1).Info("mutatingWebhookConfigurationEventHandler.onCreateFunc", "name", mutatingWebhookConfiguration.GetName())
		return true
	}
}

func (handler *mutatingWebhookConfigurationEventHandler) onUpdateFunc(webhookName string) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		mutatingWebhookConfigurationNew, ok := e.ObjectNew.(*admissionregistrationv1.MutatingWebhookConfiguration)
		if !ok {
			log.Info("mutatingWebhookConfiguration.onUpdateFunc Skip", "object", e.ObjectNew)
			return false
		}

		mutatingWebhookConfigurationOld, ok := e.ObjectOld.(*admissionregistrationv1.MutatingWebhookConfiguration)
		if !ok {
			log.Info("mutatingWebhookConfiguration.onUpdateFunc Skip", "object", e.ObjectNew)
			return false
		}

		if mutatingWebhookConfigurationOld.GetName() != webhookName || mutatingWebhookConfigurationNew.GetName() != webhookName {
			log.V(1).Info("mutatingWebhookConfiguration.onUpdateFunc Skip", "object", e.ObjectNew)
			return false
		}

		log.V(1).Info("mutatingWebhookConfigurationEventHandler.onUpdateFunc", "name", mutatingWebhookConfigurationNew.GetName())
		return true
	}
}

func (handler *mutatingWebhookConfigurationEventHandler) onDeleteFunc(webhookName string) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		return false
	}
}
