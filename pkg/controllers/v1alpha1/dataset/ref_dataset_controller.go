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
	"errors"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/core/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// reconcile for dataset which reference to other dataset

const (
	ref_finalizer = "fluid-ref-dataset-controller-finalizer"
)

// RefDatasetReconciler reconciles a dataset object which reference to other dataset
type RefDatasetReconciler struct {
	client.Client
	Recorder     record.EventRecorder
	Log          logr.Logger
	Scheme       *runtime.Scheme
	ResyncPeriod time.Duration
}

type refReconcileRequestContext struct {
	context.Context
	Log     logr.Logger
	Dataset datav1alpha1.Dataset
	types.NamespacedName
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=datasets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=datasets/status,verbs=get;update;patch

func (r *RefDatasetReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx := refReconcileRequestContext{
		Context:        context,
		Log:            r.Log.WithValues("refdataset", req.NamespacedName),
		NamespacedName: req.NamespacedName,
	}

	notFound := false
	ctx.Log.Info("process the request", "request", req)

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
		// this controller only handles ref dataset
		var pDatasets []*datav1alpha1.Dataset
		mounts := ctx.Dataset.Spec.Mounts
		for _, mount := range mounts {
			if common.IsFluidRefSchema(mount.MountPoint) {
				datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
				namespaceAndName := strings.Split(datasetPath, "/")
				dataset, err := utils.GetDataset(r.Client, namespaceAndName[1], namespaceAndName[0])
				if err != nil {
					r.Log.Error(err, "failed to reconcile dataset")
					return utils.RequeueAfterInterval(r.ResyncPeriod)
				}
				pDatasets = append(pDatasets, dataset)
			}
		}
		// dataset can not mount both dataset:// and other schema
		if pDatasets != nil && len(pDatasets) != len(mounts) {
			err := errors.New("dataset can not mount both dataset:// and other schema")
			r.Log.Error(err, "failed to reconcile dataset")
			r.Recorder.Eventf(&ctx.Dataset, v1.EventTypeWarning, common.ErrorProcessDatasetReason,
				"Failed to process dataset because err: %s", err)
			return utils.NoRequeue()
		}
		// currently not support recursively reference
		for _, dataset := range pDatasets {
			if r.isRef(dataset) {
				err := errors.New("dataset can not mount recursively")
				r.Log.Error(err, "failed to reconcile dataset")
				r.Recorder.Eventf(&ctx.Dataset, v1.EventTypeWarning, common.ErrorProcessDatasetReason,
					"Failed to process dataset because err: %s", err)
				return utils.NoRequeue()
			}
		}

		// only handle dataset which references to other dataset
		if pDatasets != nil {
			return r.reconcileRefDataset(ctx, pDatasets)
		} else {
			// handle by dataset_controller.go
			return utils.NoRequeue()
		}
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

// reconcile ref Dataset
func (r *RefDatasetReconciler) reconcileRefDataset(ctx refReconcileRequestContext,
	pDatasets []*datav1alpha1.Dataset) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileRefDataset")
	log.Info("process the dataset", "dataset", ctx.Dataset)
	// 1. Check if need to delete dataset
	if utils.HasDeletionTimestamp(ctx.Dataset.ObjectMeta) {
		return r.reconcileDatasetDeletion(ctx)
	}

	// 2.Add finalizer
	if !utils.ContainsString(ctx.Dataset.ObjectMeta.GetFinalizers(), ref_finalizer) {
		return r.addFinalizerAndRequeue(ctx)
	}

	// 3. add to its referenced dataset' spec DatasetRef field
	datasetRefName := getDatasetRef(ctx.Dataset.Name, ctx.Dataset.Namespace)
	allUpdated := true
	for _, dataset := range pDatasets {
		// update dataset ref
		if !utils.ContainsString(dataset.Status.DatasetRef, datasetRefName) {
			newDataset := dataset.DeepCopy()
			newDataset.Status.DatasetRef = append(newDataset.Status.DatasetRef, datasetRefName)
			err := r.Status().Update(context.TODO(), newDataset)
			if err != nil {
				allUpdated = false
				log.Error(err, "modify dataset ref failed, requeue and retry")
			}
		}
	}
	log.Info("update referenced dataset", "allUpdate", allUpdated)
	if !allUpdated {
		return utils.RequeueImmediately()
	}

	// 4. create pv and pvc
	for _, dataset := range pDatasets {
		runtimeInfo, err := base.GetRuntimeInfo(r.Client, dataset.Name, dataset.Namespace)
		if err != nil {
			log.Error(err, "get the runtime failed")
			return utils.RequeueAfterInterval(r.ResyncPeriod)
		}
		err = createPersistentVolumeForRefDataset(r.Client, ctx.Dataset.Name, ctx.Dataset.Namespace, runtimeInfo, log)
		if err != nil {
			log.Error(err, "get the pv failed")
			return utils.RequeueAfterInterval(r.ResyncPeriod)
		}
		err = createPersistentVolumeClaimForRefDataset(r.Client, ctx.Dataset.Name, ctx.Dataset.Namespace, runtimeInfo)
		if err != nil {
			log.Error(err, "get the pvc failed")
			return utils.RequeueAfterInterval(r.ResyncPeriod)
		}
	}
	// return utils.RequeueAfterInterval(r.ResyncPeriod)
	return utils.NoRequeue()
}

// reconcile Dataset Deletion
func (r *RefDatasetReconciler) reconcileDatasetDeletion(ctx refReconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileRefDatasetDeletion")
	log.Info("process the dataset", "dataset", ctx.Dataset)

	// 1.if there is a pod which is using the dataset (or cannot judge), then requeue
	err := kubeclient.ShouldDeleteDataset(r.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		ctx.Log.Error(err, "Failed to delete dataset", "DatasetDeleteError", ctx)
		r.Recorder.Eventf(&ctx.Dataset, v1.EventTypeWarning, common.ErrorDeleteDataset, "Failed to delete dataset because err: %s", err.Error())
		return utils.RequeueAfterInterval(r.ResyncPeriod)
	}

	// 2. remove referenced dataset's DatasetRef field
	datasetRefName := getDatasetRef(ctx.Dataset.Name, ctx.Dataset.Namespace)
	mounts := ctx.Dataset.Spec.Mounts
	for _, mount := range mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
			namespaceAndName := strings.Split(datasetPath, "/")
			dataset, err := utils.GetDataset(r.Client, namespaceAndName[1], namespaceAndName[0])
			if err != nil {
				r.Log.Error(err, "failed to reconcile dataset")
				return utils.RequeueAfterInterval(r.ResyncPeriod)
			}
			if utils.ContainsString(dataset.Status.DatasetRef, datasetRefName) {
				newDataset := dataset.DeepCopy()
				newDataset.Status.DatasetRef = utils.RemoveString(newDataset.Status.DatasetRef, datasetRefName)
				err := r.Status().Update(context.TODO(), newDataset)
				if err != nil {
					log.Error(err, "modify dataset ref failed, requeue and retry")
					return utils.RequeueAfterInterval(r.ResyncPeriod)
				}
			}
		}
	}

	// 3. delete pv and pvc
	err = kubeclient.DeletePersistentVolume(r.Client, getPvName(ctx.Dataset.Name, ctx.Dataset.Namespace))
	if err != nil {
		log.Error(err, "delete pv failed, requeue and retry")
		return utils.RequeueAfterInterval(r.ResyncPeriod)
	}
	err = kubeclient.DeletePersistentVolumeClaim(r.Client, ctx.Dataset.Name, ctx.Dataset.Namespace)
	if err != nil {
		log.Error(err, "delete pv failed, requeue and retry")
		return utils.RequeueAfterInterval(r.ResyncPeriod)
	}

	// 4. Remove finalizer
	if !ctx.Dataset.ObjectMeta.GetDeletionTimestamp().IsZero() {
		ctx.Dataset.ObjectMeta.Finalizers = utils.RemoveString(ctx.Dataset.ObjectMeta.Finalizers, ref_finalizer)
		if err := r.Update(ctx, &ctx.Dataset); err != nil {
			log.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		ctx.Log.Info("Finalizer is removed", "dataset", ctx.Dataset)
	}

	log.Info("delete the dataset successfully", "dataset", ctx.Dataset)

	return ctrl.Result{}, nil
}

func (r *RefDatasetReconciler) addFinalizerAndRequeue(ctx refReconcileRequestContext) (ctrl.Result, error) {
	ctx.Dataset.ObjectMeta.Finalizers = append(ctx.Dataset.ObjectMeta.Finalizers, ref_finalizer)
	ctx.Log.Info("Add finalizer and Requeue")
	prevGeneration := ctx.Dataset.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &ctx.Dataset); err != nil {
		ctx.Log.Error(err, "Failed to add finalizer", "StatusUpdateError", ctx)
		return utils.RequeueIfError(err)
	}

	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, ctx.Dataset.ObjectMeta.GetGeneration())
}

func (r *RefDatasetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.Dataset{}).
		Complete(r)
}

func (r *RefDatasetReconciler) isRef(dataset *datav1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			return true
		}
	}
	return false
}
