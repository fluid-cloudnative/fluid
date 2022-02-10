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
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/webhook"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type MockReconcile struct{}

func (r *MockReconcile) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

type EventHandlerForMutatingWebhookConfiguration struct {
	certBuilder *webhook.CertificateBuilder
	caCert      []byte
	webhookName string
}

// Create implements EventHandler
func (e *EventHandlerForMutatingWebhookConfiguration) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	fmt.Println("create")
	_ = e.certBuilder.PatchCABundle(e.webhookName, e.caCert)

}

// Update implements EventHandler
func (e *EventHandlerForMutatingWebhookConfiguration) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	fmt.Println("update")
	_ = e.certBuilder.PatchCABundle(e.webhookName, e.caCert)
}

// Delete implements EventHandler
func (e *EventHandlerForMutatingWebhookConfiguration) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	return
}

// Generic implements EventHandler
func (e *EventHandlerForMutatingWebhookConfiguration) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	return
}

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

		if reflect.DeepEqual(mutatingWebhookConfigurationNew.Webhooks, mutatingWebhookConfigurationOld.Webhooks) {
			log.V(1).Info("mutatingWebhookConfiguration.onUpdateFunc Skip due to Webhooks not changed")
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
