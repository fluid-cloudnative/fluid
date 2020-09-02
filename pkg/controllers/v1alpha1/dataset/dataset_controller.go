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

package dataset

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

const (
	finalizer = "fluid-dataset-controller-finalizer"
)

// DatasetReconciler reconciles a Dataset object
type DatasetReconciler struct {
	client.Client
	Recorder     record.EventRecorder
	Log          logr.Logger
	Scheme       *runtime.Scheme
	ResyncPeriod time.Duration
}

type reconcileRequestContext struct {
	context.Context
	Log     logr.Logger
	Dataset datav1alpha1.Dataset
	types.NamespacedName
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=datasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=datasets/status,verbs=get;update;patch

func (r *DatasetReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := reconcileRequestContext{
		Context:        context.Background(),
		Log:            r.Log.WithValues("dataset", req.NamespacedName),
		NamespacedName: req.NamespacedName,
	}

	notFound := false
	ctx.Log.V(1).Info("process the request", "request", req)

	/*
		###1. Load the dataset
	*/
	if err := r.Get(ctx, req.NamespacedName, &ctx.Dataset); err != nil {
		ctx.Log.Info("Unable to fetch Dataset", "reason", err)
		if utils.IgnoreNotFound(err) != nil {
			r.Log.Error(err, "failed to get dataset")
			return ctrl.Result{}, err
		} else {
			notFound = true
		}
	} else {
		return r.reconcileDataset(ctx)
	}

	/*
		### 2. we'll ignore not-found errors, since they can't be fixed by an immediate
		 requeue (we'll need to wait for a new notification), and we can get them
		 on deleted requests.
	*/
	if notFound {
		ctx.Log.V(1).Info("Not found.")
	}
	return ctrl.Result{}, nil
}

// reconcile Dataset
func (r *DatasetReconciler) reconcileDataset(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileDataset")
	log.V(1).Info("process the dataset", "dataset", ctx.Dataset)
	// 1. Check if need to delete dataset
	if utils.HasDeletionTimestamp(ctx.Dataset.ObjectMeta) {
		return r.reconcileDatasetDeletion(ctx)
	}

	// 2.Add finalizer
	if !utils.ContainsString(ctx.Dataset.ObjectMeta.GetFinalizers(), finalizer) {
		return r.addFinalizerAndRequeue(ctx)
	}

	if ctx.Dataset.Status.Phase == datav1alpha1.NoneDatasetPhase {
		dataset := ctx.Dataset.DeepCopy()
		dataset.Status.Phase = datav1alpha1.NotBoundDatasetPhase
		if len(dataset.Status.Conditions) == 0 {
			dataset.Status.Conditions = []datav1alpha1.DatasetCondition{}
		}
		if err := r.Status().Update(ctx, dataset); err != nil {
			ctx.Log.Error(err, "Failed to update the dataset", "StatusUpdateError", ctx)
			return utils.RequeueIfError(err)
		} else {
			ctx.Log.V(1).Info("Update the status of the dataset successfully", "phase", dataset.Status.Phase)
		}
	}

	// return utils.RequeueAfterInterval(r.ResyncPeriod)
	return utils.NoRequeue()
}

// reconcile Dataset Deletion
func (r *DatasetReconciler) reconcileDatasetDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileDatasetDeletion")
	log.V(1).Info("process the dataset", "dataset", ctx.Dataset)

	// 1. If runtime is not deleted, then requeue
	if ctx.Dataset.Status.Phase == datav1alpha1.BoundDatasetPhase ||
		ctx.Dataset.Status.Phase == datav1alpha1.FailedDatasetPhase ||
		ctx.Dataset.Status.Phase == datav1alpha1.PendingDatasetPhase {
		log.Info("The dataset is failed or bounded, can't be deleted.")
		return utils.RequeueAfterInterval(time.Duration(1 * time.Second))
	}

	// 2. Remove finalizer
	if !ctx.Dataset.ObjectMeta.GetDeletionTimestamp().IsZero() {
		ctx.Dataset.ObjectMeta.Finalizers = utils.RemoveString(ctx.Dataset.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, &ctx.Dataset); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		ctx.Log.V(1).Info("Finalizer is removed", "dataset", ctx.Dataset)
	}

	return ctrl.Result{}, nil
}

func (r *DatasetReconciler) addFinalizerAndRequeue(ctx reconcileRequestContext) (ctrl.Result, error) {
	ctx.Dataset.ObjectMeta.Finalizers = append(ctx.Dataset.ObjectMeta.Finalizers, finalizer)
	ctx.Log.Info("Add finalizer and Requeue")
	prevGeneration := ctx.Dataset.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &ctx.Dataset); err != nil {
		ctx.Log.Error(err, "Failed to add finalizer", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, ctx.Dataset.ObjectMeta.GetGeneration())
}

func (r *DatasetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.Dataset{}).
		Complete(r)
}
