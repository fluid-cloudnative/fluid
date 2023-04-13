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
	"reflect"
	"sync"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DataLoadReconcilerImplement implements the actual reconciliation logic of DataLoadReconciler
type DataLoadReconcilerImplement struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// NewDataLoadReconcilerImplement returns a DataLoadReconcilerImplement
func NewDataLoadReconcilerImplement(client client.Client, log logr.Logger, recorder record.EventRecorder) *DataLoadReconcilerImplement {
	r := &DataLoadReconcilerImplement{
		Client:   client,
		Log:      log,
		Recorder: recorder,
	}
	return r
}

// ReconcileDataLoadDeletion reconciles the deletion of the DataLoad
func (r *DataLoadReconcilerImplement) ReconcileDataLoadDeletion(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad, engines map[string]base.Engine, mutex *sync.Mutex) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataLoadDeletion")

	// 1. Delete release if exists
	releaseName := utils.GetDataLoadReleaseName(targetDataload.Name)
	err := helm.DeleteReleaseIfExists(releaseName, targetDataload.Namespace)
	if err != nil {
		log.Error(err, "can't delete release", "releaseName", releaseName)
		return utils.RequeueIfError(err)
	}

	// 2. Release lock on target dataset if necessary
	err = r.releaseLockOnTargetDataset(ctx, targetDataload, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", targetDataload.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 3. delete engine
	mutex.Lock()
	defer mutex.Unlock()
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	delete(engines, id)

	// 4. remove finalizer
	if utils.HasDeletionTimestamp(targetDataload.ObjectMeta) {
		targetDataload.ObjectMeta.Finalizers = utils.RemoveString(targetDataload.ObjectMeta.Finalizers, cdataload.DataloadFinalizer)
		if err := r.Update(ctx, &targetDataload); err != nil {
			log.Error(err, "failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		log.Info("Finalizer is removed")
	}
	return utils.NoRequeue()
}

// ReconcileDataLoad reconciles the DataLoad according to its phase status
func (r *DataLoadReconcilerImplement) ReconcileDataLoad(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad, engine base.Engine) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataLoad")
	log.V(1).Info("process the cdataload", "cdataload", targetDataload)

	// DataLoad's phase transition: None -> Pending -> Executing -> Loaded or Failed
	switch targetDataload.Status.Phase {
	case common.PhaseNone:
		return r.reconcileNoneDataLoad(ctx, targetDataload)
	case common.PhasePending:
		return r.reconcilePendingDataLoad(ctx, targetDataload, engine)
	case common.PhaseExecuting:
		return r.reconcileExecutingDataLoad(ctx, targetDataload, engine)
	case common.PhaseComplete:
		return r.reconcileLoadedDataLoad(ctx, targetDataload)
	case common.PhaseFailed:
		return r.reconcileFailedDataLoad(ctx, targetDataload)
	default:
		log.Info("Unknown DataLoad phase, won't reconcile it")
	}
	return utils.NoRequeue()
}

// reconcileNoneDataLoad reconciles DataLoads that are in `DataLoadPhaseNone` phase
func (r *DataLoadReconcilerImplement) reconcileNoneDataLoad(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNoneDataLoad")
	dataloadToUpdate := targetDataload.DeepCopy()
	dataloadToUpdate.Status.Phase = common.PhasePending
	if len(dataloadToUpdate.Status.Conditions) == 0 {
		dataloadToUpdate.Status.Conditions = []datav1alpha1.Condition{}
	}
	dataloadToUpdate.Status.Infos = map[string]string{}
	dataloadToUpdate.Status.Duration = "Unfinished"
	if err := r.Status().Update(context.TODO(), dataloadToUpdate); err != nil {
		log.Error(err, "failed to update the cdataload")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("Update phase of the cdataload to Pending successfully")
	return utils.RequeueImmediately()
}

// reconcilePendingDataLoad reconciles DataLoads that are in `DataLoadPhasePending` phase
func (r *DataLoadReconcilerImplement) reconcilePendingDataLoad(ctx cruntime.ReconcileRequestContext,
	targetDataload datav1alpha1.DataLoad,
	engine base.Engine) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePendingDataLoad")

	// 1. Check dataload namespace and dataset namespace need to be same
	if targetDataload.Namespace != targetDataload.Spec.Dataset.Namespace {
		r.Recorder.Eventf(&targetDataload,
			v1.EventTypeWarning,
			common.TargetDatasetNamespaceNotSame,
			"Dataload(%s) namespace is not same as dataset",
			targetDataload.Name)

		// Update DataLoad's phase to Failed, and no requeue
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			dataload, err := utils.GetDataLoad(r.Client, targetDataload.Name, targetDataload.Namespace)
			if err != nil {
				return err
			}
			dataloadToUpdate := dataload.DeepCopy()
			dataloadToUpdate.Status.Phase = common.PhaseFailed

			if !reflect.DeepEqual(dataloadToUpdate.Status, dataload.Status) {
				if err := r.Status().Update(ctx, dataloadToUpdate); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Error(err, "can't update dataload's phase status to Failed")
			return utils.RequeueIfError(err)
		}
		return utils.NoRequeue()
	}

	// 2. Check existence of the target dataset
	targetDataset, err := utils.GetDataset(r.Client, targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			log.Info("Target dataset not found", "targetDataset", targetDataload.Spec.Dataset)
			r.Recorder.Eventf(&targetDataload,
				v1.EventTypeWarning,
				common.TargetDatasetNotFound,
				"Target dataset(namespace: %s, name: %s) not found",
				targetDataload.Spec.Dataset.Namespace, targetDataload.Spec.Dataset.Name)
		} else {
			log.Error(err, "can't get target dataset", "targetDataset", targetDataload.Spec.Dataset)
		}
		return utils.RequeueAfterInterval(20 * time.Second)
	}
	log.V(1).Info("get target dataset", "targetDataset", targetDataset)

	// 3. Check if there's any Executing DataLoad jobs(conflict DataLoad)
	conflictDataLoadRef := targetDataset.GetDataOperationInProgress(cdataload.DataLoadLockName)
	myDataLoadRef := utils.GetDataLoadRef(targetDataload.Name, targetDataload.Namespace)
	if len(conflictDataLoadRef) != 0 && conflictDataLoadRef != myDataLoadRef {
		log.V(1).Info("Found other DataLoads that is in Executing phase, will backoff", "other DataLoad", conflictDataLoadRef)
		r.Recorder.Eventf(&targetDataload,
			v1.EventTypeNormal,
			common.DataLoadCollision,
			"Found other Dataload(%s) that is in Executing phase, will backoff",
			conflictDataLoadRef)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 4. Check if the bounded runtime is ready
	ready := engine.CheckRuntimeReady()
	if !ready {
		log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataset)
		r.Recorder.Eventf(&targetDataload,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 5. lock the target dataset. Make sure only one DataLoad can win the lock and
	// the losers have to requeue and go through the whole reconciliation loop.
	log.Info("No conflicts detected, try to lock the target dataset")
	datasetToUpdate := targetDataset.DeepCopy()
	datasetToUpdate.SetDataOperationInProgress(cdataload.DataLoadLockName, myDataLoadRef)

	if !reflect.DeepEqual(targetDataset.Status, datasetToUpdate.Status) {
		if err = r.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
			log.V(1).Info("fail to get target dataset's lock, will requeue")
			//todo(xuzhihao): random backoff
			return utils.RequeueAfterInterval(20 * time.Second)
		}
	}

	// 6. update phase to Executing
	// We offload the helm install logic to `reconcileExecutingDataLoad` to
	// avoid such a case that status.phase change successfully first but helm install failed,
	// where the DataLoad job will never start and all other DataLoads will be blocked forever.
	log.Info("Get lock on target dataset, try to update phase")
	dataLoadToUpdate := targetDataload.DeepCopy()
	dataLoadToUpdate.Status.Phase = common.PhaseExecuting
	if err = r.Client.Status().Update(context.TODO(), dataLoadToUpdate); err != nil {
		log.Error(err, "failed to update cdataload's status to Executing, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update cdataload's status to Executing successfully")
	return utils.RequeueImmediately()
}

// reconcileExecutingDataLoad reconciles DataLoads that are in `Executing` phase
func (r *DataLoadReconcilerImplement) reconcileExecutingDataLoad(ctx cruntime.ReconcileRequestContext,
	targetDataload datav1alpha1.DataLoad,
	engine base.Engine) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileExecutingDataLoad")

	// 1. Install the helm chart if not exists and requeue
	err := engine.LoadData(ctx, targetDataload)
	if err != nil {
		log.Error(err, "engine loaddata failed")
		return utils.RequeueIfError(err)
	}

	// 2. Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(targetDataload.Name)
	jobName := utils.GetDataLoadJobName(releaseName)
	log.V(1).Info("DataLoad chart already existed, check its running status")
	job, err := utils.GetDataLoadJob(r.Client, jobName, ctx.Namespace)
	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			log.Info("Related Job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				log.Error(err, "can't delete dataload release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return utils.RequeueIfError(err)
			}
			return utils.RequeueAfterInterval(20 * time.Second)
		}
		// other error
		log.Error(err, "can't get dataload job", "namespace", ctx.Namespace, "jobName", jobName)
		return utils.RequeueIfError(err)
	}

	if len(job.Status.Conditions) != 0 {
		if job.Status.Conditions[0].Type == batchv1.JobFailed ||
			job.Status.Conditions[0].Type == batchv1.JobComplete {
			// job either failed or complete, update DataLoad's phase status
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				dataload, err := utils.GetDataLoad(r.Client, targetDataload.Name, targetDataload.Namespace)
				if err != nil {
					return err
				}
				dataloadToUpdate := dataload.DeepCopy()
				jobCondition := job.Status.Conditions[0]
				dataloadToUpdate.Status.Conditions = []datav1alpha1.Condition{
					{
						Type:               common.ConditionType(jobCondition.Type),
						Status:             jobCondition.Status,
						Reason:             jobCondition.Reason,
						Message:            jobCondition.Message,
						LastProbeTime:      jobCondition.LastProbeTime,
						LastTransitionTime: jobCondition.LastTransitionTime,
					},
				}
				if jobCondition.Type == batchv1.JobFailed {
					dataloadToUpdate.Status.Phase = common.PhaseFailed
				} else {
					dataloadToUpdate.Status.Phase = common.PhaseComplete
				}
				dataloadToUpdate.Status.Duration = utils.CalculateDuration(dataloadToUpdate.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)

				if !reflect.DeepEqual(dataloadToUpdate.Status, dataload.Status) {
					if err := r.Status().Update(ctx, dataloadToUpdate); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				log.Error(err, "can't update dataload's phase status to Failed")
				return utils.RequeueIfError(err)
			}
			return utils.RequeueImmediately()
		}
	}

	log.V(1).Info("DataLoad job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return utils.RequeueAfterInterval(20 * time.Second)
}

