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

package dataload

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

const controllerName string = "DataLoadReconciler"

// DataLoadReconciler reconciles a DataLoad object
type DataLoadReconciler struct {
	Scheme *runtime.Scheme
	*controllers.OperationReconciler
}

var _ dataoperation.OperationInterfaceBuilder = &DataLoadReconciler{}

// NewDataLoadReconciler returns a DataLoadReconciler
func NewDataLoadReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataLoadReconciler {
	r := &DataLoadReconciler{
		Scheme: scheme,
	}
	r.OperationReconciler = controllers.NewDataOperationReconciler(r, client, log, recorder)
	return r
}

func (r *DataLoadReconciler) Build(object client.Object) (dataoperation.OperationInterface, error) {
	dataLoad, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		return nil, fmt.Errorf("object %v is not a DataLoad", object)
	}

	return &dataLoadOperation{
		Client:   r.Client,
		Log:      r.Log,
		Recorder: r.Recorder,
		dataLoad: dataLoad,
	}, nil
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=dataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=dataloads/status,verbs=get;update;patch
// Reconcile reconciles the DataLoad object
func (r *DataLoadReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx := dataoperation.ReconcileRequestContext{
		// used for create engine
		ReconcileRequestContext: cruntime.ReconcileRequestContext{
			Context:  context,
			Log:      r.Log.WithValues("DataLoad", req.NamespacedName),
			Recorder: r.Recorder,
			Client:   r.Client,
			Category: common.AccelerateCategory,
		},
		DataOpFinalizerName: cdataload.DataloadFinalizer,
	}

	// 1. Get DataLoad object
	dataload, err := utils.GetDataLoad(r.Client, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("DataLoad not found")
			return utils.NoRequeue()
		} else {
			ctx.Log.Error(err, "failed to get DataLoad")
			return utils.RequeueIfError(errors.Wrap(err, "failed to get DataLoad info"))
		}
	}
	ctx.DataObject = dataload
	ctx.OpStatus = &dataload.Status

	return r.ReconcileInternal(ctx)
}

// SetupWithManager sets up the controller with the given controller manager
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	if compatibility.IsBatchV1CronJobSupported() {
		return ctrl.NewControllerManagedBy(mgr).
			WithOptions(options).
			For(&datav1alpha1.DataLoad{}).
			Owns(&batchv1.CronJob{}).
			Complete(r)
	} else {
		ctrl.Log.Info("batch/v1 cronjobs cannnot be found in cluster, fallback to watch batch/v1beta1 cronjobs for compatibility")
		return ctrl.NewControllerManagedBy(mgr).
			WithOptions(options).
			For(&datav1alpha1.DataLoad{}).
			Owns(&batchv1beta1.CronJob{}).
			Complete(r)
	}
}

func (r *DataLoadReconciler) ControllerName() string {
	return controllerName
}
