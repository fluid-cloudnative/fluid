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
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestMutatingWebhookConfigurationEventHandler_OnCreateFunc(t *testing.T) {
	var webhookName = "test"
	var fakeWebhookName = "fakeTest"

	// 1. the Object is not mutatingWebhookConfiguration
	createEvent := event.CreateEvent{
		Object: &appsv1.DaemonSet{},
	}
	mutatingWebhookConfigurationEventHandler := &mutatingWebhookConfigurationEventHandler{}
	f := mutatingWebhookConfigurationEventHandler.onCreateFunc(webhookName)
	predicate := f(createEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", createEvent)
	}

	// 2. the Object is mutatingWebhookConfiguration
	createEvent = event.CreateEvent{
		Object: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: webhookName,
			},
		},
	}

	f = mutatingWebhookConfigurationEventHandler.onCreateFunc(webhookName)
	predicate = f(createEvent)

	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", createEvent)
	}

	// 3. the Object is mutatingWebhookConfiguration
	createEvent = event.CreateEvent{
		Object: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: fakeWebhookName,
			},
		},
	}

	f = mutatingWebhookConfigurationEventHandler.onCreateFunc(webhookName)
	predicate = f(createEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", createEvent)
	}

}

func TestMutatingWebhookConfigurationEventHandler_OnUpdateFunc(t *testing.T) {
	var webhookName = "test"
	var fakeWebhookName = "fakeTest"

	// 1. the Object is not mutatingWebhookConfiguration
	updateEvent := event.UpdateEvent{
		ObjectOld: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: webhookName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "old",
				},
			},
		},
		ObjectNew: &appsv1.DaemonSet{},
	}
	mutatingWebhookConfigurationEventHandler := &mutatingWebhookConfigurationEventHandler{}
	f := mutatingWebhookConfigurationEventHandler.onUpdateFunc(webhookName)
	predicate := f(updateEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", updateEvent)
	}

	updateEvent = event.UpdateEvent{
		ObjectOld: &appsv1.DaemonSet{},
		ObjectNew: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: webhookName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "new",
				},
			},
		},
	}
	f = mutatingWebhookConfigurationEventHandler.onUpdateFunc(webhookName)
	predicate = f(updateEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", updateEvent)
	}

	// 2. the Object is mutatingWebhookConfiguration and name is respect
	updateEvent = event.UpdateEvent{
		ObjectOld: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: webhookName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "old",
				},
			},
		},
		ObjectNew: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: webhookName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "new",
				},
			},
		},
	}

	f = mutatingWebhookConfigurationEventHandler.onUpdateFunc(webhookName)
	predicate = f(updateEvent)

	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", updateEvent)
	}

	// 3. the Object is mutatingWebhookConfiguration and name is not respecr
	updateEvent = event.UpdateEvent{
		ObjectOld: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: fakeWebhookName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "old",
				},
			},
		},
		ObjectNew: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: fakeWebhookName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "new",
				},
			},
		},
	}

	f = mutatingWebhookConfigurationEventHandler.onUpdateFunc(webhookName)
	predicate = f(updateEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", updateEvent)
	}

}

func TestMutatingWebhookConfigurationEventHandler_OnDeleteFunc(t *testing.T) {
	var webhookName = "test"

	mutatingWebhookConfigurationEventHandler := &mutatingWebhookConfigurationEventHandler{}
	f := mutatingWebhookConfigurationEventHandler.onDeleteFunc(webhookName)

	deleteEvent := event.DeleteEvent{
		Object: &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: webhookName,
			},
		},
	}

	predicate := f(deleteEvent)

	if predicate {
		t.Errorf("The event %v should not be skip, but not.", deleteEvent)
	}

}
