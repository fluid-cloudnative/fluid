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

package databackup

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// reconcileRequestContext wraps up necessary info for reconciliation
type reconcileRequestContext struct {
	context.Context
	types.NamespacedName
	Log        logr.Logger
	DataBackup datav1alpha1.DataBackup
	Dataset    *datav1alpha1.Dataset
}

// DataBackupReconciler reconciles a DataBackup object
type DataBackupReconciler struct {
	Scheme *runtime.Scheme
	*DataBackupReconcilerImplement
}

// NewDataBackupReconciler returns a DataBackupReconciler
func NewDataBackupReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataBackupReconciler {
	r := &DataBackupReconciler{
		Scheme: scheme,
	}
	r.DataBackupReconcilerImplement = NewDataBackupReconcilerImplement(client, log, recorder)
	return r
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=databackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=databackups/status,verbs=get;update;patch
// Reconcile reconciles the DataBackup object
func (r *DataBackupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := reconcileRequestContext{
		Context:        context.Background(),
		NamespacedName: req.NamespacedName,
		Log:            r.Log.WithValues("databackup", req.NamespacedName),
	}
	// 1. Get DataBackup object
	databackup, err := utils.GetDataBackup(r.Client, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("DataBackup not found")
			return utils.NoRequeue()
		} else {
			ctx.Log.Error(err, "failed to get DataBackup")
			return utils.RequeueIfError(errors.Wrap(err, "failed to get DataBackup info"))
		}
	}
	ctx.DataBackup = *databackup
	ctx.Log.V(1).Info("DataBackup found", "detail", databackup)

	// 2. Reconcile deletion of the object if necessary
	if utils.HasDeletionTimestamp(ctx.DataBackup.ObjectMeta) {
		return r.ReconcileDataBackupDeletion(ctx)
	}

	// 3. get the dataset
	dataset, err := utils.GetDataset(r.Client, databackup.Spec.Dataset, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("The dataset is not found", "dataset", ctx.NamespacedName)
			// no datset means no metadata, not necessary to ReconcileDatabackup
			return utils.RequeueAfterInterval(20 * time.Second)
		} else {
			ctx.Log.Error(err, "Failed to get the ddc dataset")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get dataset"))
		}
	}
	// 4. add finalizer and requeue
	if !utils.ContainsString(ctx.DataBackup.ObjectMeta.GetFinalizers(), cdatabackup.FINALIZER) {
		return r.addFinalierAndRequeue(ctx)
	}

	// 5. add owner and requeue
	if !utils.ContainsOwners(ctx.DataBackup.GetOwnerReferences(), dataset) {
		return r.AddOwnerAndRequeue(ctx, dataset)
	}

	ctx.Dataset = dataset
	return r.ReconcileDataBackup(ctx)
}

// AddOwnerAndRequeue adds Owner and requeue
func (r *DataBackupReconciler) AddOwnerAndRequeue(ctx reconcileRequestContext, dataset *datav1alpha1.Dataset) (ctrl.Result, error) {
	ctx.DataBackup.ObjectMeta.OwnerReferences = append(ctx.DataBackup.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	})
	if err := r.Update(ctx, &ctx.DataBackup); err != nil {
		ctx.Log.Error(err, "Failed to add ownerreference", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediately()
}

func (r *DataBackupReconciler) addFinalierAndRequeue(ctx reconcileRequestContext) (ctrl.Result, error) {
	ctx.DataBackup.ObjectMeta.Finalizers = append(ctx.DataBackup.ObjectMeta.Finalizers, cdatabackup.FINALIZER)
	ctx.Log.Info("Add finalizer and requeue", "finalizer", cdatabackup.FINALIZER)
	prevGeneration := ctx.DataBackup.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &ctx.DataBackup); err != nil {
		ctx.Log.Error(err, "failed to add finalizer to databackup", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, ctx.DataBackup.ObjectMeta.GetGeneration())
}

// SetupWithManager sets up the controller with the given controller manager
func (r *DataBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.DataBackup{}).
		Complete(r)
}
