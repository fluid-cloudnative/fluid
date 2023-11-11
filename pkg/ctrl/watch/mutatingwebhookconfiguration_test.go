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
