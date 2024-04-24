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
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	jindoutils "github.com/fluid-cloudnative/fluid/pkg/utils/jindo"
)

// OperationReconciler is the default implementation
type OperationReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder

	// engines cache for runtime engine
	engines map[string]base.Engine
	// mutex lock for engines
	mutex *sync.Mutex

	// Real implementBuilder
	implementBuilder dataoperation.OperationInterfaceBuilder
}

// NewDataOperationReconciler creates the default OperationReconciler
func NewDataOperationReconciler(operationReconcilerInterface dataoperation.OperationInterfaceBuilder, client client.Client,
	log logr.Logger, recorder record.EventRecorder) *OperationReconciler {

	r := &OperationReconciler{
		Client:           client,
		Recorder:         recorder,
		Log:              log,
		mutex:            &sync.Mutex{},
		engines:          map[string]base.Engine{},
		implementBuilder: operationReconcilerInterface,
	}
	return r
}

// ReconcileDeletion reconciles the deletion of the DataBackup
func (o *OperationReconciler) ReconcileDeletion(ctx dataoperation.ReconcileRequestContext,
	implement dataoperation.OperationInterface) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDeletion")

	// 1. Delete helm release if exists
	namespacedName := implement.GetReleaseNameSpacedName()
	err := helm.DeleteReleaseIfExists(namespacedName.Name, namespacedName.Namespace)
	if err != nil {
		log.Error(err, "can't delete release", "releaseName", namespacedName.Name)
		return utils.RequeueIfError(err)
	}

	// 2. Release lock on target dataset if necessary
	err = base.ReleaseTargetDataset(ctx.ReconcileRequestContext, implement)
	// ignore the not found error, as dataset can be deleted first, then the data operation will be deleted by owner reference.
	if utils.IgnoreNotFound(err) != nil {
		log.Error(err, "can't release lock on target dataset")
		return utils.RequeueIfError(err)
	}

	// 3. delete engine
	o.RemoveEngine(ctx)

	object := implement.GetOperationObject()
	// 4. remove finalizer
	if !object.GetDeletionTimestamp().IsZero() {
		objectMeta, err := utils.GetObjectMeta(object)
		if err != nil {
			return utils.RequeueIfError(err)
		}

		finalizers := utils.RemoveString(objectMeta.GetFinalizers(), ctx.DataOpFinalizerName)
		objectMeta.SetFinalizers(finalizers)

		if err := o.Update(ctx, object); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		log.Info("Finalizer is removed")
	}
	return utils.NoRequeue()
}

func (o *OperationReconciler) ReconcileInternal(ctx dataoperation.ReconcileRequestContext) (ctrl.Result, error) {
	var object = ctx.DataObject
	implement, err := o.implementBuilder.Build(object)

	if err != nil {
		ctx.Log.Error(err, "build operation reconcile object failed")
		return utils.RequeueIfError(err)
	}

	// 1. Reconcile deletion of the object if necessary
	if !object.GetDeletionTimestamp().IsZero() {
		return o.ReconcileDeletion(ctx, implement)
	}

	// 2. set target dataset
	targetDataset, err := implement.GetTargetDataset()
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			statusError := err.(*apierrors.StatusError)
			ctx.Log.Info("The dataset is not found", "dataset", statusError.Status().Details.Name)
			o.Recorder.Eventf(object, v1.EventTypeWarning, common.TargetDatasetNotFound, "Target dataset not found: %v", err)
			return utils.RequeueAfterInterval(20 * time.Second)
		} else {
			ctx.Log.Error(err, "Failed to get the ddc dataset")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get dataset"))
		}
	}

	// set the namespace and name for runtime
	ctx.NamespacedName = types.NamespacedName{
		Namespace: targetDataset.Namespace,
		Name:      targetDataset.Name,
	}
	ctx.Dataset = targetDataset

	// 3. set target runtime and runtimeType
	index, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)
	if index == -1 {
		ctx.Log.Info("bounded runtime with Accelerate Category is not found on the target dataset", "targetDataset", targetDataset)
		o.Recorder.Eventf(object,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// GetOrCreateEngine() will use the ctx.EngineImpl field to choose which engine to build
	ctx.RuntimeType = boundedRuntime.Type
	ctx.Runtime, ctx.EngineImpl, err = o.getRuntimeObjectAndEngineImpl(boundedRuntime.Type, boundedRuntime.Name, boundedRuntime.Namespace)

	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("The runtime is not found", "runtime", ctx.NamespacedName)
			return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get the ddc runtime")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get ddc runtime"))
		}
	}

	ctx.Log.V(1).Info("get the runtime", "runtime", ctx.Runtime)

	// 5. create or get engine
	engine, err := o.GetOrCreateEngine(ctx)
	if err != nil {
		o.Recorder.Eventf(object, v1.EventTypeWarning, common.ErrorProcessDatasetReason, "Process %s error %v", implement.GetOperationType(), err)
		return utils.RequeueIfError(errors.Wrap(err, "Failed to create or get engine"))
	}

	// 6. add finalizer and requeue
	if !utils.ContainsString(object.GetFinalizers(), ctx.DataOpFinalizerName) {
		return o.addFinalizerAndRequeue(ctx, object)
	}

	// 7. add owner and requeue
	if !utils.ContainsOwners(object.GetOwnerReferences(), targetDataset) {
		return o.addOwnerAndRequeue(ctx, object, targetDataset)
	}

	// 8. do the data operation
	return engine.Operate(ctx.ReconcileRequestContext, ctx.OpStatus, implement)
}

