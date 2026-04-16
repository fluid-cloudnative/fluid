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

package juicefs

import (
	"context"
	"fmt"
	"sync"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

// mockEngine is a minimal no-op implementation of base.Engine used in tests only.
type mockEngine struct{}

func (m *mockEngine) ID() string                                             { return "mock" }
func (m *mockEngine) Shutdown() error                                        { return nil }
func (m *mockEngine) Setup(_ cruntime.ReconcileRequestContext) (bool, error) { return true, nil }
func (m *mockEngine) CreateVolume(context.Context) error                     { return nil }
func (m *mockEngine) DeleteVolume(context.Context) error                     { return nil }
func (m *mockEngine) Sync(_ cruntime.ReconcileRequestContext) error          { return nil }
func (m *mockEngine) Validate(_ cruntime.ReconcileRequestContext) error      { return nil }
func (m *mockEngine) Operate(_ cruntime.ReconcileRequestContext, _ *datav1alpha1.OperationStatus, _ dataoperation.OperationInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// newTestJuiceFSReconciler builds a JuiceFSRuntimeReconciler seeded with the
// given scheme and runtime objects.  Pass nil scheme to get a default one.
func newTestJuiceFSReconciler(s *runtime.Scheme, objs ...runtime.Object) *JuiceFSRuntimeReconciler {
	if s == nil {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	}
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	log := ctrl.Log.WithName("juicefs-test")
	recorder := record.NewFakeRecorder(10)
	r := &JuiceFSRuntimeReconciler{
		Scheme:  s,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.RuntimeReconciler = controllers.NewRuntimeReconciler(r, fakeClient, log, recorder)
	return r
}

var _ = Describe("JuiceFSRuntimeReconciler Implement", func() {

	Describe("getRuntime", func() {
		var r *JuiceFSRuntimeReconciler

		BeforeEach(func() {
			testRuntime := &datav1alpha1.JuiceFSRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			r = newTestJuiceFSReconciler(s, testRuntime)
		})

		It("should return the runtime when it exists in the cluster", func() {
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			}
			result, err := r.getRuntime(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("test"))
			Expect(result.Namespace).To(Equal("default"))
		})

		It("should return an error when the runtime does not exist", func() {
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "default"},
			}
			result, err := r.getRuntime(ctx)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("GetOrCreateEngine", func() {
		var r *JuiceFSRuntimeReconciler

		BeforeEach(func() {
			r = newTestJuiceFSReconciler(nil)
		})

		It("should propagate engine creation errors", func() {
			patches := gomonkey.ApplyFunc(ddc.CreateEngine,
				func(_ string, _ cruntime.ReconcileRequestContext) (base.Engine, error) {
					return nil, fmt.Errorf("engine creation failed")
				})
			defer patches.Reset()

			ctx := cruntime.ReconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "fail", Namespace: "default"},
			}
			engine, err := r.GetOrCreateEngine(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("engine creation failed"))
			Expect(engine).To(BeNil())
		})

		It("should create engine on first call and return cached engine on second call", func() {
			mock := &mockEngine{}
			callCount := 0
			patches := gomonkey.ApplyFunc(ddc.CreateEngine,
				func(_ string, _ cruntime.ReconcileRequestContext) (base.Engine, error) {
					callCount++
					return mock, nil
				})
			defer patches.Reset()

			ctx := cruntime.ReconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "cached", Namespace: "default"},
			}

			// First call: engine is created and stored.
			engine1, err := r.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(engine1).To(Equal(base.Engine(mock)))
			Expect(callCount).To(Equal(1))

			// Second call: engine should be retrieved from the cache without re-creation.
			engine2, err := r.GetOrCreateEngine(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(engine2).To(Equal(base.Engine(mock)))
			Expect(callCount).To(Equal(1), "CreateEngine must not be called a second time")
		})
	})

	Describe("RemoveEngine", func() {
		var r *JuiceFSRuntimeReconciler

		BeforeEach(func() {
			r = newTestJuiceFSReconciler(nil)
		})

		It("should remove a cached engine by namespaced name", func() {
			id := ddc.GenerateEngineID(types.NamespacedName{Name: "test", Namespace: "default"})
			r.engines[id] = &mockEngine{}

			ctx := cruntime.ReconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			}
			r.RemoveEngine(ctx)

			_, found := r.engines[id]
			Expect(found).To(BeFalse())
		})

		It("should not panic when removing a non-existent engine", func() {
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "ghost", Namespace: "default"},
			}
			Expect(func() { r.RemoveEngine(ctx) }).NotTo(Panic())
		})
	})

	Describe("Reconcile", func() {
		It("should return no error when the runtime is not found", func() {
			// The fake client has no JuiceFSRuntime objects, so getRuntime will
			// return a NotFound error, which Reconcile should swallow gracefully.
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			r := newTestJuiceFSReconciler(s)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"},
			}
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
})
