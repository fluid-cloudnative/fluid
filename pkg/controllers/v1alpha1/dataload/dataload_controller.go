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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// DataLoadReconciler reconciles a DataLoad object
type DataLoadReconciler struct {
	Scheme  *runtime.Scheme
	engines map[string]base.Engine
	mutex   *sync.Mutex
	*DataLoadReconcilerImplement
}

// NewDataLoadReconciler returns a DataLoadReconciler
func NewDataLoadReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *DataLoadReconciler {
	r := &DataLoadReconciler{
		Scheme:  scheme,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.DataLoadReconcilerImplement = NewDataLoadReconcilerImplement(client, log, recorder)
	return r
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=dataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=dataloads/status,verbs=get;update;patch
// Reconcile reconciles the DataLoad object
func (r *DataLoadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := cruntime.ReconcileRequestContext{
		Context:  context.Background(),
		Log:      r.Log.WithValues("dataload", req.NamespacedName),
		Recorder: r.Recorder,
		Client:   r.Client,
		Category: common.AccelerateCategory,
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

	targetDataload := *dataload
	ctx.Log.V(1).Info("DataLoad found", "detail", dataload)

	// 2. Reconcile deletion of the object if necessary
	if utils.HasDeletionTimestamp(dataload.ObjectMeta) {
		return r.ReconcileDataLoadDeletion(ctx)
	}

	// 3. get the dataset
	dataset, err := utils.GetDataset(r.Client, targetDataload.Spec.Dataset.Name, req.Namespace)
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
	if !utils.ContainsString(targetDataload.ObjectMeta.GetFinalizers(), cdataload.DATALOAD_FINALIZER) {
		return r.addFinalierAndRequeue(ctx, targetDataload)
	}

	// 5. add owner and requeue
	if !utils.ContainsOwners(targetDataload.GetOwnerReferences(), dataset) {
		return r.AddOwnerAndRequeue(ctx, targetDataload, targetDataset)
	}

	return r.ReconcileDataLoad(ctx)
}

// AddOwnerAndRequeue adds Owner and requeue
func (r *DataLoadReconciler) AddOwnerAndRequeue(ctx cruntime.ReconcileRequestContext, targetDataLoad datav1alpha1.DataLoad, targetDataset *datav1alpha1.Dataset) (ctrl.Result, error) {
	targetDataset.ObjectMeta.OwnerReferences = append(targetDataLoad.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: targetDataset.APIVersion,
		Kind:       targetDataset.Kind,
		Name:       targetDataset.Name,
		UID:        targetDataset.UID,
	})
	if err := r.Update(ctx, &targetDataLoad); err != nil {
		ctx.Log.Error(err, "Failed to add ownerreference", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediately()
}

func (r *DataLoadReconciler) addFinalierAndRequeue(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (ctrl.Result, error) {
	targetDataload.ObjectMeta.Finalizers = append(targetDataload.ObjectMeta.Finalizers, cdataload.DATALOAD_FINALIZER)
	ctx.Log.Info("Add finalizer and requeue", "finalizer", cdataload.DATALOAD_FINALIZER)
	prevGeneration := targetDataload.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &targetDataload); err != nil {
		ctx.Log.Error(err, "failed to add finalizer to dataload", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, targetDataload.ObjectMeta.GetGeneration())
}

// SetupWithManager sets up the controller with the given controller manager
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.DataLoad{}).
		Complete(r)
}
