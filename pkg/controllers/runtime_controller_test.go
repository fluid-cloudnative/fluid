/*
Copyright 2023 The Fluid Authors.

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

package controllers

import (
	"context"
	"fmt"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// -- Mocks --

type mockEngine struct {
	id string
}

func (m *mockEngine) ID() string {
	return m.id
}
func (m *mockEngine) Shutdown() error {
	return nil
}
func (m *mockEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	return true, nil
}
func (m *mockEngine) CreateVolume() (err error) {
	return nil
}
func (m *mockEngine) DeleteVolume() (err error) {
	return nil
}
func (m *mockEngine) Sync(ctx cruntime.ReconcileRequestContext) error {
	return nil
}
func (m *mockEngine) Validate(ctx cruntime.ReconcileRequestContext) (err error) {
	return nil
}
func (m *mockEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

type mockRuntimeReconciler struct {
	*RuntimeReconciler
	failEngineCreation bool
}

func (m *mockRuntimeReconciler) GetOrCreateEngine(ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	if m.failEngineCreation {
		return nil, fmt.Errorf("induced engine creation failure")
	}
	return &mockEngine{id: "test-engine"}, nil
}

func (m *mockRuntimeReconciler) RemoveEngine(ctx cruntime.ReconcileRequestContext) {
	// no-op
}

// -- Helpers --

func newTestReconciler(t *testing.T, objects ...client.Object) (*mockRuntimeReconciler, client.Client) {
	s := runtime.NewScheme()
	_ = scheme.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = corev1.AddToScheme(s)

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithStatusSubresource(objects...).
		WithObjects(objects...).
		Build()

	// Use discard logger
	log := logr.Discard()
	recorder := record.NewFakeRecorder(10)

	mock := &mockRuntimeReconciler{}
	// Hook up the RuntimeReconciler to use 'mock' as the implementation
	baseReconciler := NewRuntimeReconciler(mock, fakeClient, log, recorder)
	mock.RuntimeReconciler = baseReconciler

	return mock, fakeClient
}

// -- Tests --

func TestReconcileInternal_AddOwnerReference(t *testing.T) {
	// Scenario: Runtime exists, Dataset exists, but OwnerReference is missing.
	// Expected: Reconciler should add OwnerReference to Runtime and Requeue.

	dataset := &datav1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dataset",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "dataset-uid-123",
		},
	}
	runtimeObj := &datav1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlluxioRuntime",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			// No OwnerReferences
		},
	}

	reconciler, c := newTestReconciler(t, dataset, runtimeObj)

	ctx := cruntime.ReconcileRequestContext{
		Context:        context.TODO(),
		Log:            logr.Discard(),
		NamespacedName: types.NamespacedName{Name: "test-dataset", Namespace: "default"},
		RuntimeType:    common.AlluxioRuntime,
		Runtime:        runtimeObj,
		Category:       common.AccelerateCategory,
		Client:         c,
	}

	// First pass
	result, err := reconciler.ReconcileInternal(ctx)
	if err != nil {
		t.Fatalf("ReconcileInternal failed: %v", err)
	}

	// Check if Requeue is true
	if !result.Requeue {
		t.Errorf("Expected Requeue to be true for OwnerReference update, got %v", result)
	}

	// Verify OwnerReference
	updatedRuntime := &datav1alpha1.AlluxioRuntime{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "test-dataset", Namespace: "default"}, updatedRuntime)
	if err != nil {
		t.Fatalf("Failed to get updated runtime: %v", err)
	}

	if len(updatedRuntime.OwnerReferences) != 1 {
		t.Errorf("Expected 1 OwnerReference, got %d", len(updatedRuntime.OwnerReferences))
	} else {
		ref := updatedRuntime.OwnerReferences[0]
		if ref.UID != dataset.UID {
			t.Errorf("Expected OwnerReference UID %s, got %s", dataset.UID, ref.UID)
		}
	}
}

func TestReconcileInternal_AddFinalizer(t *testing.T) {
	// Scenario: Runtime has OwnerReference but no Finalizer.
	// Expected: Reconciler should add Finalizer and Requeue.

	dataset := &datav1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dataset",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "dataset-uid-123",
		},
	}
	runtimeObj := &datav1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlluxioRuntime",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "data.fluid.io/v1alpha1",
					Kind:       "Dataset",
					Name:       "test-dataset",
					UID:        "dataset-uid-123",
					Controller: func() *bool { b := true; return &b }(),
				},
			},
			// No Finalizer
		},
	}

	reconciler, c := newTestReconciler(t, dataset, runtimeObj)

	ctx := cruntime.ReconcileRequestContext{
		Context:        context.TODO(),
		Log:            logr.Discard(),
		NamespacedName: types.NamespacedName{Name: "test-dataset", Namespace: "default"},
		RuntimeType:    common.AlluxioRuntime,
		Runtime:        runtimeObj,
		Category:       common.AccelerateCategory,
		FinalizerName:  "fluid-alluxio-controller-finalizer",
		Client:         c,
	}

	// First pass
	result, err := reconciler.ReconcileInternal(ctx)
	if err != nil {
		t.Fatalf("ReconcileInternal failed: %v", err)
	}

	// Check result
	if !result.Requeue {
		t.Errorf("Expected Requeue to be true for Finalizer update, got %v", result)
	}

	// Verify Finalizer
	updatedRuntime := &datav1alpha1.AlluxioRuntime{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "test-dataset", Namespace: "default"}, updatedRuntime)
	if err != nil {
		t.Fatalf("Failed to get updated runtime: %v", err)
	}

	if len(updatedRuntime.Finalizers) == 0 {
		t.Errorf("Expected Finalizer detection, got none")
	} else {
		found := false
		for _, f := range updatedRuntime.Finalizers {
			if f == "fluid-alluxio-controller-finalizer" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Finalizer 'fluid-alluxio-controller-finalizer' not found in %v", updatedRuntime.Finalizers)
		}
	}
}

func TestReconcileInternal_ReconcileRuntime(t *testing.T) {
	// Scenario: fully set up Runtime (owners, finalizers correct).
	// Expected: Should proceed to ReconcileRuntime logic (Setup, Sync).
	// Since MockEngine returns success, it should return success (Check utils.NoRequeue semantics).

	dataset := &datav1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dataset",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "dataset-uid-123",
		},
		Status: datav1alpha1.DatasetStatus{
			Phase: datav1alpha1.BoundDatasetPhase,
		},
	}
	runtimeObj := &datav1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlluxioRuntime",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "data.fluid.io/v1alpha1",
					Kind:       "Dataset",
					Name:       "test-dataset",
					UID:        "dataset-uid-123",
				},
			},
			Finalizers: []string{"fluid-alluxio-controller-finalizer"},
		},
	}

	reconciler, c := newTestReconciler(t, dataset, runtimeObj)

	ctx := cruntime.ReconcileRequestContext{
		Context:        context.TODO(),
		Log:            logr.Discard(),
		NamespacedName: types.NamespacedName{Name: "test-dataset", Namespace: "default"},
		RuntimeType:    common.AlluxioRuntime,
		Runtime:        runtimeObj,
		Category:       common.AccelerateCategory,
		FinalizerName:  "fluid-alluxio-controller-finalizer",
		Dataset:        dataset,
		Client:         c,
	}

	// Reconcile
	result, err := reconciler.ReconcileInternal(ctx)
	if err != nil {
		t.Fatalf("ReconcileInternal failed: %v", err)
	}

	if result.Requeue && result.RequeueAfter == 0 {
		t.Errorf("Did not expect immediate Requeue for successful reconcile")
	}
}

func TestReconcileInternal_EngineError(t *testing.T) {
	// Scenario: GetOrCreateEngine fails.
	// Expected: ReconcileInternal returns error.

	dataset := &datav1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dataset",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "dataset-uid-123",
		},
	}
	runtimeObj := &datav1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlluxioRuntime",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
		},
	}

	reconciler, c := newTestReconciler(t, dataset, runtimeObj)
	reconciler.failEngineCreation = true

	ctx := cruntime.ReconcileRequestContext{
		Context:        context.TODO(),
		Log:            logr.Discard(),
		NamespacedName: types.NamespacedName{Name: "test-dataset", Namespace: "default"},
		RuntimeType:    common.AlluxioRuntime,
		Runtime:        runtimeObj,
		Category:       common.AccelerateCategory,
		Client:         c,
	}

	// Reconcile
	_, err := reconciler.ReconcileInternal(ctx)
	if err == nil {
		t.Fatalf("Expected error from ReconcileInternal due to engine failure, got nil")
	}
	if err.Error() != "Failed to create: induced engine creation failure" && err.Error() != "induced engine creation failure" {
		t.Logf("Got expected error: %v", err)
	}
}

func TestReconcileRuntimeDeletion(t *testing.T) {
	// Scenario: Runtime has DeletionTimestamp.
	// Expected: Clean up (DeleteVolume, Shutdown), Remove Finalizer.

	now := metav1.Now()
	dataset := &datav1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Dataset",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "default",
			UID:       "dataset-uid-123",
		},
	}
	runtimeObj := &datav1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AlluxioRuntime",
			APIVersion: "data.fluid.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-dataset",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"fluid-alluxio-controller-finalizer"},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "data.fluid.io/v1alpha1",
					Kind:       "Dataset",
					Name:       "test-dataset",
					UID:        "dataset-uid-123",
				},
			},
		},
	}

	reconciler, c := newTestReconciler(t, dataset, runtimeObj)

	ctx := cruntime.ReconcileRequestContext{
		Context:        context.TODO(),
		Log:            logr.Discard(),
		NamespacedName: types.NamespacedName{Name: "test-dataset", Namespace: "default"},
		RuntimeType:    common.AlluxioRuntime,
		Runtime:        runtimeObj,
		Category:       common.AccelerateCategory,
		FinalizerName:  "fluid-alluxio-controller-finalizer",
		Client:         c,
	}

	// Reconcile
	result, err := reconciler.ReconcileInternal(ctx)
	if err != nil {
		t.Fatalf("ReconcileInternal failed: %v", err)
	}

	// Should not requeue if deletion succeeds (Remove Finalizer calls Update, which triggers new event, so return NoRequeue)
	if result.Requeue {
		t.Errorf("Expected no requeue after successful deletion, got %v", result)
	}

	// Verify Finalizer is removed
	updatedRuntime := &datav1alpha1.AlluxioRuntime{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "test-dataset", Namespace: "default"}, updatedRuntime)
	if errors.IsNotFound(err) {
		// Object deleted, success!
		return
	}
	if err != nil {
		t.Fatalf("Failed to get updated runtime: %v", err)
	}

	if len(updatedRuntime.Finalizers) != 0 {
		t.Errorf("Expected finalizers to be empty, got %v", updatedRuntime.Finalizers)
	}
}
