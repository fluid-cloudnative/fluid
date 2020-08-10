/*

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
	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
	cdataload "github.com/cloudnativefluid/fluid/pkg/dataload"
	"github.com/cloudnativefluid/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DataLoadReconciler reconciles a AlluxioDataLoad object
type DataLoadReconciler struct {
	Scheme *runtime.Scheme
	*ReconcilerImplement
}

//Reconcile reconciles the AlluxioDataLoad Object
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads/status,verbs=get;update;patch

func NewDataLoadReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataLoadReconciler {
	r := &DataLoadReconciler{
		Scheme: scheme,
	}
	r.ReconcilerImplement = NewReconcilerImplement(client, log)
	return r
}

func (r *DataLoadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := cdataload.ReconcileRequestContext{
		Context:        context.Background(),
		Log:            r.Log.WithValues("alluxiodataload", req.NamespacedName),
		NamespacedName: req.NamespacedName,
	}

	ctx.Log.V(1).Info("Reconciling dataload request", "request", req)

	/*
		1. Load the DataLoad resource object
	*/
	dataload, err := utils.GetDataLoad(r, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Dataload not found", "dataload", ctx.NamespacedName)
			return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get dataload info")
			return utils.RequeueIfError(errors.Wrap(err, "Failed to get dataload info"))
		}
	}
	ctx.DataLoad = *dataload
	ctx.Log.Info("Dataload found.", "dataload", ctx.DataLoad)

	/*
		2. delete dataload if necessary
	*/
	if utils.HasDeletionTimestamp(ctx.DataLoad.ObjectMeta) {
		return r.ReconcileDataloadDeletion(ctx)
	}

	/*
		3. Add finalizer
	*/
	if !utils.ContainsString(ctx.DataLoad.ObjectMeta.GetFinalizers(), common.Finalizer) {
		return r.addFinalizerAndRequeue(ctx)
	}

	/*
		4. Do dataload reconciling
	*/
	return r.ReconcileDataload(ctx)
}

func (r *DataLoadReconciler) addFinalizerAndRequeue(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	ctx.DataLoad.ObjectMeta.Finalizers = append(ctx.DataLoad.ObjectMeta.Finalizers, common.Finalizer)
	ctx.Log.Info("Add finalizer and Requeue", "finalizer", common.Finalizer)
	prevGeneration := ctx.DataLoad.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &ctx.DataLoad); err != nil {
		ctx.Log.Error(err, "Failed to add finalizer to dataload", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, ctx.DataLoad.ObjectMeta.GetGeneration())
}

//SetupWithManager setups the manager with AlluxioDataLoad
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.AlluxioDataLoad{}).
		Complete(r)
}
