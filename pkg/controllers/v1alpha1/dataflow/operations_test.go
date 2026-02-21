/*
Copyright 2025 The Fluid Authors.

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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("reconcileOperationDataFlow", func() {
	var (
		scheme    *runtime.Scheme
		namespace string
		name      string
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		namespace = "default"
		name = "test-dataload"
	})

	Context("when runAfter is nil", func() {
		It("should not panic and should clear the waiting status", func() {
			// Create a DataLoad with WaitingFor.OperationComplete = true but no RunAfter
			// This simulates the case where a user removes RunAfter after it was set
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
					// RunAfter is nil - simulating user removing it after status was set
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			updateStatusCalled := false
			updateStatusFn := func() error {
				updateStatusCalled = true
				// Simulate updating the status
				tmp := &datav1alpha1.DataLoad{}
				err := fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, tmp)
				if err != nil {
					return err
				}
				toUpdate := tmp.DeepCopy()
				toUpdate.Status.WaitingFor.OperationComplete = ptr.To(false)
				return fakeClient.Status().Update(context.TODO(), toUpdate)
			}

			// This should NOT panic even though runAfter is nil
			needRequeue, err := reconcileOperationDataFlow(ctx, dataLoad, nil, dataLoad.Status, updateStatusFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
			Expect(updateStatusCalled).To(BeTrue())

			// Verify the status was updated
			updatedDataLoad := &datav1alpha1.DataLoad{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, updatedDataLoad)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataLoad.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updatedDataLoad.Status.WaitingFor.OperationComplete).To(BeFalse())
		})

		It("should return error and requeue when updateStatusFn fails", func() {
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			// updateStatusFn that always fails
			updateStatusFn := func() error {
				return context.DeadlineExceeded
			}

			needRequeue, err := reconcileOperationDataFlow(ctx, dataLoad, nil, dataLoad.Status, updateStatusFn)

			Expect(err).To(HaveOccurred())
			Expect(needRequeue).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring("failed to clear operation waiting status when runAfter is nil"))
		})
	})

	Context("when runAfter is valid and preceding operation is complete", func() {
		It("should clear waiting status and not requeue", func() {
			// Create preceding DataLoad that is complete
			precedingDataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "preceding-dataload",
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}

			// Create DataLoad waiting for preceding operation
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "preceding-dataload",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad, precedingDataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			updateStatusCalled := false
			updateStatusFn := func() error {
				updateStatusCalled = true
				tmp := &datav1alpha1.DataLoad{}
				err := fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, tmp)
				if err != nil {
					return err
				}
				toUpdate := tmp.DeepCopy()
				toUpdate.Status.WaitingFor.OperationComplete = ptr.To(false)
				return fakeClient.Status().Update(context.TODO(), toUpdate)
			}

			needRequeue, err := reconcileOperationDataFlow(ctx, dataLoad, dataLoad.Spec.RunAfter, dataLoad.Status, updateStatusFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
			Expect(updateStatusCalled).To(BeTrue())
		})
	})

	Context("when runAfter is valid but preceding operation is not complete", func() {
		It("should requeue without error", func() {
			// Create preceding DataLoad that is still executing
			precedingDataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "preceding-dataload",
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseExecuting,
				},
			}

			// Create DataLoad waiting for preceding operation
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "preceding-dataload",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad, precedingDataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			updateStatusCalled := false
			updateStatusFn := func() error {
				updateStatusCalled = true
				return nil
			}

			needRequeue, err := reconcileOperationDataFlow(ctx, dataLoad, dataLoad.Spec.RunAfter, dataLoad.Status, updateStatusFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())
			// updateStatusFn should NOT be called because we're still waiting
			Expect(updateStatusCalled).To(BeFalse())
		})
	})

	Context("when runAfter references non-existent operation", func() {
		It("should requeue without error", func() {
			// Create DataLoad referencing a non-existent preceding operation
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "non-existent-dataload",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			updateStatusFn := func() error {
				return nil
			}

			needRequeue, err := reconcileOperationDataFlow(ctx, dataLoad, dataLoad.Spec.RunAfter, dataLoad.Status, updateStatusFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())
		})
	})

	Context("when runAfter has custom namespace", func() {
		It("should use the specified namespace to find preceding operation", func() {
			otherNamespace := "other-namespace"

			// Create preceding DataLoad in different namespace that is complete
			precedingDataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "preceding-dataload",
					Namespace: otherNamespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: otherNamespace,
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}

			// Create DataLoad waiting for preceding operation in different namespace
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind:      "DataLoad",
							Name:      "preceding-dataload",
							Namespace: otherNamespace,
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad, precedingDataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			updateStatusCalled := false
			updateStatusFn := func() error {
				updateStatusCalled = true
				return nil
			}

			needRequeue, err := reconcileOperationDataFlow(ctx, dataLoad, dataLoad.Spec.RunAfter, dataLoad.Status, updateStatusFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
			Expect(updateStatusCalled).To(BeTrue())
		})
	})
})

var _ = Describe("reconcileDataLoad", func() {
	var (
		scheme    *runtime.Scheme
		namespace string
		name      string
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		namespace = "default"
		name = "test-dataload"
	})

	Context("when DataLoad does not exist", func() {
		It("should return without error and not requeue", func() {
			fakeClient := fake.NewFakeClientWithScheme(scheme)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataLoad(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})

	Context("when DataLoad exists with nil RunAfter", func() {
		It("should not panic and clear waiting status", func() {
			// This test verifies the fix for the nil pointer panic
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: namespace,
					},
					// RunAfter is nil - this is the bug scenario
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataLoad)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			// This should NOT panic
			needRequeue, err := reconcileDataLoad(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			// Verify the waiting status was cleared
			updatedDataLoad := &datav1alpha1.DataLoad{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, updatedDataLoad)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataLoad.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updatedDataLoad.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataMigrate", func() {
	var (
		scheme    *runtime.Scheme
		namespace string
		name      string
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		namespace = "default"
		name = "test-datamigrate"
	})

	Context("when DataMigrate exists with nil RunAfter", func() {
		It("should not panic and clear waiting status", func() {
			dataMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataMigrateSpec{
					// RunAfter is nil
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataMigrate)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			// This should NOT panic
			needRequeue, err := reconcileDataMigrate(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			// Verify the waiting status was cleared
			updatedDataMigrate := &datav1alpha1.DataMigrate{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, updatedDataMigrate)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataMigrate.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updatedDataMigrate.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataBackup", func() {
	var (
		scheme    *runtime.Scheme
		namespace string
		name      string
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		namespace = "default"
		name = "test-databackup"
	})

	Context("when DataBackup exists with nil RunAfter", func() {
		It("should not panic and clear waiting status", func() {
			dataBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataBackupSpec{
					// RunAfter is nil
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataBackup)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			// This should NOT panic
			needRequeue, err := reconcileDataBackup(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			// Verify the waiting status was cleared
			updatedDataBackup := &datav1alpha1.DataBackup{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, updatedDataBackup)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataBackup.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updatedDataBackup.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataProcess", func() {
	var (
		scheme    *runtime.Scheme
		namespace string
		name      string
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		namespace = "default"
		name = "test-dataprocess"
	})

	Context("when DataProcess exists with nil RunAfter", func() {
		It("should not panic and clear waiting status", func() {
			dataProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: datav1alpha1.DataProcessSpec{
					// RunAfter is nil
				},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, dataProcess)
			recorder := record.NewFakeRecorder(10)

			ctx := reconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
				Client:         fakeClient,
				Log:            ctrl.Log.WithName("test"),
				Recorder:       recorder,
			}

			// This should NOT panic
			needRequeue, err := reconcileDataProcess(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			// Verify the waiting status was cleared
			updatedDataProcess := &datav1alpha1.DataProcess{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, updatedDataProcess)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataProcess.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updatedDataProcess.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})
})
