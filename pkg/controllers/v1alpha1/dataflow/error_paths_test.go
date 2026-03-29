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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const serverUnavailable = "server unavailable"

// newGetErrorClient returns a client that fails every Get with a generic error.
func newGetErrorClient(s *runtime.Scheme) client.Client {
	injectErr := fmt.Errorf(serverUnavailable)
	fakeBase := fakeclient.NewClientBuilder().WithScheme(s).Build()
	return interceptor.NewClient(fakeBase, interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			return injectErr
		},
	})
}

var _ = Describe("Reconcile error path: Get failure propagates to error return", func() {

	var s *runtime.Scheme

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should return an error when the underlying client Get fails", func() {
		log := logf.Log.WithName("dataflow-error-test")
		recorder := record.NewFakeRecorder(32)
		r := NewDataFlowReconciler(newGetErrorClient(s), log, recorder, 30*time.Second)
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"},
		}
		result, err := r.Reconcile(context.TODO(), req)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(serverUnavailable))
		// RequeueIfError returns ctrl.Result{} + the error
		Expect(result).To(Equal(ctrl.Result{}))
	})
})

var _ = Describe("reconcileDataLoad: outer Get failure", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should return wrapped error when client Get fails for DataLoad", func() {
		ctx := reconcileRequestContext{
			Context:        context.TODO(),
			NamespacedName: types.NamespacedName{Name: "test", Namespace: namespace},
			Client:         newGetErrorClient(s),
			Log:            logf.Log.WithName("test"),
			Recorder:       record.NewFakeRecorder(32),
		}

		needRequeue, err := reconcileDataLoad(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get dataload"))
		Expect(needRequeue).To(BeTrue())
	})
})

var _ = Describe("reconcileDataBackup: outer Get failure", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should return wrapped error when client Get fails for DataBackup", func() {
		ctx := reconcileRequestContext{
			Context:        context.TODO(),
			NamespacedName: types.NamespacedName{Name: "test", Namespace: namespace},
			Client:         newGetErrorClient(s),
			Log:            logf.Log.WithName("test"),
			Recorder:       record.NewFakeRecorder(32),
		}

		needRequeue, err := reconcileDataBackup(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get databackup"))
		Expect(needRequeue).To(BeTrue())
	})
})

var _ = Describe("reconcileDataMigrate: outer Get failure", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should return wrapped error when client Get fails for DataMigrate", func() {
		ctx := reconcileRequestContext{
			Context:        context.TODO(),
			NamespacedName: types.NamespacedName{Name: "test", Namespace: namespace},
			Client:         newGetErrorClient(s),
			Log:            logf.Log.WithName("test"),
			Recorder:       record.NewFakeRecorder(32),
		}

		needRequeue, err := reconcileDataMigrate(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get datamigrate"))
		Expect(needRequeue).To(BeTrue())
	})
})

var _ = Describe("reconcileDataProcess: outer Get failure", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should return wrapped error when client Get fails for DataProcess", func() {
		ctx := reconcileRequestContext{
			Context:        context.TODO(),
			NamespacedName: types.NamespacedName{Name: "test", Namespace: namespace},
			Client:         newGetErrorClient(s),
			Log:            logf.Log.WithName("test"),
			Recorder:       record.NewFakeRecorder(32),
		}

		needRequeue, err := reconcileDataProcess(ctx)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to get dataprocess"))
		Expect(err.Error()).To(ContainSubstring(serverUnavailable))
		Expect(needRequeue).To(BeTrue())
	})
})