// reconcileLoadedDataLoad reconciles DataLoads that are in `DataLoadPhaseComplete` phase
func (r *DataLoadReconcilerImplement) reconcileLoadedDataLoad(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileLoadedDataLoad")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, targetDataload, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", targetDataload.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. record event and no requeue
	log.Info("DataLoad Loaded, no need to requeue")
	jobName := utils.GetDataLoadJobName(utils.GetDataLoadReleaseName(targetDataload.Name))
	r.Recorder.Eventf(&targetDataload, v1.EventTypeNormal, common.DataLoadJobComplete, "DataLoad job %s succeeded", jobName)
	return utils.NoRequeue()
}

// reconcileFailedDataLoad reconciles DataLoads that are in `DataLoadPhaseComplete` phase
func (r *DataLoadReconcilerImplement) reconcileFailedDataLoad(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailedDataLoad")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, targetDataload, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", targetDataload.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. record event and no requeue
	log.Info("DataLoad failed, won't requeue")
	jobName := utils.GetDataLoadJobName(utils.GetDataLoadReleaseName(targetDataload.Name))
	r.Recorder.Eventf(&targetDataload, v1.EventTypeNormal, common.DataLoadJobFailed, "DataLoad job %s failed", jobName)
	return utils.NoRequeue()
}

// releaseLockOnTargetDataset releases lock on target dataset if the lock currently belongs to reconciling DataLoad.
// We use a key-value pair on the target dataset's status as the lock. To release the lock, we can simply set the value to empty.
func (r *DataLoadReconcilerImplement) releaseLockOnTargetDataset(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad, log logr.Logger) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(r.Client, targetDataload.Spec.Dataset.Name, targetDataload.Spec.Dataset.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				log.Info("can't find target dataset, won't release lock", "targetDataset", targetDataload.Spec.Dataset.Name)
				return nil
			}
			// other error
			return err
		}

		currentRef := dataset.GetDataOperationInProgress(cdataload.DataLoadLockName)
		if currentRef != utils.GetDataLoadRef(targetDataload.Name, targetDataload.Namespace) {
			log.Info("Found DataLoadRef inconsistent with the reconciling DataLoad, won't release this lock, ignore it", "DataLoadRef", currentRef)
			return nil
		}
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.RemoveDataOperationInProgress(cdataload.DataLoadLockName)
		if !reflect.DeepEqual(datasetToUpdate.Status, dataset.Status) {
			if err := r.Status().Update(ctx, datasetToUpdate); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
