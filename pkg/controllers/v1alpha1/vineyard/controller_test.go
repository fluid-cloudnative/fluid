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

package vineyard

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("VineyardRuntime Controller", func() {
	Describe("NewRuntimeReconciler", func() {
		It("should create a RuntimeReconciler with non-nil fields", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())
			mockClient := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(16)

			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			Expect(r).NotTo(BeNil())
			Expect(r.Scheme).NotTo(BeNil())
			Expect(r.engines).NotTo(BeNil())
			Expect(r.mutex).NotTo(BeNil())
			Expect(r.RuntimeReconciler).NotTo(BeNil())
		})
	})

	Describe("ControllerName", func() {
		It("should return the correct controller name", func() {
			s := runtime.NewScheme()
			mockClient := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(16)

			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			Expect(r.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("Reconcile", func() {
		It("should return empty result when runtime is not found", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())
			mockClient := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(16)

			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "nonexistent-runtime",
					Namespace: "default",
				},
			}

			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("should requeue with error when runtime name violates the controller DNS-1035 validation rule", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			// Runtime name starting with a digit fails the controller's DNS-1035 validation.
			invalidName := "20-vineyard"
			vineyardRuntime := &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      invalidName,
					Namespace: "default",
				},
			}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      invalidName,
					Namespace: "default",
				},
			}

			mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime, dataset)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      invalidName,
					Namespace: "default",
				},
			}

			_, err := r.Reconcile(context.TODO(), req)
			// Invalid runtime name causes an error path
			Expect(err).To(HaveOccurred())
		})
	})
})
