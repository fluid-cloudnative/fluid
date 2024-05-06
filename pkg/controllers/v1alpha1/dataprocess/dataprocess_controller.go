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

package dataprocess

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cdataprocess "github.com/fluid-cloudnative/fluid/pkg/dataprocess"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

const controllerName string = "DataProcessReconciler"

// DataProcessReconciler reconciles a DataProcess object
type DataProcessReconciler struct {
	Scheme *runtime.Scheme
	*controllers.OperationReconciler
}

var _ dataoperation.OperationInterfaceBuilder = &DataProcessReconciler{}

func (r *DataProcessReconciler) Build(object client.Object) (dataoperation.OperationInterface, error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		return nil, fmt.Errorf("object %v is not a DataProcess", object)
	}

	return &dataProcessOperation{
		Client:      r.Client,
		Reader:      r.Reader,
		Log:         r.Log,
		Recorder:    r.Recorder,
		dataProcess: dataProcess,
	}, nil
}

func NewDataProcessReconciler(client client.Client, reader client.Reader,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataProcessReconciler {
	r := &DataProcessReconciler{
		Scheme: scheme,
	}
	r.OperationReconciler = controllers.NewDataOperationReconciler(r, client, reader, log, recorder)
	return r
}

//+kubebuilder:rbac:groups=data.fluid.io,resources=dataprocesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.fluid.io,resources=dataprocesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=data.fluid.io,resources=dataprocesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DataProcess object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *DataProcessReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx := dataoperation.ReconcileRequestContext{
		ReconcileRequestContext: cruntime.ReconcileRequestContext{
			Context:  context,
			Log:      r.Log.WithValues("DataProcess", req.NamespacedName),
			Recorder: r.Recorder,
			Client:   r.Client,
			Category: common.AccelerateCategory,
		},
		DataOpFinalizerName: cdataprocess.DataProcessFinalizer,
	}

	dataprocess, err := utils.GetDataProcess(r.Client, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("DataProcess not found")
			return utils.NoRequeue()
		} else {
			ctx.Log.Error(err, "failed to get DataProcess")
			return utils.RequeueIfError(errors.Wrap(err, "failed to get DataProcess info"))
		}
	}
	ctx.DataObject = dataprocess
	ctx.OpStatus = &dataprocess.Status

	return r.ReconcileInternal(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataProcessReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.DataProcess{}).
		Complete(r)
}

func (r *DataProcessReconciler) ControllerName() string {
	return controllerName
}
