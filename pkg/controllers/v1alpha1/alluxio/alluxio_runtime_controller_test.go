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

package alluxio

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

var _ = Describe("AlluxioRuntimeController", func() {
	const (
		testName      = "alluxio-test"
		testNamespace = "default"
	)

	var (
		s *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
	})

	Describe("NewRuntimeReconciler", func() {
		It("creates a RuntimeReconciler with expected fields", func() {
			c := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(10)
			r := NewRuntimeReconciler(c, fake.NullLogger(), s, recorder)

			Expect(r).ToNot(BeNil())
			Expect(r.Scheme).To(Equal(s))
			Expect(r.engines).ToNot(BeNil())
			Expect(r.mutex).ToNot(BeNil())
			Expect(r.RuntimeReconciler).ToNot(BeNil())
		})
	})

	Describe("ControllerName", func() {
		It("returns the expected controller name", func() {
			c := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(10)
			r := NewRuntimeReconciler(c, fake.NullLogger(), s, recorder)

			Expect(r.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("Reconcile", func() {
		It("returns empty result when the AlluxioRuntime is not found", func() {
			c := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(10)
			r := NewRuntimeReconciler(c, fake.NullLogger(), s, recorder)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "nonexistent",
					Namespace: testNamespace,
				},
			}
			result, err := r.Reconcile(context.Background(), req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("proceeds past getRuntime when the AlluxioRuntime exists", func() {
			rt := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testName,
					Namespace: testNamespace,
				},
			}
			c := fake.NewFakeClientWithScheme(s, rt)
			recorder := record.NewFakeRecorder(10)
			r := NewRuntimeReconciler(c, fake.NullLogger(), s, recorder)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      testName,
					Namespace: testNamespace,
				},
			}
			// Reconcile proceeds past getRuntime, builds the engine, then hits the
			// "no dataset bound" branch which returns RequeueAfter(5s) with no error.
			result, err := r.Reconcile(context.Background(), req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.RequeueAfter).ToNot(BeZero())
		})
	})
})
