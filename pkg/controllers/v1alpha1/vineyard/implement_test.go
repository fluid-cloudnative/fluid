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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

var _ = Describe("VineyardRuntime Implement", func() {
	const (
		runtimeName      = "test-vineyard"
		runtimeNamespace = "fluid"
	)

	Describe("getRuntime", func() {
		It("should return the runtime when it exists", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			vineyardRuntime := &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}

			mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
				Client: mockClient,
				Log:    fake.NullLogger(),
			}

			got, err := r.getRuntime(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.Name).To(Equal(runtimeName))
			Expect(got.Namespace).To(Equal(runtimeNamespace))
		})

		It("should return error when runtime does not exist", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			mockClient := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      "nonexistent",
					Namespace: runtimeNamespace,
				},
				Client: mockClient,
				Log:    fake.NullLogger(),
			}

			got, err := r.getRuntime(ctx)
			Expect(err).To(HaveOccurred())
			Expect(got).To(BeNil())
		})
	})

	Describe("GetOrCreateEngine", func() {
		It("should create a new engine when not present in cache", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			vineyardRuntime := &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}

			mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime, dataset)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
				Client:      mockClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.VineyardRuntime,
				EngineImpl:  common.VineyardEngineImpl,
				Runtime:     vineyardRuntime,
			}

			engine, err := r.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
		})

		It("should return cached engine on second call with same context", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			vineyardRuntime := &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}

			mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime, dataset)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
				Client:      mockClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.VineyardRuntime,
				EngineImpl:  common.VineyardEngineImpl,
				Runtime:     vineyardRuntime,
			}

			engine1, err := r.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())

			engine2, err := r.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())

			// Both calls should return the same engine instance (cached)
			Expect(engine1).To(BeIdenticalTo(engine2))
		})

		It("should return error for unknown engine impl", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			vineyardRuntime := &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: runtimeNamespace,
				},
			}

			mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      "test",
					Namespace: runtimeNamespace,
				},
				Client:      mockClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.VineyardRuntime,
				EngineImpl:  "unknown-engine-impl",
				Runtime:     vineyardRuntime,
			}

			_, err := r.GetOrCreateEngine(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RemoveEngine", func() {
		It("should remove an engine from the cache", func() {
			s := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(s)).NotTo(HaveOccurred())

			vineyardRuntime := &datav1alpha1.VineyardRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
			}

			mockClient := fake.NewFakeClientWithScheme(s, vineyardRuntime, dataset)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      runtimeName,
					Namespace: runtimeNamespace,
				},
				Client:      mockClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.VineyardRuntime,
				EngineImpl:  common.VineyardEngineImpl,
				Runtime:     vineyardRuntime,
			}

			// First create an engine
			_, err := r.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())

			// Remove it
			r.RemoveEngine(ctx)

			// engines map should now be empty
			r.mutex.Lock()
			defer r.mutex.Unlock()
			Expect(r.engines).To(BeEmpty())
		})

		It("should be a no-op when engine is not in cache", func() {
			s := runtime.NewScheme()
			mockClient := fake.NewFakeClientWithScheme(s)
			recorder := record.NewFakeRecorder(16)
			r := NewRuntimeReconciler(mockClient, fake.NullLogger(), s, recorder)

			ctx := cruntime.ReconcileRequestContext{
				Context: context.TODO(),
				NamespacedName: types.NamespacedName{
					Name:      "missing",
					Namespace: runtimeNamespace,
				},
				Log: fake.NullLogger(),
			}

			// Should not panic or error
			Expect(func() { r.RemoveEngine(ctx) }).NotTo(Panic())
		})
	})
})
