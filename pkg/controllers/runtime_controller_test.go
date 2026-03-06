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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	fakeutil "github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultNamespace           = "default"
	demoRuntimeName            = "demo-runtime"
	demoDatasetName            = "demo-dataset"
	demoPhysicalDatasetMount   = "dataset://demo/physical-dataset"
	runtimeProtectionFinalizer = "fluid.io/runtime-protection"
)

var _ = Describe("RuntimeReconciler", func() {
	Describe("GetRuntimeObjectMeta", func() {
		It("returns the runtime metadata", func() {
			r := &RuntimeReconciler{}
			ctx := cruntime.ReconcileRequestContext{Runtime: newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)}

			got, err := r.GetRuntimeObjectMeta(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(got.GetName()).To(Equal(demoRuntimeName))
		})

		It("returns an error when runtime is nil", func() {
			r := &RuntimeReconciler{}

			got, err := r.GetRuntimeObjectMeta(cruntime.ReconcileRequestContext{})

			Expect(err).To(MatchError("runtime is nil"))
			Expect(got).To(BeNil())
		})
	})

	Describe("GetDataset", func() {
		It("returns an existing dataset", func() {
			r := newTestRuntimeReconciler(newTestDataset(defaultNamespace, demoDatasetName))
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: defaultNamespace, Name: demoDatasetName},
			}

			got, err := r.GetDataset(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(got.Name).To(Equal(demoDatasetName))
		})

		It("returns a not found error when the dataset is missing", func() {
			r := newTestRuntimeReconciler()
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Namespace: "default", Name: "missing-dataset"},
			}

			got, err := r.GetDataset(ctx)

			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			Expect(got).To(BeNil())
		})
	})

	Describe("CheckIfReferenceDatasetIsSupported", func() {
		It("allows a normal dataset for cache runtimes", func() {
			r := &RuntimeReconciler{}

			supported, reason := r.CheckIfReferenceDatasetIsSupported(cruntime.ReconcileRequestContext{
				Dataset:     newTestDataset(defaultNamespace, demoDatasetName),
				RuntimeType: common.AlluxioRuntime,
			})

			Expect(supported).To(BeTrue())
			Expect(reason).To(BeEmpty())
		})

		It("allows a reference dataset for thin runtimes", func() {
			r := &RuntimeReconciler{}

			supported, reason := r.CheckIfReferenceDatasetIsSupported(cruntime.ReconcileRequestContext{
				Dataset: newTestDatasetWithMounts(defaultNamespace, demoDatasetName, []datav1alpha1.Mount{{
					MountPoint: demoPhysicalDatasetMount,
				}}),
				RuntimeType: common.ThinRuntime,
			})

			Expect(supported).To(BeTrue())
			Expect(reason).To(BeEmpty())
		})

		It("rejects a reference dataset for non-thin runtimes", func() {
			r := &RuntimeReconciler{}

			supported, reason := r.CheckIfReferenceDatasetIsSupported(cruntime.ReconcileRequestContext{
				Dataset: newTestDatasetWithMounts(defaultNamespace, demoDatasetName, []datav1alpha1.Mount{{
					MountPoint: demoPhysicalDatasetMount,
				}}),
				RuntimeType: common.AlluxioRuntime,
			})

			Expect(supported).To(BeFalse())
			Expect(reason).NotTo(BeEmpty())
		})
	})

	Describe("AddOwnerAndRequeue", func() {
		It("adds the dataset as an owner reference and requeues", func() {
			dataset := newTestDataset(defaultNamespace, demoDatasetName)
			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			r := newTestRuntimeReconciler(dataset, runtimeObj)

			result, err := r.AddOwnerAndRequeue(newTestRequestContext(r.Client, runtimeObj, dataset), dataset)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeTrue())

			updated := &datav1alpha1.AlluxioRuntime{}
			Expect(r.Get(context.Background(), types.NamespacedName{Namespace: runtimeObj.Namespace, Name: runtimeObj.Name}, updated)).To(Succeed())
			Expect(updated.OwnerReferences).To(HaveLen(1))
			Expect(updated.OwnerReferences[0].Name).To(Equal(dataset.Name))
			Expect(updated.OwnerReferences[0].Kind).To(Equal(dataset.Kind))
			Expect(updated.OwnerReferences[0].UID).To(Equal(dataset.UID))
		})
	})

	Describe("AddFinalizerAndRequeue", func() {
		It("adds the finalizer and requeues", func() {

			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			r := newTestRuntimeReconciler(runtimeObj)

			result, err := r.AddFinalizerAndRequeue(newTestRequestContext(r.Client, runtimeObj, nil), runtimeProtectionFinalizer)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeTrue())

			updated := &datav1alpha1.AlluxioRuntime{}
			Expect(r.Get(context.Background(), types.NamespacedName{Namespace: runtimeObj.Namespace, Name: runtimeObj.Name}, updated)).To(Succeed())
			Expect(updated.Finalizers).To(ContainElement(runtimeProtectionFinalizer))
		})
	})

	Describe("ReportDatasetNotReadyCondition", func() {
		It("reports the dataset not ready condition", func() {
			dataset := newTestDataset(defaultNamespace, demoDatasetName)
			r := newTestRuntimeReconciler(dataset)
			notReadyErr := errors.New("setup failed")

			err := r.ReportDatasetNotReadyCondition(newTestRequestContext(r.Client, nil, dataset), notReadyErr)

			Expect(err).NotTo(HaveOccurred())

			updated := &datav1alpha1.Dataset{}
			Expect(r.Get(context.Background(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, updated)).To(Succeed())
			Expect(updated.Status.Conditions).To(HaveLen(1))
			Expect(updated.Status.Conditions[0].Type).To(Equal(datav1alpha1.DatasetNotReady))
			Expect(updated.Status.Conditions[0].Reason).To(Equal(datav1alpha1.DatasetFailedToSetupReason))
			Expect(updated.Status.Conditions[0].Message).To(Equal(notReadyErr.Error()))
			Expect(updated.Status.Conditions[0].Status).To(Equal(corev1.ConditionTrue))
		})
	})

	Describe("NewRuntimeReconciler", func() {
		It("stores the provided dependencies", func() {
			clientObj := newTestRuntimeReconciler().Client
			recorder := record.NewFakeRecorder(8)
			impl := &testRuntimeReconcilerImplement{}

			r := NewRuntimeReconciler(impl, clientObj, fakeutil.NullLogger(), recorder)

			Expect(r.Client).To(Equal(clientObj))
			Expect(r.Recorder).To(Equal(recorder))
			Expect(r.implement).To(Equal(impl))
		})
	})

	Describe("ForgetMetrics", func() {
		It("forgets dataset and runtime metrics", func() {
			r := &RuntimeReconciler{}
			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			dataset := newTestDataset(defaultNamespace, demoRuntimeName)
			ctx := newTestRequestContext(nil, runtimeObj, dataset)

			runtimeMetricBefore := metrics.GetOrCreateRuntimeMetrics(runtimeObj.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name)
			datasetMetricBefore := metrics.GetOrCreateDatasetMetrics(ctx.Namespace, ctx.Name)

			Expect(func() { r.ForgetMetrics(ctx) }).NotTo(Panic())

			runtimeMetricAfter := metrics.GetOrCreateRuntimeMetrics(runtimeObj.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name)
			datasetMetricAfter := metrics.GetOrCreateDatasetMetrics(ctx.Namespace, ctx.Name)

			Expect(runtimeMetricAfter).NotTo(BeIdenticalTo(runtimeMetricBefore))
			Expect(datasetMetricAfter).NotTo(BeIdenticalTo(datasetMetricBefore))
		})
	})

	Describe("ReconcileInternal", func() {
		It("returns an error when the runtime is missing", func() {
			r := newTestRuntimeReconcilerWithImplement(nil)

			result, err := r.ReconcileInternal(cruntime.ReconcileRequestContext{Context: context.Background()})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to find the runtime"))
			Expect(result).To(Equal(ctrl.Result{}))
		})

		It("requeues when the dataset is not found", func() {
			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			impl := &testRuntimeReconcilerImplement{
				getOrCreateEngine: func(cruntime.ReconcileRequestContext) (base.Engine, error) {
					return &testEngine{}, nil
				},
				getRuntimeObjectMeta: func(ctx cruntime.ReconcileRequestContext) (metav1.Object, error) {
					return runtimeObj, nil
				},
			}
			r := newTestRuntimeReconcilerWithImplement(impl, runtimeObj)

			result, err := r.ReconcileInternal(newTestRequestContext(r.Client, runtimeObj, nil))

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(5 * time.Second))
		})

		It("delegates runtime deletion and removes the engine on error", func() {
			dataset := newBoundDataset(defaultNamespace, demoRuntimeName)
			runtimeObj := newDeletingRuntime(defaultNamespace, demoRuntimeName, runtimeProtectionFinalizer)
			deletionErr := errors.New("delete failed")
			engine := &testEngine{}
			impl := &testRuntimeReconcilerImplement{
				getOrCreateEngine: func(cruntime.ReconcileRequestContext) (base.Engine, error) {
					return engine, nil
				},
				reconcileRuntimeDeletion: func(gotEngine base.Engine, gotCtx cruntime.ReconcileRequestContext) (ctrl.Result, error) {
					Expect(gotEngine).To(Equal(engine))
					Expect(gotCtx.Dataset.Name).To(Equal(dataset.Name))
					return ctrl.Result{RequeueAfter: time.Second}, deletionErr
				},
			}
			r := newTestRuntimeReconcilerWithImplement(impl, dataset, runtimeObj)

			result, err := r.ReconcileInternal(newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).To(MatchError(deletionErr))
			Expect(result.RequeueAfter).To(Equal(time.Second))
			Expect(impl.removeEngineCalls).To(Equal(1))
		})

		It("requeues when reference datasets are not supported by the runtime", func() {
			dataset := newBoundDatasetWithMounts(defaultNamespace, demoRuntimeName, []datav1alpha1.Mount{{MountPoint: demoPhysicalDatasetMount}})
			runtimeObj := newRuntimeWithOwnersAndFinalizers(defaultNamespace, demoRuntimeName, dataset, runtimeProtectionFinalizer)
			impl := &testRuntimeReconcilerImplement{
				getOrCreateEngine: func(cruntime.ReconcileRequestContext) (base.Engine, error) {
					return &testEngine{}, nil
				},
			}
			r := newTestRuntimeReconcilerWithImplement(impl, dataset, runtimeObj)

			result, err := r.ReconcileInternal(newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(20 * time.Second))
			Expect(impl.reconcileRuntimeCalls).To(Equal(0))
		})

		It("delegates runtime reconciliation when setup prerequisites are satisfied", func() {
			dataset := newBoundDataset(defaultNamespace, demoRuntimeName)
			runtimeObj := newRuntimeWithOwnersAndFinalizers(defaultNamespace, demoRuntimeName, dataset, runtimeProtectionFinalizer)
			engine := &testEngine{}
			impl := &testRuntimeReconcilerImplement{
				getOrCreateEngine: func(cruntime.ReconcileRequestContext) (base.Engine, error) {
					return engine, nil
				},
				reconcileRuntime: func(gotEngine base.Engine, gotCtx cruntime.ReconcileRequestContext) (ctrl.Result, error) {
					Expect(gotEngine).To(Equal(engine))
					Expect(gotCtx.Dataset.Name).To(Equal(dataset.Name))
					return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
				},
			}
			r := newTestRuntimeReconcilerWithImplement(impl, dataset, runtimeObj)

			result, err := r.ReconcileInternal(newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(2 * time.Second))
			Expect(impl.reconcileRuntimeCalls).To(Equal(1))
		})
	})

	Describe("ReconcileRuntime", func() {
		It("requeues when validation temporarily fails", func() {
			dataset := newBoundDataset(defaultNamespace, demoRuntimeName)
			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			r := newTestRuntimeReconciler(runtimeObj, dataset)
			engine := &testEngine{validateErr: fluiderrs.NewTemporaryValidationFailed("spec invalid")}

			result, err := r.ReconcileRuntime(engine, newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(20 * time.Second))
			Expect(engine.validateCalls).To(Equal(5))
		})

		It("reports the dataset not ready when setup is not ready", func() {
			dataset := newBoundDataset(defaultNamespace, demoRuntimeName)
			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			r := newTestRuntimeReconciler(runtimeObj, dataset)
			engine := &testEngine{setupErr: errors.New("setup failed")}

			result, err := r.ReconcileRuntime(engine, newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(20 * time.Second))

			updated := &datav1alpha1.Dataset{}
			Expect(r.Get(context.Background(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, updated)).To(Succeed())
			Expect(updated.Status.Conditions).To(ContainElement(HaveField("Type", Equal(datav1alpha1.DatasetNotReady))))
		})

		It("creates the volume, syncs the engine, and requeues on success", func() {
			dataset := newReadyDataset(defaultNamespace, demoRuntimeName)
			runtimeObj := newTestAlluxioRuntime(defaultNamespace, demoRuntimeName)
			r := newTestRuntimeReconciler(runtimeObj, dataset)
			engine := &testEngine{}

			result, err := r.ReconcileRuntime(engine, newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(engine.setupCalls).To(Equal(0))
			Expect(engine.createVolumeCalls).To(Equal(1))
			Expect(engine.syncCalls).To(Equal(1))
			Expect(result.RequeueAfter).To(BeNumerically(">", 0))
		})
	})

	Describe("ReconcileRuntimeDeletion", func() {
		It("requeues when deleting the volume fails", func() {
			dataset := newBoundDataset(defaultNamespace, demoRuntimeName)
			runtimeObj := newDeletingRuntime(defaultNamespace, demoRuntimeName, runtimeProtectionFinalizer)
			r := newTestRuntimeReconciler(runtimeObj, dataset)
			engine := &testEngine{deleteVolumeErr: errors.New("volume delete failed")}

			result, err := r.ReconcileRuntimeDeletion(engine, newTestRequestContext(r.Client, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(20 * time.Second))
			Expect(engine.shutdownCalls).To(Equal(0))
		})

		It("resets the dataset state and removes the runtime finalizer", func() {
			dataset := newBoundDataset(defaultNamespace, demoRuntimeName)
			dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
			dataset.Status.UfsTotal = "123"
			dataset.Status.Conditions = []datav1alpha1.DatasetCondition{{Type: datav1alpha1.DatasetReady}}
			dataset.Status.CacheStates = common.CacheStateList{common.Cached: "9"}
			dataset.Status.Runtimes = []datav1alpha1.Runtime{{Name: demoRuntimeName, Namespace: defaultNamespace, Category: common.AccelerateCategory}}
			dataset.Status.HCFSStatus = &datav1alpha1.HCFSStatus{}
			dataset.Status.FileNum = "9"

			runtimeObj := newDeletingRuntime(defaultNamespace, demoRuntimeName, runtimeProtectionFinalizer)
			impl := &testRuntimeReconcilerImplement{}
			baseClient := newTestClient(runtimeObj, dataset)
			recordingClient := &runtimeUpdateRecordingClient{Client: baseClient}
			r := &RuntimeReconciler{
				Client:    recordingClient,
				Log:       fakeutil.NullLogger(),
				Recorder:  record.NewFakeRecorder(32),
				implement: impl,
			}

			result, err := r.ReconcileRuntimeDeletion(&testEngine{}, newTestRequestContext(recordingClient, runtimeObj, dataset))

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
			Expect(impl.removeEngineCalls).To(Equal(1))

			updatedDataset := &datav1alpha1.Dataset{}
			Expect(r.Get(context.Background(), types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}, updatedDataset)).To(Succeed())
			Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.NotBoundDatasetPhase))
			Expect(updatedDataset.Status.UfsTotal).To(BeEmpty())
			Expect(updatedDataset.Status.Conditions).To(BeEmpty())
			Expect(updatedDataset.Status.CacheStates).To(BeEmpty())
			Expect(updatedDataset.Status.Runtimes).To(BeEmpty())
			Expect(updatedDataset.Status.HCFSStatus).To(BeNil())
			Expect(updatedDataset.Status.FileNum).To(BeEmpty())

			Expect(recordingClient.updatedRuntime).NotTo(BeNil())
			Expect(recordingClient.updatedRuntime.Finalizers).NotTo(ContainElement(runtimeProtectionFinalizer))
		})
	})
})

