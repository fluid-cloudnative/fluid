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

package datamigrate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/jindo"
)

const controllerName string = "DataMigrateReconciler"

// DataMigrateReconciler reconciles a DataMigrate object
type DataMigrateReconciler struct {
	Scheme  *runtime.Scheme
	engines map[string]base.Engine
	mutex   *sync.Mutex
	*DataMigrateReconcilerImplement
}

// NewDataMigrateReconciler returns a DataMigrateReconciler
func NewDataMigrateReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataMigrateReconciler {
	r := &DataMigrateReconciler{
		Scheme:  scheme,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.DataMigrateReconcilerImplement = NewDataMigrateReconcilerImplement(client, log, recorder)
	return r
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=datamigrates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=datamigrates/status,verbs=get;update;patch
// Reconcile reconciles the DataMigrate object
func (r *DataMigrateReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx := cruntime.ReconcileRequestContext{
		Context:  context,
		Log:      r.Log.WithValues("datamigrate", req.NamespacedName),
		Recorder: r.Recorder,
		Client:   r.Client,
		Category: common.AccelerateCategory,
	}

	// 1. Get DataMigrate object
	dataMigrate, err := utils.GetDataMigrate(r.Client, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("DataMigrate not found")
			return utils.NoRequeue()
		} else {
			ctx.Log.Error(err, "failed to get DataMigrate")
			return utils.RequeueIfError(errors.Wrap(err, "failed to get DataMigrate info"))
		}
	}

	targetDataMigrate := *dataMigrate
	ctx.Log.V(1).Info("dataMigrate found", "detail", dataMigrate)

	// 2. Reconcile deletion of the object if necessary
	if utils.HasDeletionTimestamp(dataMigrate.ObjectMeta) {
		return r.ReconcileDataMigrateDeletion(ctx, targetDataMigrate, r.engines, r.mutex)
	}

	// 3. get target dataset
	targetDataset, err := utils.GetTargetDatasetOfMigrate(r.Client, targetDataMigrate)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("can't find target dataset", "dataMigrate", targetDataMigrate.Name)
			return utils.RequeueAfterInterval(20 * time.Second)
		}
		// other error
		ctx.Log.Error(err, "Failed to get the ddc dataset")
		return utils.RequeueIfError(errors.Wrap(err, "Unable to get dataset"))
	}
	ctx.Dataset = targetDataset
	ctx.NamespacedName = types.NamespacedName{
		Name:      targetDataset.Name,
		Namespace: targetDataset.Namespace,
	}

	// 4. get the runtime
	index, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)
	if index == -1 {
		ctx.Log.Info("bounded runtime with Accelerate Category is not found on the target dataset", "targetDataset", targetDataset)
		r.Recorder.Eventf(&targetDataMigrate,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}
	ctx.RuntimeType = boundedRuntime.Type

	var fluidRuntime client.Object
	switch ctx.RuntimeType {
	case common.AlluxioRuntime:
		fluidRuntime, err = utils.GetAlluxioRuntime(ctx.Client, boundedRuntime.Name, boundedRuntime.Namespace)
	case common.JindoRuntime:
		fluidRuntime, err = utils.GetJindoRuntime(ctx.Client, boundedRuntime.Name, boundedRuntime.Namespace)
		ctx.RuntimeType = jindo.GetRuntimeType()
	case common.GooseFSRuntime:
		fluidRuntime, err = utils.GetGooseFSRuntime(ctx.Client, boundedRuntime.Name, boundedRuntime.Namespace)
	case common.JuiceFSRuntime:
		fluidRuntime, err = utils.GetJuiceFSRuntime(ctx.Client, boundedRuntime.Name, boundedRuntime.Namespace)
	default:
		ctx.Log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "runtime", boundedRuntime)
		r.Recorder.Eventf(&targetDataMigrate,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not supported")
	}

	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("The runtime is not found", "runtime", ctx.NamespacedName)
			return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get the ddc runtime")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get ddc runtime"))
		}
	}
	ctx.Runtime = fluidRuntime
	ctx.Log.V(1).Info("get the runtime", "runtime", ctx.Runtime)

	// 5. create or get engine
	engine, err := r.GetOrCreateEngine(ctx)
	if err != nil {
		r.Recorder.Eventf(&targetDataMigrate, v1.EventTypeWarning, common.ErrorProcessDatasetReason, "Process dataMigrate error %v", err)
		return utils.RequeueIfError(errors.Wrap(err, "Failed to create or get engine"))
	}

	// 6. add finalizer and requeue
	if !utils.ContainsString(targetDataMigrate.ObjectMeta.GetFinalizers(), cdatamigrate.DataMigrateFinalizer) {
		return r.addFinalizerAndRequeue(ctx, targetDataMigrate)
	}

	// 7. add owner and requeue
	if !utils.ContainsOwners(targetDataMigrate.GetOwnerReferences(), targetDataset) {
		return r.AddOwnerAndRequeue(ctx, targetDataMigrate, targetDataset)
	}

	return r.ReconcileDataMigrate(ctx, targetDataMigrate, engine)
}

// AddOwnerAndRequeue adds Owner and requeue
func (r *DataMigrateReconciler) AddOwnerAndRequeue(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate, targetDataset *datav1alpha1.Dataset) (ctrl.Result, error) {
	targetDataMigrate.ObjectMeta.OwnerReferences = append(targetDataMigrate.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: targetDataset.APIVersion,
		Kind:       targetDataset.Kind,
		Name:       targetDataset.Name,
		UID:        targetDataset.UID,
	})
	if err := r.Update(ctx, &targetDataMigrate); err != nil {
		ctx.Log.Error(err, "Failed to add ownerreference", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediately()
}

func (r *DataMigrateReconciler) addFinalizerAndRequeue(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (ctrl.Result, error) {
	targetDataMigrate.ObjectMeta.Finalizers = append(targetDataMigrate.ObjectMeta.Finalizers, cdatamigrate.DataMigrateFinalizer)
	ctx.Log.Info("Add finalizer and requeue", "finalizer", cdatamigrate.DataMigrateFinalizer)
	prevGeneration := targetDataMigrate.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &targetDataMigrate); err != nil {
		ctx.Log.Error(err, "failed to add finalizer to dataMigrate", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, targetDataMigrate.ObjectMeta.GetGeneration())
}

// SetupWithManager sets up the controller with the given controller manager
func (r *DataMigrateReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.DataMigrate{}).
		Complete(r)
}

// GetOrCreateEngine gets the Engine
func (r *DataMigrateReconciler) GetOrCreateEngine(
	ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error) {
	found := false
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if engine, found = r.engines[id]; !found {
		engine, err = ddc.CreateEngine(id,
			ctx)
		if err != nil {
			return nil, err
		}
		r.engines[id] = engine
		r.Log.V(1).Info("Put Engine to engine map")
	} else {
		r.Log.V(1).Info("Get Engine from engine map")
	}

	return engine, err
}

func (r *DataMigrateReconciler) ControllerName() string {
	return controllerName
}
