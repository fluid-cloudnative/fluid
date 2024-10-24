/*
Copyright 2022 The Fluid Authors.

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
