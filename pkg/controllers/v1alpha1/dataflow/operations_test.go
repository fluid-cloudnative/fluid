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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	backupName  = "backup-1"
	loadName    = "load-1"
	migrateName = "migrate-1"
	processName = "process-1"

	contextWhenPrecedingOperationIsComplete               = "when preceding operation is complete"
	contextWhenWaitingForOperationCompleteIsAlreadyFalse  = "when WaitingFor.OperationComplete is already false"
	itShouldClearWaitingForOperationCompleteAndNotRequeue = "should clear WaitingFor.OperationComplete and not requeue"
	itShouldRecordWarningAndRequeue                       = "should record a warning and requeue"
	itShouldSkipReconcilingAndNotRequeue                  = "should skip reconciling and not requeue"
	itShouldSucceedWithoutCallingStatusUpdate             = "should succeed without calling Status().Update"
)

// makeTestCtx creates a reconcileRequestContext with a fake client seeded with objs.
func makeTestCtx(s *runtime.Scheme, name, namespace string, objs ...runtime.Object) reconcileRequestContext {
	if s == nil {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	}
	return reconcileRequestContext{
		Context:        context.TODO(),
		NamespacedName: types.NamespacedName{Name: name, Namespace: namespace},
		Client:         fake.NewFakeClientWithScheme(s, objs...),
		Log:            logf.Log.WithName("dataflow-ops-test"),
		Recorder:       record.NewFakeRecorder(32),
	}
}

var _ = Describe("reconcileOperationDataFlow", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when the preceding operation is not found", func() {
		It("should record a warning event and request requeue", func() {
			// Build a DataLoad as the waiting object; preceding DataLoad does not exist.
			waitingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "waiting", Namespace: namespace},
				Spec: datav1alpha1.DataLoadSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "nonexistent",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "waiting", Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, waitingLoad),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())

			// Expect a warning event to be recorded.
			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationNotFound)))
		})
	})

	Context("when the preceding operation is not yet complete", func() {
		It("should record a normal waiting event and request requeue", func() {
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseExecuting,
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

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: "waiting", Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, precedingLoad, waitingLoad),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())

			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationWaiting)))
		})
	})

	Context("when the preceding operation is complete", func() {
		It("should clear WaitingFor.OperationComplete and not requeue", func() {
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
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

			ctx := makeTestCtx(s, "waiting", namespace, precedingLoad, waitingLoad)

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			updated := &datav1alpha1.DataLoad{}
			Expect(ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "waiting", Namespace: namespace}, updated)).To(Succeed())
			Expect(updated.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updated.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})

	Context("when the waiting DataLoad is not found", func() {
		It(itShouldSkipReconcilingAndNotRequeue, func() {
			ctx := makeTestCtx(s, "missing", namespace)

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataBackup", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when the waiting DataBackup is not found", func() {
		It(itShouldSkipReconcilingAndNotRequeue, func() {
			ctx := makeTestCtx(s, "missing", namespace)

			needRequeue, err := reconcileDataBackup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})

	Context(contextWhenPrecedingOperationIsComplete, func() {
		It(itShouldClearWaitingForOperationCompleteAndNotRequeue, func() {
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: loadName, Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{Name: backupName, Namespace: namespace},
				Spec: datav1alpha1.DataBackupSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: loadName,
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			ctx := makeTestCtx(s, backupName, namespace, precedingLoad, waitingBackup)

			needRequeue, err := reconcileDataBackup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			updated := &datav1alpha1.DataBackup{}
			Expect(ctx.Client.Get(context.TODO(), types.NamespacedName{Name: backupName, Namespace: namespace}, updated)).To(Succeed())
			Expect(updated.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updated.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})

	Context("when preceding operation is not complete", func() {
		It("should request requeue and record a waiting event", func() {
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: loadName, Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhasePending,
				},
			}
			waitingBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{Name: backupName, Namespace: namespace},
				Spec: datav1alpha1.DataBackupSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: loadName,
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: backupName, Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, precedingLoad, waitingBackup),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataBackup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())

			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationWaiting)))
		})
	})
})

