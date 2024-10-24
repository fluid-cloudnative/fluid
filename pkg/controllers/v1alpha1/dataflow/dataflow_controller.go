/*
Copyright 2023 The Fluid Author.

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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/discovery"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const controllerName string = "DataFlowReconciler"

type DataFlowReconciler struct {
	client.Client
	Recorder     record.EventRecorder
	Log          logr.Logger
	ResyncPeriod time.Duration
}

func NewDataFlowReconciler(client client.Client,
	log logr.Logger,
	recorder record.EventRecorder,
	resyncPeriod time.Duration) *DataFlowReconciler {
	return &DataFlowReconciler{
		Client:       client,
		Recorder:     recorder,
		Log:          log,
		ResyncPeriod: resyncPeriod,
	}
}

var reconcileKinds = map[string]client.Object{
	"databackup":  &datav1alpha1.DataBackup{},
	"dataload":    &datav1alpha1.DataLoad{},
	"datamigrate": &datav1alpha1.DataMigrate{},
	"dataprocess": &datav1alpha1.DataProcess{},
}

func setupWatches(bld *builder.Builder, handler *handler.EnqueueRequestForObject, predicates builder.Predicates) *builder.Builder {
	toSetup := []client.Object{}
	for kind, obj := range reconcileKinds {
		if discovery.GetFluidDiscovery().ResourceEnabled(kind) {
			toSetup = append(toSetup, obj)
		}
	}

	for i, obj := range toSetup {
		if i == 0 {
			bld.For(obj, predicates)
		} else {
			bld.Watches(obj, handler, predicates)
		}
	}

	return bld
}

func DataFlowEnabled() bool {
	for kind := range reconcileKinds {
		if discovery.GetFluidDiscovery().ResourceEnabled(kind) {
			return true
		}
	}

	return false
}

type reconcileRequestContext struct {
	context.Context
	types.NamespacedName
	Client   client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

var reconcileFuncs = []func(reconcileRequestContext) (bool, error){
	reconcileDataLoad,
	reconcileDataMigrate,
	reconcileDataProcess,
	reconcileDataBackup,
}

func (r *DataFlowReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("dataflow request", req.NamespacedName)
	log.Info("reconcile starts")
	defer log.Info("reconcile ends")

	ctx := reconcileRequestContext{
		Context:        context,
		NamespacedName: req.NamespacedName,
		Client:         r.Client,
		Log:            r.Log,
		Recorder:       r.Recorder,
	}

	needRequeue := false
	for _, reconcileFn := range reconcileFuncs {
		opNeedRequeue, err := reconcileFn(ctx)
		if err != nil {
			return utils.RequeueIfError(err)
		}

		// requeue if any of the operation needs requeue
		needRequeue = needRequeue || opNeedRequeue
	}

	if needRequeue {
		return utils.RequeueAfterInterval(r.ResyncPeriod)
	}
	return utils.NoRequeue()
}

func (r *DataFlowReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	handler := &handler.EnqueueRequestForObject{}

	predicates := builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
		if !obj.GetDeletionTimestamp().IsZero() {
			// No need to trigger reconcilations for deleted objects
			return false
		}

		objNamespacedName := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
		opStatus, err := utils.GetOperationStatus(obj)
		if err != nil {
			r.Log.Info(fmt.Sprintf("skip enqueue object: %v", err.Error()), "object", objNamespacedName)
			return false
		}

		// DataFlowReconciler only reconciles data operations that are waiting for other data operations
		if opStatus.WaitingFor.OperationComplete != nil && *opStatus.WaitingFor.OperationComplete {
			return true
		}

		r.Log.V(1).Info("skip enqueue object: operatin not waiting", "object", objNamespacedName)
		return false
	}))

	bld := ctrl.NewControllerManagedBy(mgr).WithOptions(options)
	return setupWatches(bld, handler, predicates).Complete(r)
}

func (r *DataFlowReconciler) ControllerName() string {
	return controllerName
}
