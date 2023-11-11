/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package dataflow

import (
	"context"
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		// controller is forced to For a kind by controller-runtime
		For(&datav1alpha1.DataBackup{}, predicates).
		Watches(&source.Kind{Type: &datav1alpha1.DataLoad{}}, handler, predicates).
		// Watches(&source.Kind{Type: &datav1alpha1.DataBackup{}}, handler, predicates).
		Watches(&source.Kind{Type: &datav1alpha1.DataMigrate{}}, handler, predicates).
		Watches(&source.Kind{Type: &datav1alpha1.DataProcess{}}, handler, predicates).
		Complete(r)
}

func (r *DataFlowReconciler) ControllerName() string {
	return controllerName
}