var _ = Describe("reconcileDataMigrate", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when the waiting DataMigrate is not found", func() {
		It(itShouldSkipReconcilingAndNotRequeue, func() {
			ctx := makeTestCtx(s, "missing", namespace)

			needRequeue, err := reconcileDataMigrate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})

	Context(contextWhenPrecedingOperationIsComplete, func() {
		It(itShouldClearWaitingForOperationCompleteAndNotRequeue, func() {
			precedingBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{Name: backupName, Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{Name: migrateName, Namespace: namespace},
				Spec: datav1alpha1.DataMigrateSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataBackup",
							Name: backupName,
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			ctx := makeTestCtx(s, migrateName, namespace, precedingBackup, waitingMigrate)

			needRequeue, err := reconcileDataMigrate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			updated := &datav1alpha1.DataMigrate{}
			Expect(ctx.Client.Get(context.TODO(), types.NamespacedName{Name: migrateName, Namespace: namespace}, updated)).To(Succeed())
			Expect(updated.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updated.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataProcess", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when the waiting DataProcess is not found", func() {
		It(itShouldSkipReconcilingAndNotRequeue, func() {
			ctx := makeTestCtx(s, "missing", namespace)

			needRequeue, err := reconcileDataProcess(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})

	Context(contextWhenPrecedingOperationIsComplete, func() {
		It(itShouldClearWaitingForOperationCompleteAndNotRequeue, func() {
			precedingMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{Name: migrateName, Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{Name: processName, Namespace: namespace},
				Spec: datav1alpha1.DataProcessSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataMigrate",
							Name: migrateName,
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			ctx := makeTestCtx(s, processName, namespace, precedingMigrate, waitingProcess)

			needRequeue, err := reconcileDataProcess(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())

			updated := &datav1alpha1.DataProcess{}
			Expect(ctx.Client.Get(context.TODO(), types.NamespacedName{Name: processName, Namespace: namespace}, updated)).To(Succeed())
			Expect(updated.Status.WaitingFor.OperationComplete).NotTo(BeNil())
			Expect(*updated.Status.WaitingFor.OperationComplete).To(BeFalse())
		})
	})

	Context("when preceding operation is not complete", func() {
		It("should request requeue and record a waiting event", func() {
			precedingMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{Name: migrateName, Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseExecuting,
				},
			}
			waitingProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{Name: processName, Namespace: namespace},
				Spec: datav1alpha1.DataProcessSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataMigrate",
							Name: migrateName,
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: processName, Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, precedingMigrate, waitingProcess),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataProcess(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())

			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationWaiting)))
		})
	})
})

var _ = Describe("reconcileOperationDataFlow with cross-namespace preceding op", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when RunAfter specifies a different namespace", func() {
		It("should look up the preceding op in the specified namespace and not requeue when complete", func() {
			// Preceding load in a different namespace.
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: "other-ns"},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "waiting", Namespace: namespace},
				Spec: datav1alpha1.DataLoadSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind:      "DataLoad",
							Name:      "preceding",
							Namespace: "other-ns",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			ctx := makeTestCtx(s, "waiting", namespace, precedingLoad, waitingLoad)

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileOperationDataFlow with unsupported RunAfter kind", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when RunAfter.Kind is not supported", func() {
		It("should return an error and request requeue", func() {
			waitingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "waiting", Namespace: namespace},
				Spec: datav1alpha1.DataLoadSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "UnknownKind",
							Name: "some-op",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			ctx := makeTestCtx(s, "waiting", namespace, waitingLoad)

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get preceding operation status"))
			Expect(needRequeue).To(BeTrue())
		})
	})
})

