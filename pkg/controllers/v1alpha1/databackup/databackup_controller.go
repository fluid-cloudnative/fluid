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

package databackup

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

const controllerName string = "DataBackupController"

// DataBackupReconciler reconciles a DataBackup object
type DataBackupReconciler struct {
	Scheme *runtime.Scheme
	*controllers.OperationReconciler
}

var _ dataoperation.OperationInterfaceBuilder = &DataBackupReconciler{}

// NewDataBackupReconciler returns a DataBackupReconciler
func NewDataBackupReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataBackupReconciler {
	r := &DataBackupReconciler{
		Scheme: scheme,
	}
	r.OperationReconciler = controllers.NewDataOperationReconciler(r, client, log, recorder)
	return r
}

func (r *DataBackupReconciler) Build(object client.Object) (dataoperation.OperationInterface, error) {
	dataBackup, ok := object.(*datav1alpha1.DataBackup)
	if !ok {
		return nil, fmt.Errorf("object %v is not a DataBackup", object)
	}

	return &dataBackupOperation{
		Client:     r.Client,
		Log:        r.Log,
		Recorder:   r.Recorder,
		dataBackup: dataBackup,
	}, nil
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=databackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=databackups/status,verbs=get;update;patch
// Reconcile reconciles the DataBackup object
func (r *DataBackupReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {

	ctx := dataoperation.ReconcileRequestContext{
		// used for create engine
		ReconcileRequestContext: cruntime.ReconcileRequestContext{
			Context:  context,
			Log:      r.Log.WithValues("DataBackup", req.NamespacedName),
			Recorder: r.Recorder,
			Client:   r.Client,
			Category: common.AccelerateCategory,
		},
		DataOpFinalizerName: cdatabackup.Finalizer,
	}
	// 1. Get DataBackup object
	dataBackup, err := utils.GetDataBackup(r.Client, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("DataBackup not found")
			return utils.NoRequeue()
		} else {
			ctx.Log.Error(err, "failed to get DataBackup")
			return utils.RequeueIfError(errors.Wrap(err, "failed to get DataBackup info"))
		}
	}
	ctx.DataObject = dataBackup
	ctx.OpStatus = &dataBackup.Status

	return r.ReconcileInternal(ctx)
}

// SetupWithManager sets up the controller with the given controller manager
func (r *DataBackupReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.DataBackup{}).
		Complete(r)
}

func (r *DataBackupReconciler) ControllerName() string {
	return controllerName
}
