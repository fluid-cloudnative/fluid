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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mutatingWebhookConfigurationEventHandler", func() {
	var (
		webhookName     = "test"
		fakeWebhookName = "fakeTest"
		handler         *mutatingWebhookConfigurationEventHandler
	)

	BeforeEach(func() {
		handler = &mutatingWebhookConfigurationEventHandler{}
	})

	Describe("onCreateFunc", func() {
		It("should not reconcile if the object is not a MutatingWebhookConfiguration", func() {
			createEvent := event.CreateEvent{
				Object: &appsv1.DaemonSet{},
			}
			f := handler.onCreateFunc(webhookName)
			Expect(f(createEvent)).To(BeFalse())
		})

		It("should reconcile if the object is a MutatingWebhookConfiguration with the correct name", func() {
			createEvent := event.CreateEvent{
				Object: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: webhookName,
					},
				},
			}
			f := handler.onCreateFunc(webhookName)
			Expect(f(createEvent)).To(BeTrue())
		})

		It("should not reconcile if the object is a MutatingWebhookConfiguration with a different name", func() {
			createEvent := event.CreateEvent{
				Object: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: fakeWebhookName,
					},
				},
			}
			f := handler.onCreateFunc(webhookName)
			Expect(f(createEvent)).To(BeFalse())
		})
	})

	Describe("onUpdateFunc", func() {
		It("should not reconcile if the new object is not a MutatingWebhookConfiguration", func() {
			updateEvent := event.UpdateEvent{
				ObjectOld: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: webhookName,
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{Name: "old"},
					},
				},
				ObjectNew: &appsv1.DaemonSet{},
			}
			f := handler.onUpdateFunc(webhookName)
			Expect(f(updateEvent)).To(BeFalse())
		})

		It("should not reconcile if the old object is not a MutatingWebhookConfiguration", func() {
			updateEvent := event.UpdateEvent{
				ObjectOld: &appsv1.DaemonSet{},
				ObjectNew: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: webhookName,
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{Name: "new"},
					},
				},
			}
			f := handler.onUpdateFunc(webhookName)
			Expect(f(updateEvent)).To(BeFalse())
		})

		It("should reconcile if both objects are MutatingWebhookConfiguration with the correct name", func() {
			updateEvent := event.UpdateEvent{
				ObjectOld: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: webhookName,
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{Name: "old"},
					},
				},
				ObjectNew: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: webhookName,
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{Name: "new"},
					},
				},
			}
			f := handler.onUpdateFunc(webhookName)
			Expect(f(updateEvent)).To(BeTrue())
		})

		It("should not reconcile if both objects are MutatingWebhookConfiguration with a different name", func() {
			updateEvent := event.UpdateEvent{
				ObjectOld: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: fakeWebhookName,
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{Name: "old"},
					},
				},
				ObjectNew: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: fakeWebhookName,
					},
					Webhooks: []admissionregistrationv1.MutatingWebhook{
						{Name: "new"},
					},
				},
			}
			f := handler.onUpdateFunc(webhookName)
			Expect(f(updateEvent)).To(BeFalse())
		})
	})

	Describe("onDeleteFunc", func() {
		It("should not reconcile on delete", func() {
			deleteEvent := event.DeleteEvent{
				Object: &admissionregistrationv1.MutatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name: webhookName,
					},
				},
			}
			f := handler.onDeleteFunc(webhookName)
			Expect(f(deleteEvent)).To(BeFalse())
		})
	})
})