func newTestRuntimeReconciler(objects ...runtime.Object) *RuntimeReconciler {
	return newTestRuntimeReconcilerWithImplement(&testRuntimeReconcilerImplement{}, objects...)
}

func newTestClient(objects ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	return fakeutil.NewFakeClientWithScheme(scheme, objects...)
}

func newTestRuntimeReconcilerWithImplement(impl *testRuntimeReconcilerImplement, objects ...runtime.Object) *RuntimeReconciler {
	r := &RuntimeReconciler{
		Client:   newTestClient(objects...),
		Log:      fakeutil.NullLogger(),
		Recorder: record.NewFakeRecorder(32),
	}
	if impl == nil {
		impl = &testRuntimeReconcilerImplement{}
	}
	r.implement = impl

	return r
}

func newTestRequestContext(clientObj client.Client, runtimeObj *datav1alpha1.AlluxioRuntime, dataset *datav1alpha1.Dataset) cruntime.ReconcileRequestContext {
	ctx := cruntime.ReconcileRequestContext{
		Context: context.Background(),
		Client:  clientObj,
		Log:     fakeutil.NullLogger(),
	}

	if runtimeObj != nil {
		ctx.Runtime = runtimeObj
		ctx.NamespacedName = types.NamespacedName{Namespace: runtimeObj.Namespace, Name: runtimeObj.Name}
	}

	if dataset != nil {
		ctx.Dataset = dataset
		if ctx.NamespacedName == (types.NamespacedName{}) {
			ctx.NamespacedName = types.NamespacedName{Namespace: dataset.Namespace, Name: dataset.Name}
		}
	}

	ctx.Category = common.AccelerateCategory
	ctx.RuntimeType = common.AlluxioRuntime
	ctx.FinalizerName = runtimeProtectionFinalizer

	return ctx
}

