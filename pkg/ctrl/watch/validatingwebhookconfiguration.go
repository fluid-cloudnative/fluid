/*
Copyright 2026 The Fluid Authors.

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

// validatingWebhookConfigurationEventHandler enqueues the ValidatingWebhookConfiguration
// named webhookName, so that the webhook reconciler keeps its caBundle patched.
type validatingWebhookConfigurationEventHandler struct{}

func (handler *validatingWebhookConfigurationEventHandler) onCreateFunc(webhookName string) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		validatingWebhookConfiguration, ok := e.Object.(*admissionregistrationv1.ValidatingWebhookConfiguration)
		if !ok {
			log.Info("validatingWebhookConfiguration.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if validatingWebhookConfiguration.GetName() != webhookName {
			log.V(1).Info("validatingWebhookConfiguration.onCreateFunc Skip", "object", e.Object)
			return false
		}

		log.V(1).Info("validatingWebhookConfigurationEventHandler.onCreateFunc", "name", validatingWebhookConfiguration.GetName())
		return true
	}
}

func (handler *validatingWebhookConfigurationEventHandler) onUpdateFunc(webhookName string) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		validatingWebhookConfigurationNew, ok := e.ObjectNew.(*admissionregistrationv1.ValidatingWebhookConfiguration)
		if !ok {
			log.Info("validatingWebhookConfiguration.onUpdateFunc Skip", "object", e.ObjectNew)
			return false
		}

		validatingWebhookConfigurationOld, ok := e.ObjectOld.(*admissionregistrationv1.ValidatingWebhookConfiguration)
		if !ok {
			log.Info("validatingWebhookConfiguration.onUpdateFunc Skip", "object", e.ObjectOld)
			return false
		}

		if validatingWebhookConfigurationOld.GetName() != webhookName || validatingWebhookConfigurationNew.GetName() != webhookName {
			log.V(1).Info("validatingWebhookConfiguration.onUpdateFunc Skip", "object", e.ObjectNew)
			return false
		}

		log.V(1).Info("validatingWebhookConfigurationEventHandler.onUpdateFunc", "name", validatingWebhookConfigurationNew.GetName())
		return true
	}
}

func (handler *validatingWebhookConfigurationEventHandler) onDeleteFunc(webhookName string) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		return false
	}
}
