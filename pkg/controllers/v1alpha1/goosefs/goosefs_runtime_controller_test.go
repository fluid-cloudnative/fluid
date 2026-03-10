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

package goosefs

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// buildTestReconcileCtx builds a minimal ReconcileRequestContext for unit tests.
func buildTestReconcileCtx(req ctrl.Request) cruntime.ReconcileRequestContext {
	return cruntime.ReconcileRequestContext{
		Context:        context.Background(),
		NamespacedName: req.NamespacedName,
		Log:            zap.New(zap.UseDevMode(true)),
	}
}

const (
	testRuntimeName      = "test-goosefs-runtime"
	testRuntimeNamespace = "default"
)

var _ = Describe("RuntimeReconciler", func() {
	var (
		testScheme *runtime.Scheme
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
	})

	Describe("NewRuntimeReconciler", func() {
		It("should create a RuntimeReconciler with non-nil fields", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme)
			fakeRecorder := record.NewFakeRecorder(10)
			log := zap.New(zap.UseDevMode(true))

			reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)

			Expect(reconciler).NotTo(BeNil())
			Expect(reconciler.Scheme).NotTo(BeNil())
			Expect(reconciler.RuntimeReconciler).NotTo(BeNil())
			Expect(reconciler.engines).NotTo(BeNil())
			Expect(reconciler.mutex).NotTo(BeNil())
		})
	})

	Describe("ControllerName", func() {
		It("should return the expected controller name", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme)
			fakeRecorder := record.NewFakeRecorder(10)
			log := zap.New(zap.UseDevMode(true))

			reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)
			Expect(reconciler.ControllerName()).To(Equal(controllerName))
		})
	})

	Describe("Reconcile", func() {
		Context("when the GooseFSRuntime does not exist", func() {
			It("should return empty result and no error (not-found is ignored)", func() {
				fakeClient := fake.NewFakeClientWithScheme(testScheme)
				fakeRecorder := record.NewFakeRecorder(10)
				log := zap.New(zap.UseDevMode(true))

				reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)

				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      "nonexistent-runtime",
						Namespace: testRuntimeNamespace,
					},
				}
				result, err := reconciler.Reconcile(context.Background(), req)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})

		Context("when a GooseFSRuntime exists", func() {
			It("should attempt to reconcile the runtime without crashing", func() {
				runtime := &datav1alpha1.GooseFSRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      testRuntimeName,
						Namespace: testRuntimeNamespace,
					},
					Spec: datav1alpha1.GooseFSRuntimeSpec{},
				}

				fakeClient := fake.NewFakeClientWithScheme(testScheme, runtime)
				fakeRecorder := record.NewFakeRecorder(10)
				log := zap.New(zap.UseDevMode(true))

				reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)

				req := ctrl.Request{
					NamespacedName: types.NamespacedName{
						Name:      testRuntimeName,
						Namespace: testRuntimeNamespace,
					},
				}
				// Reconcile will call ReconcileInternal which requires a full engine;
				// the result may be an error from engine creation but must not panic.
				_, _ = reconciler.Reconcile(context.Background(), req)
				// Primary assertion: reconciler does not panic and returns
			})
		})
	})

	Describe("GetOrCreateEngine", func() {
		It("should return an error when EngineImpl is unset (no engine created)", func() {
			gooseFSRuntime := &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testRuntimeName,
					Namespace: testRuntimeNamespace,
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, gooseFSRuntime)
			fakeRecorder := record.NewFakeRecorder(10)
			log := zap.New(zap.UseDevMode(true))

			reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)
			Expect(reconciler.engines).To(BeEmpty())

			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      testRuntimeName,
					Namespace: testRuntimeNamespace,
				},
				Client:   fakeClient,
				Log:      log,
				Recorder: fakeRecorder,
				Runtime:  gooseFSRuntime,
				// EngineImpl intentionally left empty — simulates unknown impl
			}

			engine, err := reconciler.GetOrCreateEngine(ctx)
			// An empty EngineImpl causes CreateEngine to return an error; no engine stored.
			Expect(err).To(HaveOccurred())
			Expect(engine).To(BeNil())
			Expect(reconciler.engines).To(BeEmpty())
		})
	})

	Describe("RemoveEngine", func() {
		It("should not panic when removing a non-existent engine", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme)
			fakeRecorder := record.NewFakeRecorder(10)
			log := zap.New(zap.UseDevMode(true))

			reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)

			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      testRuntimeName,
					Namespace: testRuntimeNamespace,
				},
				Client:   fakeClient,
				Log:      log,
				Recorder: fakeRecorder,
			}

			// RemoveEngine on an empty map must not panic.
			Expect(func() {
				reconciler.RemoveEngine(ctx)
			}).NotTo(Panic())
			Expect(reconciler.engines).To(BeEmpty())
		})
	})

	Describe("getRuntime", func() {
		It("should return error when runtime does not exist", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme)
			fakeRecorder := record.NewFakeRecorder(10)
			log := zap.New(zap.UseDevMode(true))

			reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "missing",
					Namespace: testRuntimeNamespace,
				},
			}
			ctx := buildTestReconcileCtx(req)
			rt, err := reconciler.getRuntime(ctx)
			Expect(err).To(HaveOccurred())
			Expect(rt).To(BeNil())
		})

		It("should return the runtime when it exists", func() {
			gooseFSRuntime := &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testRuntimeName,
					Namespace: testRuntimeNamespace,
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, gooseFSRuntime)
			fakeRecorder := record.NewFakeRecorder(10)
			log := zap.New(zap.UseDevMode(true))

			reconciler := NewRuntimeReconciler(fakeClient, log, testScheme, fakeRecorder)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      testRuntimeName,
					Namespace: testRuntimeNamespace,
				},
			}
			ctx := buildTestReconcileCtx(req)
			rt, err := reconciler.getRuntime(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).NotTo(BeNil())
			Expect(rt.Name).To(Equal(testRuntimeName))
			Expect(rt.Namespace).To(Equal(testRuntimeNamespace))
		})
	})

	Describe("GooseFSRuntime CRD via envtest", func() {
		It("should create and retrieve a GooseFSRuntime via the k8sClient", func() {
			gooseFSRuntime := &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "envtest-runtime",
					Namespace: "default",
				},
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}

			By("creating the GooseFSRuntime")
			Expect(k8sClient.Create(context.Background(), gooseFSRuntime)).To(Succeed())

			By("retrieving the GooseFSRuntime")
			var fetched datav1alpha1.GooseFSRuntime
			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      "envtest-runtime",
				Namespace: "default",
			}, &fetched)).To(Succeed())
			Expect(fetched.Name).To(Equal("envtest-runtime"))

			By("deleting the GooseFSRuntime")
			Expect(k8sClient.Delete(context.Background(), gooseFSRuntime)).To(Succeed())
		})
	})
})