var _ = Describe("reconcileOperationDataFlow: updateStatusFn inner Get NotFound path", func() {
	// Verify that when the inner updateStatusFn re-fetches the object and gets NotFound,
	// reconcileDataLoad returns (false, nil) — i.e., the not-found is treated as a no-op.

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should succeed (no error, no requeue) when updateStatusFn inner Get returns NotFound", func() {
		precedingLoad := &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
			Status: datav1alpha1.OperationStatus{
				Phase: "Complete",
			},
		}
		waitingLoad := &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{Name: "waiting", Namespace: namespace},
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

		// Use a call-counting interceptor:
		// calls 1–2: outer reconcileDataLoad fetches "waiting" and "preceding" → pass through
		// call 3+: inside updateStatusFn fetch → return a custom not-found error
		callCount := 0
		fakeBase := fakeclient.NewClientBuilder().WithScheme(s).
			WithRuntimeObjects(precedingLoad, waitingLoad).
			WithStatusSubresource(waitingLoad).
			Build()

		countClient := interceptor.NewClient(fakeBase, interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				callCount++
				if callCount >= 3 && key.Name == "waiting" {
					// Wrap a notFound to exercise the IgnoreNotFound branch in updateStatusFn
					return &statusNotFoundError{name: key.Name}
				}
				return c.Get(ctx, key, obj, opts...)
			},
		})

		rCtx := reconcileRequestContext{
			Context:        context.TODO(),
			NamespacedName: types.NamespacedName{Name: "waiting", Namespace: namespace},
			Client:         countClient,
			Log:            logf.Log.WithName("test"),
			Recorder:       record.NewFakeRecorder(32),
		}

		needRequeue, err := reconcileDataLoad(rCtx)
		Expect(err).NotTo(HaveOccurred())
		Expect(needRequeue).To(BeFalse())
	})
})

// statusNotFoundError satisfies the k8s.io/apimachinery/pkg/api/errors IsNotFound check.
// utils.IgnoreNotFound uses apierrors.IsNotFound which checks for Reason == StatusReasonNotFound.
type statusNotFoundError struct{ name string }

func (e *statusNotFoundError) Error() string { return fmt.Sprintf("%q not found", e.name) }
func (e *statusNotFoundError) Status() metav1.Status {
	return metav1.Status{Reason: metav1.StatusReasonNotFound}
}

var _ = Describe("reconcileDataBackup: updateStatusFn inner Get NotFound path", func() {
	// Verify that when updateStatusFn re-fetches the DataBackup and gets NotFound,
	// reconcileDataBackup returns (false, nil) — the not-found is treated as a no-op.

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	It("should succeed when updateStatusFn inner Get returns NotFound", func() {
		precedingLoad := &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{Name: "load-1", Namespace: namespace},
			Status: datav1alpha1.OperationStatus{
				Phase: "Complete",
			},
		}
		waitingBackup := &datav1alpha1.DataBackup{
			ObjectMeta: metav1.ObjectMeta{Name: backupName, Namespace: namespace},
			Spec: datav1alpha1.DataBackupSpec{
				RunAfter: &datav1alpha1.OperationRef{
					ObjectRef: datav1alpha1.ObjectRef{
						Kind: "DataLoad",
						Name: "load-1",
					},
				},
			},
			Status: datav1alpha1.OperationStatus{
				WaitingFor: datav1alpha1.WaitingStatus{
					OperationComplete: ptr.To(true),
				},
			},
		}

		callCount := 0
		fakeBase := fakeclient.NewClientBuilder().WithScheme(s).
			WithRuntimeObjects(precedingLoad, waitingBackup).
			WithStatusSubresource(waitingBackup).
			Build()

		countClient := interceptor.NewClient(fakeBase, interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				callCount++
				if callCount >= 3 && key.Name == backupName {
					return &statusNotFoundError{name: key.Name}
				}
				return c.Get(ctx, key, obj, opts...)
			},
		})

		rCtx := reconcileRequestContext{
			Context:        context.TODO(),
			NamespacedName: types.NamespacedName{Name: backupName, Namespace: namespace},
			Client:         countClient,
			Log:            logf.Log.WithName("test"),
			Recorder:       record.NewFakeRecorder(32),
		}

		needRequeue, err := reconcileDataBackup(rCtx)
		Expect(err).NotTo(HaveOccurred())
		Expect(needRequeue).To(BeFalse())
	})
})