type runtimeUpdateRecordingClient struct {
	client.Client
	updatedRuntime *datav1alpha1.AlluxioRuntime
}

func (c *runtimeUpdateRecordingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if runtimeObj, ok := obj.(*datav1alpha1.AlluxioRuntime); ok {
		c.updatedRuntime = runtimeObj.DeepCopy()
	}

	return c.Client.Update(ctx, obj, opts...)
}

type testRuntimeReconcilerImplement struct {
	getOrCreateEngine        func(cruntime.ReconcileRequestContext) (base.Engine, error)
	getRuntimeObjectMeta     func(cruntime.ReconcileRequestContext) (metav1.Object, error)
	reconcileRuntime         func(base.Engine, cruntime.ReconcileRequestContext) (ctrl.Result, error)
	reconcileRuntimeDeletion func(base.Engine, cruntime.ReconcileRequestContext) (ctrl.Result, error)
	addFinalizerAndRequeue   func(cruntime.ReconcileRequestContext, string) (ctrl.Result, error)
	removeEngine             func(cruntime.ReconcileRequestContext)

	reconcileRuntimeCalls int
	removeEngineCalls     int
}

func (t *testRuntimeReconcilerImplement) ReconcileRuntimeDeletion(engine base.Engine, ctx cruntime.ReconcileRequestContext) (ctrl.Result, error) {
	if t.reconcileRuntimeDeletion != nil {
		return t.reconcileRuntimeDeletion(engine, ctx)
	}
	return ctrl.Result{}, nil
}