// GetOrCreateEngine gets the Engine
// require each runtime must use ddc.CreateEngine to create engine
func (o *OperationReconciler) GetOrCreateEngine(
	ctx dataoperation.ReconcileRequestContext) (engine base.Engine, err error) {
	found := false
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if engine, found = o.engines[id]; !found {
		engine, err = ddc.CreateEngine(id, ctx.ReconcileRequestContext)
		if err != nil {
			return nil, err
		}
		o.engines[id] = engine
		o.Log.V(1).Info("Put Engine to engine map")
	} else {
		o.Log.V(1).Info("Get Engine from engine map")
	}

	return engine, err
}

func (o *OperationReconciler) RemoveEngine(ctx dataoperation.ReconcileRequestContext) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	delete(o.engines, id)
}

func (o *OperationReconciler) addFinalizerAndRequeue(ctx dataoperation.ReconcileRequestContext, object client.Object) (ctrl.Result, error) {
	objectMeta, err := utils.GetObjectMeta(object)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	finalizers := append(objectMeta.GetFinalizers(), ctx.DataOpFinalizerName)
	objectMeta.SetFinalizers(finalizers)
	ctx.Log.Info("Add finalizer and requeue", "finalizer", ctx.Finalizers)

	prevGeneration := object.GetGeneration()
	if err := o.Update(ctx, object); err != nil {
		ctx.Log.Error(err, "failed to add finalizer", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, object.GetGeneration())

}

func (o *OperationReconciler) addOwnerAndRequeue(ctx dataoperation.ReconcileRequestContext, object client.Object, dataset *datav1alpha1.Dataset) (ctrl.Result, error) {
	objectMeta, err := utils.GetObjectMeta(object)
	if err != nil {
		return utils.RequeueIfError(err)
	}

	ownerReferences := append(object.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	})
	objectMeta.SetOwnerReferences(ownerReferences)

	if err := o.Update(ctx, object); err != nil {
		ctx.Log.Error(err, "Failed to add owner reference", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediately()
}

// getRuntimeObjectAndEngineImpl firstly gets a runtime object given its name, namespace and runtimeType, and then infer its engine implementation
// from the retrieved runtime object.
// TODO: Maybe the function can be separated into two sub-function and make them globally accessible.
func (o *OperationReconciler) getRuntimeObjectAndEngineImpl(runtimeType, name, namespace string) (obj client.Object, engineImpl string, err error) {
	// support all runtime
	var runtime base.RuntimeInterface
	switch runtimeType {
	case common.AlluxioRuntime:
		runtime, err = utils.GetAlluxioRuntime(o.Client, name, namespace)
	case common.JindoRuntime:
		runtime, err = utils.GetJindoRuntime(o.Client, name, namespace)
	case common.GooseFSRuntime:
		runtime, err = utils.GetGooseFSRuntime(o.Client, name, namespace)
	case common.JuiceFSRuntime:
		runtime, err = utils.GetJuiceFSRuntime(o.Client, name, namespace)
	case common.EFCRuntime:
		runtime, err = utils.GetEFCRuntime(o.Client, name, namespace)
	case common.ThinRuntime:
		runtime, err = utils.GetThinRuntime(o.Client, name, namespace)
	case common.VineyardRuntime:
		runtime, err = utils.GetVineyardRuntime(o.Client, name, namespace)
	}

	if err != nil {
		return
	}

	if runtimeType == common.ThinRuntime {
		// ThinRuntime cannot use ddc.InferEngineImpl because ReferenceDatasetEngine inherits valueFile property
		// from its physical runtime.
		// TODO: We should determine whether ReferenceDatasetEngine is an engineImpl of ThinRuntime.
		return runtime, common.ThinEngineImpl, nil
	}

	switch runtimeType {
	case common.AlluxioRuntime:
		return runtime, ddc.InferEngineImpl(*runtime.GetStatus(), common.AlluxioEngineImpl), nil
	case common.JindoRuntime:
		return runtime, ddc.InferEngineImpl(*runtime.GetStatus(), jindoutils.GetDefaultEngineImpl()), nil
	case common.GooseFSRuntime:
		return runtime, ddc.InferEngineImpl(*runtime.GetStatus(), common.GooseFSEngineImpl), nil
	case common.JuiceFSRuntime:
		return runtime, ddc.InferEngineImpl(*runtime.GetStatus(), common.JuiceFSEngineImpl), nil
	case common.EFCRuntime:
		return runtime, ddc.InferEngineImpl(*runtime.GetStatus(), common.EFCEngineImpl), nil
	case common.VineyardRuntime:
		return runtime, ddc.InferEngineImpl(*runtime.GetStatus(), common.VineyardEngineImpl), nil
	}

	err = fmt.Errorf("runtimeType %s is not supported", runtimeType)
	return
}
