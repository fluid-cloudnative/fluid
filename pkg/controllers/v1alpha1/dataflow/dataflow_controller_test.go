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

package dataflow

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// newTestDataFlowReconciler builds a DataFlowReconciler for unit tests.
func newTestDataFlowReconciler(s *runtime.Scheme, objs ...runtime.Object) *DataFlowReconciler {
	if s == nil {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	}
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	log := logf.Log.WithName("dataflow-test")
	recorder := record.NewFakeRecorder(32)
	return NewDataFlowReconciler(fakeClient, log, recorder, 30*time.Second)
}

var _ = Describe("DataFlowReconciler", func() {

	Describe("ControllerName", func() {
		It("should return the expected controller name", func() {
			r := newTestDataFlowReconciler(nil)
			Expect(r.ControllerName()).To(Equal("DataFlowReconciler"))
		})
	})

	Describe("NewDataFlowReconciler", func() {
		It("should create a non-nil reconciler with correct fields", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			r := newTestDataFlowReconciler(s)
			Expect(r).NotTo(BeNil())
			Expect(r.Client).NotTo(BeNil())
			Expect(r.Recorder).NotTo(BeNil())
			Expect(r.ResyncPeriod).To(Equal(30 * time.Second))
		})
	})

	Describe("Reconcile", func() {
		It("should return no error and no requeue when no operation objects exist for a given name", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)
			r := newTestDataFlowReconciler(s)
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"},
			}
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("should requeue when a DataLoad with RunAfter exists and preceding op is not complete", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)

			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: "default"},
				Status: datav1alpha1.OperationStatus{
					Phase: "Executing",
				},
			}

			waitingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: datav1alpha1.DataLoadSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "preceding",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			r := newTestDataFlowReconciler(s, precedingLoad, waitingLoad)
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			}
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(30 * time.Second))
		})

		It("should not requeue when a DataLoad with RunAfter exists and preceding op is complete", func() {
			s := runtime.NewScheme()
			_ = datav1alpha1.AddToScheme(s)

			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: "default"},
				Status: datav1alpha1.OperationStatus{
					Phase: "Complete",
				},
			}

			waitingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: datav1alpha1.DataLoadSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "preceding",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			r := newTestDataFlowReconciler(s, precedingLoad, waitingLoad)
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
			}
			result, err := r.Reconcile(context.TODO(), req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})
})

// DataFlowEnabled and setupWatches depend on discovery.GetFluidDiscovery() which uses a sync.Once
// singleton that requires a live cluster connection. We test the code paths that avoid touching
// the discovery singleton by temporarily substituting an empty reconcileKinds map so the loop body
// (which calls GetFluidDiscovery) is never entered. This covers the for-loop entry, toSetup
// construction, and return-false / return-bld paths without any cluster dependency.

var _ = Describe("DataFlowEnabled with empty reconcileKinds", func() {
	var saved map[string]client.Object

	BeforeEach(func() {
		saved = reconcileKinds
		reconcileKinds = map[string]client.Object{}
	})

	AfterEach(func() {
		reconcileKinds = saved
	})

	It("should return false when no resource kinds are registered", func() {
		Expect(DataFlowEnabled()).To(BeFalse())
	})
})

var _ = Describe("setupWatches with empty reconcileKinds", func() {
	var saved map[string]client.Object

	BeforeEach(func() {
		saved = reconcileKinds
		reconcileKinds = map[string]client.Object{}
	})

	AfterEach(func() {
		reconcileKinds = saved
	})

	It("should return the builder unchanged when no resource kinds are registered", func() {
		// With an empty reconcileKinds the toSetup slice stays empty, neither bld.For nor
		// bld.Watches is called, and the function returns bld as-is. We pass nil as bld to
		// keep the test self-contained without requiring a real controller-runtime Builder.
		result := setupWatches(nil, nil, builder.Predicates{})
		Expect(result).To(BeNil())
	})
})