var _ = Describe("reconcileDataLoad updateStatusFn no-op path", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("when WaitingFor.OperationComplete is already false when updateStatusFn runs", func() {
		It("should succeed without calling Status().Update (reflect.DeepEqual skips update)", func() {
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			// OperationComplete is already false — updateStatusFn will find no delta and skip the update.
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
						OperationComplete: ptr.To(false),
					},
				},
			}

			ctx := makeTestCtx(s, "waiting", namespace, precedingLoad, waitingLoad)

			needRequeue, err := reconcileDataLoad(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataBackup no-op path", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context(contextWhenWaitingForOperationCompleteIsAlreadyFalse, func() {
		It(itShouldSucceedWithoutCallingStatusUpdate, func() {
			precedingLoad := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{Name: backupName, Namespace: namespace},
				Spec: datav1alpha1.DataBackupSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "preceding",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(false),
					},
				},
			}

			ctx := makeTestCtx(s, backupName, namespace, precedingLoad, waitingBackup)

			needRequeue, err := reconcileDataBackup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataMigrate no-op path", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context(contextWhenWaitingForOperationCompleteIsAlreadyFalse, func() {
		It(itShouldSucceedWithoutCallingStatusUpdate, func() {
			precedingBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{Name: migrateName, Namespace: namespace},
				Spec: datav1alpha1.DataMigrateSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataBackup",
							Name: "preceding",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(false),
					},
				},
			}

			ctx := makeTestCtx(s, migrateName, namespace, precedingBackup, waitingMigrate)

			needRequeue, err := reconcileDataMigrate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataProcess no-op path", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context(contextWhenWaitingForOperationCompleteIsAlreadyFalse, func() {
		It(itShouldSucceedWithoutCallingStatusUpdate, func() {
			precedingMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{Name: "preceding", Namespace: namespace},
				Status: datav1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}
			waitingProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{Name: processName, Namespace: namespace},
				Spec: datav1alpha1.DataProcessSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataMigrate",
							Name: "preceding",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(false),
					},
				},
			}

			ctx := makeTestCtx(s, processName, namespace, precedingMigrate, waitingProcess)

			needRequeue, err := reconcileDataProcess(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeFalse())
		})
	})
})

var _ = Describe("reconcileDataMigrate and reconcileDataProcess not-found events", func() {

	var (
		s         *runtime.Scheme
		namespace = "default"
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(s)
	})

	Context("reconcileDataMigrate with preceding op not found", func() {
		It(itShouldRecordWarningAndRequeue, func() {
			waitingMigrate := &datav1alpha1.DataMigrate{
				ObjectMeta: metav1.ObjectMeta{Name: migrateName, Namespace: namespace},
				Spec: datav1alpha1.DataMigrateSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "nonexistent",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: migrateName, Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, waitingMigrate),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataMigrate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())
			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationNotFound)))
		})
	})

	Context("reconcileDataProcess with preceding op not found", func() {
		It(itShouldRecordWarningAndRequeue, func() {
			waitingProcess := &datav1alpha1.DataProcess{
				ObjectMeta: metav1.ObjectMeta{Name: processName, Namespace: namespace},
				Spec: datav1alpha1.DataProcessSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataBackup",
							Name: "nonexistent",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: processName, Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, waitingProcess),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataProcess(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())
			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationNotFound)))
		})
	})

	Context("reconcileDataBackup with preceding op not found", func() {
		It(itShouldRecordWarningAndRequeue, func() {
			waitingBackup := &datav1alpha1.DataBackup{
				ObjectMeta: metav1.ObjectMeta{Name: backupName, Namespace: namespace},
				Spec: datav1alpha1.DataBackupSpec{
					RunAfter: &datav1alpha1.OperationRef{
						ObjectRef: datav1alpha1.ObjectRef{
							Kind: "DataLoad",
							Name: "nonexistent",
						},
					},
				},
				Status: datav1alpha1.OperationStatus{
					WaitingFor: datav1alpha1.WaitingStatus{
						OperationComplete: ptr.To(true),
					},
				},
			}

			recorder := record.NewFakeRecorder(32)
			ctx := reconcileRequestContext{
				Context:        context.TODO(),
				NamespacedName: types.NamespacedName{Name: backupName, Namespace: namespace},
				Client:         fake.NewFakeClientWithScheme(s, waitingBackup),
				Log:            logf.Log.WithName("test"),
				Recorder:       recorder,
			}

			needRequeue, err := reconcileDataBackup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(needRequeue).To(BeTrue())
			Eventually(recorder.Events).Should(Receive(ContainSubstring(common.DataOperationNotFound)))
		})
	})
})
