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
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// ReconcileRequestContext wraps up necessary info for reconciliation
type ReconcileRequestContext struct {
	context.Context
	types.NamespacedName
	Log      logr.Logger
	DataLoad datav1alpha1.DataLoad
}

// DataLoadReconciler reconciles a DataLoad object
type DataLoadReconciler struct {
	Scheme *runtime.Scheme
	*DataLoadReconcilerImplement
}

// NewDataLoadReconciler returns a DataLoadReconciler
func NewDataLoadReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataLoadReconciler {
	r := &DataLoadReconciler{
		Scheme: scheme,
	}
	r.DataLoadReconcilerImplement = NewDataLoadReconcilerImplement(client, log, recorder)
	return r
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=dataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=dataloads/status,verbs=get;update;patch
// Reconcile reconciles the DataLoad object
func (r *DataLoadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := ReconcileRequestContext{
		Context:        context.Background(),
		NamespacedName: req.NamespacedName,
		Log:            r.Log.WithValues("dataload", req.NamespacedName),
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

	ctx.DataLoad = *dataload
	ctx.Log.V(1).Info("DataLoad found", "detail", dataload)

	// 2. Reconcile deletion of the object if necessary
	if utils.HasDeletionTimestamp(ctx.DataLoad.ObjectMeta) {
		return r.ReconcileDataLoadDeletion(ctx)
	}

	// 3. get the dataset
	dataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("The dataset is not found", "dataset", ctx.NamespacedName)
			// no datset means no metadata, not necessary to ReconcileDataLoad
			return utils.RequeueAfterInterval(20 * time.Second)
		} else {
			ctx.Log.Error(err, "Failed to get the ddc dataset")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get dataset"))
		}
	}

	// 4. add finalizer and requeue
	if !utils.ContainsString(ctx.DataLoad.ObjectMeta.GetFinalizers(), cdataload.DATALOAD_FINALIZER) {
		return r.addFinalierAndRequeue(ctx)
	}

	// 5. add owner and requeue
	if !utils.ContainsOwners(ctx.DataLoad.GetOwnerReferences(), dataset) {
		return r.AddOwnerAndRequeue(ctx, dataset)
	}

	return r.ReconcileDataLoad(ctx)
}

// AddOwnerAndRequeue adds Owner and requeue
func (r *DataLoadReconciler) AddOwnerAndRequeue(ctx ReconcileRequestContext, dataset *datav1alpha1.Dataset) (ctrl.Result, error) {
	ctx.DataLoad.ObjectMeta.OwnerReferences = append(ctx.DataLoad.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	})
	if err := r.Update(ctx, &ctx.DataLoad); err != nil {
		ctx.Log.Error(err, "Failed to add ownerreference", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediately()
}

func (r *DataLoadReconciler) addFinalierAndRequeue(ctx ReconcileRequestContext) (ctrl.Result, error) {
	ctx.DataLoad.ObjectMeta.Finalizers = append(ctx.DataLoad.ObjectMeta.Finalizers, cdataload.DATALOAD_FINALIZER)
	ctx.Log.Info("Add finalizer and requeue", "finalizer", cdataload.DATALOAD_FINALIZER)
	prevGeneration := ctx.DataLoad.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &ctx.DataLoad); err != nil {
		ctx.Log.Error(err, "failed to add finalizer to dataload", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, ctx.DataLoad.ObjectMeta.GetGeneration())
}

// SetupWithManager sets up the controller with the given controller manager
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.DataLoad{}).
		Complete(r)
}