func (t *testRuntimeReconcilerImplement) ReconcileRuntime(engine base.Engine, ctx cruntime.ReconcileRequestContext) (ctrl.Result, error) {
	t.reconcileRuntimeCalls++
	if t.reconcileRuntime != nil {
		return t.reconcileRuntime(engine, ctx)
	}
	return ctrl.Result{}, nil
}

func (t *testRuntimeReconcilerImplement) AddFinalizerAndRequeue(ctx cruntime.ReconcileRequestContext, finalizer string) (ctrl.Result, error) {
	if t.addFinalizerAndRequeue != nil {
		return t.addFinalizerAndRequeue(ctx, finalizer)
	}
	return ctrl.Result{}, nil
}

func (t *testRuntimeReconcilerImplement) GetDataset(cruntime.ReconcileRequestContext) (*datav1alpha1.Dataset, error) {
	return nil, nil
}

func (t *testRuntimeReconcilerImplement) GetOrCreateEngine(ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	if t.getOrCreateEngine != nil {
		return t.getOrCreateEngine(ctx)
	}
	return nil, nil
}

func (t *testRuntimeReconcilerImplement) RemoveEngine(ctx cruntime.ReconcileRequestContext) {
	t.removeEngineCalls++
	if t.removeEngine != nil {
		t.removeEngine(ctx)
	}
}

