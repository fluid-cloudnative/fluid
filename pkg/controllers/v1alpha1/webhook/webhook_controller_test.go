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

package webhook

import (
	"context"
	"errors"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluid-cloudnative/fluid/pkg/webhook"
)

var _ = Describe("WebhookReconciler", func() {
	const (
		webhookName           = "fluid-webhook"
		validatingWebhookName = "fluid-validating-webhook"
	)

	var (
		caCert = []byte("test-ca-cert")
	)

	Describe("Reconcile", func() {
		var (
			reconciler  *WebhookReconciler
			certBuilder *webhook.CertificateBuilder
		)

		BeforeEach(func() {
			certBuilder = &webhook.CertificateBuilder{}
			reconciler = &WebhookReconciler{
				CertBuilder:           certBuilder,
				WebhookName:           webhookName,
				ValidatingWebhookName: validatingWebhookName,
				CaCert:                caCert,
			}
		})

		Context("when the request names the ValidatingWebhookConfiguration", func() {
			It("should patch the validating caBundle only", func() {
				validatingPatched := false
				patches := gomonkey.ApplyMethod(
					certBuilder,
					"PatchValidatingCABundle",
					func(_ *webhook.CertificateBuilder, name string, ca []byte) error {
						Expect(name).To(Equal(validatingWebhookName))
						Expect(ca).To(Equal(caCert))
						validatingPatched = true
						return nil
					},
				).ApplyMethod(
					certBuilder,
					"PatchCABundle",
					func(_ *webhook.CertificateBuilder, _ string, _ []byte) error {
						Fail("PatchCABundle must not be called for the validating configuration")
						return nil
					},
				)
				defer patches.Reset()

				result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: validatingWebhookName}})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(BeZero())
				Expect(validatingPatched).To(BeTrue())
			})

			It("should requeue after 10 seconds when the validating patch fails", func() {
				patch := gomonkey.ApplyMethod(
					certBuilder,
					"PatchValidatingCABundle",
					func(_ *webhook.CertificateBuilder, _ string, _ []byte) error {
						return errors.New("patch failed")
					},
				)
				defer patch.Reset()

				result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: validatingWebhookName}})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(10 * time.Second))
			})
		})

		Context("when the request names the MutatingWebhookConfiguration", func() {
			It("should patch the mutating caBundle only", func() {
				patches := gomonkey.ApplyMethod(
					certBuilder,
					"PatchCABundle",
					func(_ *webhook.CertificateBuilder, name string, _ []byte) error {
						Expect(name).To(Equal(webhookName))
						return nil
					},
				).ApplyMethod(
					certBuilder,
					"PatchValidatingCABundle",
					func(_ *webhook.CertificateBuilder, _ string, _ []byte) error {
						Fail("PatchValidatingCABundle must not be called for the mutating configuration")
						return nil
					},
				)
				defer patches.Reset()

				result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: types.NamespacedName{Name: webhookName}})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(BeZero())
			})
		})

		Context("when PatchCABundle succeeds", func() {
			It("should return no requeue and no error", func() {
				patch := gomonkey.ApplyMethod(
					certBuilder,
					"PatchCABundle",
					func(_ *webhook.CertificateBuilder, name string, ca []byte) error {
						Expect(name).To(Equal(webhookName))
						Expect(ca).To(Equal(caCert))
						return nil
					},
				)
				defer patch.Reset()

				result, err := reconciler.Reconcile(context.Background(), ctrl.Request{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(BeZero())
			})
		})

		Context("when PatchCABundle fails", func() {
			It("should requeue after 10 seconds and return no error", func() {
				patch := gomonkey.ApplyMethod(
					certBuilder,
					"PatchCABundle",
					func(_ *webhook.CertificateBuilder, _ string, _ []byte) error {
						return errors.New("patch failed")
					},
				)
				defer patch.Reset()

				result, err := reconciler.Reconcile(context.Background(), ctrl.Request{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(10 * time.Second))
			})
		})
	})
})
