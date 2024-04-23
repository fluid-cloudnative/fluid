/*
Copyright  2020 The Fluid Authors.

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
	"reflect"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/fluid-cloudnative/fluid/pkg/dump"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// var _ RuntimeReconcilerInterface = (*RuntimeReconciler)(nil)

// RuntimeReconciler is the default implementation
type RuntimeReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	// Real implement
	implement RuntimeReconcilerInterface
}

// NewRuntimeReconciler creates the default RuntimeReconciler
func NewRuntimeReconciler(reconciler RuntimeReconcilerInterface, client client.Client, log logr.Logger, recorder record.EventRecorder) *RuntimeReconciler {
	// install gorouting dump generator for all runtime reconciler
	dump.InstallgoroutineDumpGenerator()
	r := &RuntimeReconciler{
		implement: reconciler,
		Client:    client,
		Recorder:  recorder,
		Log:       log,
	}
	return r
}

// ReconcileInternal handles the logic of reconcile runtime
func (r *RuntimeReconciler) ReconcileInternal(ctx cruntime.ReconcileRequestContext) (ctrl.Result, error) {
	// 0.Get the runtime
	runtime := ctx.Runtime
	if runtime == nil {
		return utils.RequeueIfError(fmt.Errorf("failed to find the runtime"))
	}

	// 1.Validate name is prefixed with a number such as "20-hbase".
	if errs := validation.IsDNS1035Label(runtime.GetName()); len(runtime.GetName()) > 0 && len(errs) > 0 {
		err := field.Invalid(field.NewPath("metadata").Child("name"), runtime.GetName(), strings.Join(errs, ","))
		ctx.Log.Error(err, "Failed to validate runtime name")
		r.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to create ddc engine due to error %v", err)
		return utils.RequeueIfError(errors.Wrap(err, "Failed to create"))
	}

	// 2.Get or create the engine
	engine, err := r.implement.GetOrCreateEngine(ctx)
	if err != nil {
		r.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Process Runtime error %v", err)
		return utils.RequeueIfError(errors.Wrap(err, "Failed to create"))
	}

	// 3.Get the ObjectMeta of runtime
	objectMeta, err := r.implement.GetRuntimeObjectMeta(ctx)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	// 4.Get the dataset
	dataset, err := r.GetDataset(ctx)
	if err != nil {
		// r.Recorder.Eventf(ctx.Dataset, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Process Runtime error %v", err)
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("The dataset is not found", "dataset", ctx.NamespacedName)
			dataset = nil
			// return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get the ddc dataset")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get dataset"))
		}
	}
	ctx.Dataset = dataset

	// 5.Reconcile delete the runtime
	// it should be after getting the dataset because need to edit the dataset during deleting
	if !objectMeta.GetDeletionTimestamp().IsZero() {
		result, err := r.implement.ReconcileRuntimeDeletion(engine, ctx)
		if err != nil {
			r.implement.RemoveEngine(ctx)
		}
		return result, err
	}

	if dataset != nil {
		// 6.Add the OwnerReference of runtime and requeue
		if !utils.ContainsOwners(objectMeta.GetOwnerReferences(), dataset) {
			return r.AddOwnerAndRequeue(ctx, dataset)
		}
		if !dataset.CanbeBound(ctx.Name, ctx.Namespace, ctx.Category) {
			ctx.Log.Info("the dataset can't be bound to the runtime, because it's already bound to another runtime ",
				"dataset", dataset.Name)
			r.Recorder.Eventf(runtime, corev1.EventTypeWarning,
				common.ErrorProcessRuntimeReason,
				"the dataset can't be bound to the runtime, because it's already bound to another runtime %s",
				dataset.Name)
			return utils.RequeueAfterInterval(time.Duration(20 * time.Second))
		}
		// check reference dataset support
		isSupport, reason := r.CheckIfReferenceDatasetIsSupported(ctx)
		if !isSupport {
			ctx.Log.Info(reason, "dataset", dataset.Name)
			r.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, reason)
			return utils.RequeueAfterInterval(time.Duration(20 * time.Second))
		}

		// 7. Add Finalizer of runtime and requeue
		if !utils.ContainsString(objectMeta.GetFinalizers(), ctx.FinalizerName) {
			return r.implement.AddFinalizerAndRequeue(ctx, ctx.FinalizerName)
		} else {
			ctx.Log.V(1).Info("The finalizer has been added")
		}
	} else {
		// If dataset is nil, need to wait because the user may have not created dataset
		ctx.Log.Info("No dataset can be bound to the runtime, waiting.")
		r.Recorder.Event(runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "No dataset can be bound to the runtime, waiting.")
		return utils.RequeueAfterInterval(time.Duration(5 * time.Second))
	}

	// 8.Start to reconcile runtime
	return r.implement.ReconcileRuntime(engine, ctx)
}

// ReconcileRuntimeDeletion reconciles runtime deletion
func (r *RuntimeReconciler) ReconcileRuntimeDeletion(engine base.Engine, ctx cruntime.ReconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileRuntimeDeletion")
	log.V(1).Info("process the Runtime Deletion", "Runtime", ctx.NamespacedName)

	// 0. Delete the volume
	err := engine.DeleteVolume()
	if err != nil {
		r.Recorder.Eventf(ctx.Runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to delete volume %v", err)
		// return utils.RequeueIfError(errors.Wrap(err, "Failed to delete volume"))
		log.Error(err, "Failed to delete volume", "Runtime", ctx.NamespacedName)
		return utils.RequeueAfterInterval(time.Duration(20 * time.Second))
	}

	// 1. Delete the implementation of the the runtime
	err = engine.Shutdown()
	if err != nil {
		r.Recorder.Eventf(ctx.Runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to shutdown engine %v", err)
		// return utils.RequeueIfError(errors.Wrap(err, "Failed to shutdown the engine"))
		log.Error(err, "Failed to shutdown the engine", "Runtime", ctx.NamespacedName)
		return utils.RequeueAfterInterval(time.Duration(20 * time.Second))
	}

	// 2. Set the dataset's status as unbound
	r.implement.RemoveEngine(ctx)
	r.ForgetMetrics(ctx)
	// r.removeEngine(engine.ID())
	dataset := ctx.Dataset.DeepCopy()
	if dataset != nil {
		dataset.Status.Phase = datav1alpha1.NotBoundDatasetPhase
		dataset.Status.UfsTotal = ""
		dataset.Status.Conditions = []datav1alpha1.DatasetCondition{}
		dataset.Status.CacheStates = common.CacheStateList{}
		// dataset.Status.RuntimeName = ""
		// dataset.Status.RuntimeType = ""
		// dataset.Status.RuntimeNamespace = ""
		// if len(dataset.Status.Runtimes) == 0 {
		dataset.Status.Runtimes = []datav1alpha1.Runtime{}
		dataset.Status.HCFSStatus = nil
		dataset.Status.FileNum = ""
		// }

		if err := r.Status().Update(ctx, dataset); err != nil {
			log.Error(err, "Failed to unbind the dataset", "dataset", dataset.Name)
			return utils.RequeueIfError(err)
		}
	}

	// 3. Remove finalizer
	r.Log.Info("before clean up finalizer", "runtime", ctx.Runtime)
	objectMeta, err := r.implement.GetRuntimeObjectMeta(ctx)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	if !objectMeta.GetDeletionTimestamp().IsZero() {
		finalizers := utils.RemoveString(objectMeta.GetFinalizers(), ctx.FinalizerName)
		objectMeta.SetFinalizers(finalizers)
		r.Log.Info("After clean up finalizer", "runtime", ctx.Runtime)
		if err := r.Update(ctx, ctx.Runtime); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		ctx.Log.V(1).Info("Finalizer is removed", "runtime", ctx.Runtime)
	}

	return ctrl.Result{}, nil
}

// ReconcileRuntime reconciles runtime
func (r *RuntimeReconciler) ReconcileRuntime(engine base.Engine, ctx cruntime.ReconcileRequestContext) (result ctrl.Result, err error) {
	var (
		log = ctx.Log.WithName("reconcileRuntime")
	)
	log.V(1).Info("process the Runtime", "Runtime", ctx.NamespacedName)

	// 1.Setup the ddc engine, and wait it ready
	if !utils.IsSetupDone(ctx.Dataset) {
		ready, err := engine.Setup(ctx)
		if err != nil {
			r.Recorder.Eventf(ctx.Runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to setup ddc engine due to error %v", err)
			log.Error(err, "Failed to setup the ddc engine")
			reportErr := r.ReportDatasetNotReadyCondition(ctx, err)
			if reportErr != nil {
				// ignore report error
				log.Error(err, "failed to report not ready condition to dataset", "dataset", types.NamespacedName{Namespace: ctx.Namespace, Name: ctx.Name})
			}
			// return utils.RequeueIfError(errors.Wrap(err, "Failed to steup the ddc engine"))
		}
		if !ready {
			return utils.RequeueAfterInterval(time.Duration(20 * time.Second))
		}
	} else {
		log.Info("The runtime is already setup.")
	}

	// 2.Setup the volume
	err = engine.CreateVolume()
	if err != nil && utils.IgnoreAlreadyExists(err) != nil {
		r.Recorder.Eventf(ctx.Runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to setup volume due to error %v", err)
		log.Error(err, "Failed to steup the volume")
		// return utils.RequeueIfError(errors.Wrap(err, "Failed to steup the ddc engine"))
	}

	// 3.sync up
	err = engine.Sync(ctx)
	if err != nil {
		r.Recorder.Eventf(ctx.Runtime, corev1.EventTypeWarning, common.ErrorProcessRuntimeReason, "Failed to sync the ddc due to %v", err)
		return utils.RequeueAfterInterval(time.Duration(20 * time.Second))
	}

	return utils.RequeueAfterInterval(time.Duration(90 * time.Second))
}

// AddFinalizerAndRequeue add  finalizer and requeue
func (r *RuntimeReconciler) AddFinalizerAndRequeue(ctx cruntime.ReconcileRequestContext, finalizerName string) (ctrl.Result, error) {
	log := ctx.Log.WithName("AddFinalizerAndRequeue")
	log.Info("add finalizer and requeue", "Runtime", ctx.NamespacedName)
	// objectMetaAccessor, isOM := ctx.Runtime.(metav1.ObjectMetaAccessor)
	// if !isOM {
	// 	return utils.RequeueIfError(fmt.Errorf("object is not ObjectMetaAccessor"))
	// }
	// objectMeta := objectMetaAccessor.GetObjectMeta()
	objectMeta, err := r.implement.GetRuntimeObjectMeta(ctx)
	if err != nil {
		return utils.RequeueIfError(err)
	}
	prevGeneration := objectMeta.GetGeneration()
	objectMeta.SetFinalizers(append(objectMeta.GetFinalizers(), finalizerName))
	if err := r.Update(ctx, ctx.Runtime); err != nil {
		ctx.Log.Error(err, "Failed to add finalizer", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}
	// controllerutil.AddFinalizer(ctx.Runtime, finalizer)
	currentGeneration := objectMeta.GetGeneration()
	ctx.Log.Info("RequeueImmediatelyUnlessGenerationChanged", "prevGeneration", prevGeneration,
		"currentGeneration", currentGeneration)

	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, currentGeneration)
}

// AddOwnerAndRequeue add Owner and requeue
func (r *RuntimeReconciler) AddOwnerAndRequeue(ctx cruntime.ReconcileRequestContext, dataset *datav1alpha1.Dataset) (ctrl.Result, error) {
	objectMeta, err := r.implement.GetRuntimeObjectMeta(ctx)
	if err != nil {
		return utils.RequeueIfError(err)
	}
	objectMeta.SetOwnerReferences(append(objectMeta.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}))
	if err := r.Update(ctx, ctx.Runtime); err != nil {
		ctx.Log.Error(err, "Failed to add ownerreference", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediately()
}

// GetRuntimeObjectMeta gets runtime object meta
func (r *RuntimeReconciler) GetRuntimeObjectMeta(ctx cruntime.ReconcileRequestContext) (objectMeta metav1.Object, err error) {
	objectMetaAccessor, isOM := ctx.Runtime.(metav1.ObjectMetaAccessor)
	if !isOM {
		// return utils.RequeueIfError(fmt.Errorf("object is not ObjectMetaAccessor"))
		err = fmt.Errorf("object is not ObjectMetaAccessor")
		return
	}
	objectMeta = objectMetaAccessor.GetObjectMeta()
	return
}

// GetDataset gets the dataset
func (r *RuntimeReconciler) GetDataset(ctx cruntime.ReconcileRequestContext) (*datav1alpha1.Dataset, error) {
	var dataset datav1alpha1.Dataset
	if err := r.Get(ctx, ctx.NamespacedName, &dataset); err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *RuntimeReconciler) CheckIfReferenceDatasetIsSupported(ctx cruntime.ReconcileRequestContext) (bool, string) {
	mounted := base.GetPhysicalDatasetFromMounts(ctx.Dataset.Spec.Mounts)

	if len(mounted) > 0 && ctx.RuntimeType != common.ThinRuntime {
		return false, "dataset mounting another dataset can only use thin runtime"
	}
	return true, ""
}

func (r *RuntimeReconciler) ReportDatasetNotReadyCondition(ctx cruntime.ReconcileRequestContext, notReadyReason error) error {
	datasetToReport := ctx.Dataset
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(r.Client, datasetToReport.Name, datasetToReport.Namespace)
		if err != nil {
			return err
		}

		datasetToUpdate := dataset.DeepCopy()
		cond := utils.NewDatasetCondition(datav1alpha1.DatasetNotReady, datav1alpha1.DatasetFailedToSetupReason, notReadyReason.Error(), corev1.ConditionTrue)

		datasetToUpdate.Status.Conditions = utils.UpdateDatasetCondition(datasetToUpdate.Status.Conditions, cond)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = r.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// ForgetMetrics deletes related metrics in prometheus metrics to avoid memory inflation
func (r *RuntimeReconciler) ForgetMetrics(ctx cruntime.ReconcileRequestContext) {
	metrics.GetOrCreateRuntimeMetrics(ctx.Runtime.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name).Forget()
	metrics.GetOrCreateDatasetMetrics(ctx.Namespace, ctx.Name).Forget()
}

// The interface of RuntimeReconciler
type RuntimeReconcilerInterface interface {
	// ReconcileRuntimeDeletion reconcile runtime deletion
	ReconcileRuntimeDeletion(engine base.Engine, ctx cruntime.ReconcileRequestContext) (ctrl.Result, error)

	// ReconcileRuntime reconciles runtime
	ReconcileRuntime(engine base.Engine, ctx cruntime.ReconcileRequestContext) (ctrl.Result, error)

	// AddFinalizerAndRequeue add  finalizer and requeue
	AddFinalizerAndRequeue(ctx cruntime.ReconcileRequestContext, finalizerName string) (ctrl.Result, error)

	// GetDataset gets the dataset
	GetDataset(ctx cruntime.ReconcileRequestContext) (*datav1alpha1.Dataset, error)

	// GetOrCreateEngine gets or creates engine
	GetOrCreateEngine(
		ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error)

	// RemoveEngine removes engine
	RemoveEngine(ctx cruntime.ReconcileRequestContext)

	// GetRuntimeObjectMeta get runtime objectmeta
	GetRuntimeObjectMeta(ctx cruntime.ReconcileRequestContext) (metav1.Object, error)

	// ReconcileInternal
	ReconcileInternal(ctx cruntime.ReconcileRequestContext) (ctrl.Result, error)
}