func (t *testRuntimeReconcilerImplement) GetRuntimeObjectMeta(ctx cruntime.ReconcileRequestContext) (metav1.Object, error) {
	if t.getRuntimeObjectMeta != nil {
		return t.getRuntimeObjectMeta(ctx)
	}
	objectMetaAccessor, ok := ctx.Runtime.(metav1.ObjectMetaAccessor)
	if !ok {
		return nil, errors.New("object is not ObjectMetaAccessor")
	}

	return objectMetaAccessor.GetObjectMeta(), nil
}

func (t *testRuntimeReconcilerImplement) ReconcileInternal(cruntime.ReconcileRequestContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func newTestDataset(namespace, name string) *datav1alpha1.Dataset {
	return newTestDatasetWithMounts(namespace, name, nil)
}

func newTestDatasetWithMounts(namespace, name string, mounts []datav1alpha1.Mount) *datav1alpha1.Dataset {
	return &datav1alpha1.Dataset{
		TypeMeta: metav1.TypeMeta{
			APIVersion: datav1alpha1.GroupVersion.String(),
			Kind:       "Dataset",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			UID:       types.UID(name + "-uid"),
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: mounts,
		},
	}
}

func newTestAlluxioRuntime(namespace, name string) *datav1alpha1.AlluxioRuntime {
	return &datav1alpha1.AlluxioRuntime{
		TypeMeta: metav1.TypeMeta{
			APIVersion: datav1alpha1.GroupVersion.String(),
			Kind:       "AlluxioRuntime",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func newDeletingRuntime(namespace, name, finalizer string) *datav1alpha1.AlluxioRuntime {
	runtimeObj := newTestAlluxioRuntime(namespace, name)
	runtimeObj.Finalizers = []string{finalizer}
	runtimeObj.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	return runtimeObj
}

func newRuntimeWithOwnersAndFinalizers(namespace, name string, dataset *datav1alpha1.Dataset, finalizer string) *datav1alpha1.AlluxioRuntime {
	runtimeObj := newTestAlluxioRuntime(namespace, name)
	runtimeObj.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}}
	runtimeObj.Finalizers = []string{finalizer}
	return runtimeObj
}

func newBoundDataset(namespace, name string) *datav1alpha1.Dataset {
	dataset := newTestDataset(namespace, name)
	dataset.Status.Runtimes = []datav1alpha1.Runtime{{
		Name:      name,
		Namespace: namespace,
		Category:  common.AccelerateCategory,
	}}
	return dataset
}

func newBoundDatasetWithMounts(namespace, name string, mounts []datav1alpha1.Mount) *datav1alpha1.Dataset {
	dataset := newTestDatasetWithMounts(namespace, name, mounts)
	dataset.Status.Runtimes = []datav1alpha1.Runtime{{
		Name:      name,
		Namespace: namespace,
		Category:  common.AccelerateCategory,
	}}
	return dataset
}

func newReadyDataset(namespace, name string) *datav1alpha1.Dataset {
	dataset := newBoundDataset(namespace, name)
	dataset.Status.Conditions = []datav1alpha1.DatasetCondition{{Type: datav1alpha1.DatasetReady}}
	return dataset
}

type testEngine struct {
	validateErr     error
	setupReady      bool
	setupErr        error
	createVolumeErr error
	deleteVolumeErr error
	syncErr         error
	shutdownErr     error

	validateCalls     int
	setupCalls        int
	createVolumeCalls int
	deleteVolumeCalls int
	shutdownCalls     int
	syncCalls         int
}

func (e *testEngine) ID() string { return "test-engine" }

func (e *testEngine) Shutdown() error {
	e.shutdownCalls++
	return e.shutdownErr
}

func (e *testEngine) Setup(cruntime.ReconcileRequestContext) (bool, error) {
	e.setupCalls++
	return e.setupReady, e.setupErr
}

func (e *testEngine) CreateVolume() error {
	e.createVolumeCalls++
	return e.createVolumeErr
}

func (e *testEngine) DeleteVolume() error {
	e.deleteVolumeCalls++
	return e.deleteVolumeErr
}

func (e *testEngine) Sync(cruntime.ReconcileRequestContext) error {
	e.syncCalls++
	return e.syncErr
}

func (e *testEngine) Validate(cruntime.ReconcileRequestContext) error {
	e.validateCalls++
	return e.validateErr
}

func (e *testEngine) Operate(cruntime.ReconcileRequestContext, *datav1alpha1.OperationStatus, dataoperation.OperationInterface) (ctrl.Result, error) {
	return ctrl.Result{}, fmt.Errorf("not implemented")
}
