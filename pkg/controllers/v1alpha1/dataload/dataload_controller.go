/*
Copyright 2022 The Fluid Authors.

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

var _ dataoperation.OperationReconcilerInterfaceBuilder = &DataLoadReconciler{}

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

func (r *DataLoadReconciler) Build(object client.Object) (dataoperation.OperationReconcilerInterface, error) {
	dataLoad, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		return nil, fmt.Errorf("object %v is not a DataLoad", object)
	}

	return &dataLoadReconciler{
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
