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

package datamigrate

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

type DataMigrateReconcilerImplement struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

func NewDataMigrateReconcilerImplement(client client.Client, log logr.Logger, recorder record.EventRecorder) *DataMigrateReconcilerImplement {
	return &DataMigrateReconcilerImplement{
		Client:   client,
		Log:      log,
		Recorder: recorder,
	}
}

func (r *DataMigrateReconcilerImplement) ReconcileDataMigrateDeletion(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate, engines map[string]base.Engine, mutex *sync.Mutex) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataMigrateDeletion")

	dataset := ctx.Dataset
	// 1. Delete release if exists
	releaseName := utils.GetDataMigrateReleaseName(targetDataMigrate.Name)
	err := helm.DeleteReleaseIfExists(releaseName, targetDataMigrate.Namespace)
	if err != nil {
		log.Error(err, "can't delete release", "releaseName", releaseName)
		return utils.RequeueIfError(err)
	}

	// 2. Release lock on target dataset if necessary
	err = r.releaseLockOnTargetDataset(ctx, targetDataMigrate, dataset)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", dataset.Name)
		return utils.RequeueIfError(err)
	}

	// 3. delete engine
	mutex.Lock()
	defer mutex.Unlock()
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	delete(engines, id)

	// 4. remove finalizer
	if utils.HasDeletionTimestamp(targetDataMigrate.ObjectMeta) {
		targetDataMigrate.ObjectMeta.Finalizers = utils.RemoveString(targetDataMigrate.ObjectMeta.Finalizers, cdatamigrate.DataMigrateFinalizer)
		if err := r.Update(ctx, &targetDataMigrate); err != nil {
			log.Error(err, "failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		log.Info("Finalizer is removed")
	}
	return utils.NoRequeue()

}

// ReconcileDataMigrate reconciles the DataMigrate according to its phase status
func (r *DataMigrateReconcilerImplement) ReconcileDataMigrate(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate, engine base.Engine) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataMigrate")
	log.V(1).Info("process the cDataMigrate", "cDataMigrate", targetDataMigrate)

	// DataMigrate's phase transition: None -> Pending -> Executing -> Complete or Failed
	switch targetDataMigrate.Status.Phase {
	case common.PhaseNone:
		return r.reconcileNoneDataMigrate(ctx, targetDataMigrate)
	case common.PhasePending:
		return r.reconcilePendingDataMigrate(ctx, targetDataMigrate, engine)
	case common.PhaseExecuting:
		return r.reconcileExecutingDataMigrate(ctx, targetDataMigrate, engine)
	case common.PhaseComplete:
		return r.reconcileCompleteDataMigrate(ctx, targetDataMigrate)
	case common.PhaseFailed:
		return r.reconcileFailedDataMigrate(ctx, targetDataMigrate)
	default:
		log.Info("Unknown DataMigrate phase, won't reconcile it")
	}
	return utils.NoRequeue()
}

// reconcileNoneDataMigrate reconciles DataMigrates that are in `DataMigratePhaseNone` phase
func (r *DataMigrateReconcilerImplement) reconcileNoneDataMigrate(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNoneDataMigrate")
	DataMigrateToUpdate := targetDataMigrate.DeepCopy()
	DataMigrateToUpdate.Status.Phase = common.PhasePending
	if len(DataMigrateToUpdate.Status.Conditions) == 0 {
		DataMigrateToUpdate.Status.Conditions = []datav1alpha1.Condition{}
	}
	DataMigrateToUpdate.Status.Duration = "Unfinished"
	if err := r.Status().Update(context.TODO(), DataMigrateToUpdate); err != nil {
		log.Error(err, "failed to update the cDataMigrate")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("Update phase of the cDataMigrate to Pending successfully")
	return utils.RequeueImmediately()
}

// reconcilePendingDataMigrate reconciles DataMigrates that are in `DataMigratePhasePending` phase
func (r *DataMigrateReconcilerImplement) reconcilePendingDataMigrate(ctx cruntime.ReconcileRequestContext,
	targetDataMigrate datav1alpha1.DataMigrate,
	engine base.Engine) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePendingDataMigrate")

	targetDataSet := ctx.Dataset
	// 1. Check DataMigrate namespace and dataset namespace need to be same
	if targetDataMigrate.Namespace != targetDataSet.Namespace {
		r.Recorder.Eventf(&targetDataMigrate,
			v1.EventTypeWarning,
			common.TargetDatasetNamespaceNotSame,
			"DataMigrate(%s) namespace is not same as dataset",
			targetDataMigrate.Name)

		// Update DataMigrate's phase to Failed, and no requeue
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			DataMigrate, err := utils.GetDataMigrate(r.Client, targetDataMigrate.Name, targetDataMigrate.Namespace)
			if err != nil {
				return err
			}
			DataMigrateToUpdate := DataMigrate.DeepCopy()
			DataMigrateToUpdate.Status.Phase = common.PhaseFailed

			if !reflect.DeepEqual(DataMigrateToUpdate.Status, DataMigrate.Status) {
				if err := r.Status().Update(ctx, DataMigrateToUpdate); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			log.Error(err, "can't update DataMigrate's phase status to Failed")
			return utils.RequeueIfError(err)
		}
		return utils.NoRequeue()
	}

	// 2. Check if there's any Executing DataMigrate jobs(conflict DataMigrate)
	conflictDataMigrateRef := targetDataSet.Status.DataMigrateRef
	myDataMigrateRef := utils.GetDataMigrateRef(targetDataMigrate.Name, targetDataMigrate.Namespace)
	if len(conflictDataMigrateRef) != 0 && conflictDataMigrateRef != myDataMigrateRef {
		log.V(1).Info("Found other DataMigrates that is in Executing phase, will backoff", "other DataMigrate", conflictDataMigrateRef)
		r.Recorder.Eventf(&targetDataMigrate,
			v1.EventTypeNormal,
			common.DataMigrateCollision,
			"Found other DataMigrate(%s) that is in Executing phase, will backoff",
			conflictDataMigrateRef)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 3. Check if the bounded runtime is ready
	ready := engine.CheckRuntimeReady()
	if !ready {
		log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataSet)
		r.Recorder.Eventf(&targetDataMigrate,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 4. lock the target dataset. Make sure only one DataMigrate can win the lock and
	// the losers have to requeue and go through the whole reconciliation loop.
	log.Info("No conflicts detected, try to lock the target dataset")
	datasetToUpdate := targetDataSet.DeepCopy()
	datasetToUpdate.Status.DataMigrateRef = myDataMigrateRef
	if !reflect.DeepEqual(targetDataSet.Status, datasetToUpdate.Status) {
		if err := r.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
			log.V(1).Info("fail to get target dataset's lock, will requeue")
			//todo(xuzhihao): random backoff
			return utils.RequeueAfterInterval(20 * time.Second)
		}
	}

	// 5. update phase to Executing
	// We offload the helm install logic to `reconcileExecutingDataMigrate` to
	// avoid such a case that status.phase change successfully first but helm install failed,
	// where the DataMigrate job will never start and all other DataMigrates will be blocked forever.
	log.Info("Get lock on target dataset, try to update phase")
	DataMigrateToUpdate := targetDataMigrate.DeepCopy()
	DataMigrateToUpdate.Status.Phase = common.PhaseExecuting
	if err := r.Client.Status().Update(context.TODO(), DataMigrateToUpdate); err != nil {
		log.Error(err, "failed to update DataMigrate's status to Executing, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update DataMigrate's status to Executing successfully")

	// 6. update dataset phase to DataMigrating
	log.Info("Get lock on target dataset, try to update phase")
	DataSetToUpdate := targetDataSet.DeepCopy()
	DataSetToUpdate.Status.Phase = datav1alpha1.DataMigrating
	if err := r.Client.Status().Update(context.TODO(), DataSetToUpdate); err != nil {
		log.Error(err, "failed to update DataSet's status to DataMigrating, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update DataSet's status to DataMigrating successfully")
	return utils.RequeueImmediately()
}

// reconcileExecutingDataMigrate reconciles DataMigrates that are in `Executing` phase
func (r *DataMigrateReconcilerImplement) reconcileExecutingDataMigrate(ctx cruntime.ReconcileRequestContext,
	targetDataMigrate datav1alpha1.DataMigrate,
	engine base.Engine) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileExecutingDataMigrate")

	// 1. Install the helm chart if not exists and requeue
	err := engine.MigrateData(ctx, targetDataMigrate)
	if err != nil {
		log.Error(err, "engine migrate data failed")
		return utils.RequeueIfError(err)
	}

	// 2. Check running status of the DataMigrate job
	releaseName := utils.GetDataMigrateReleaseName(targetDataMigrate.Name)
	jobName := utils.GetDataMigrateJobName(releaseName)
	log.V(1).Info("DataMigrate chart already existed, check its running status")
	job, err := utils.GetDataMigrateJob(r.Client, jobName, ctx.Namespace)
	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			log.Info("Related Job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				log.Error(err, "can't delete DataMigrate release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return utils.RequeueIfError(err)
			}
			return utils.RequeueAfterInterval(20 * time.Second)
		}
		// other error
		log.Error(err, "can't get DataMigrate job", "namespace", ctx.Namespace, "jobName", jobName)
		return utils.RequeueIfError(err)
	}

	if len(job.Status.Conditions) != 0 {
		if job.Status.Conditions[0].Type == batchv1.JobFailed ||
			job.Status.Conditions[0].Type == batchv1.JobComplete {
			// job either failed or complete, update DataMigrate's phase status
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				DataMigrate, err := utils.GetDataMigrate(r.Client, targetDataMigrate.Name, targetDataMigrate.Namespace)
				if err != nil {
					return err
				}
				DataMigrateToUpdate := DataMigrate.DeepCopy()
				jobCondition := job.Status.Conditions[0]
				DataMigrateToUpdate.Status.Conditions = []datav1alpha1.Condition{
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
					DataMigrateToUpdate.Status.Phase = common.PhaseFailed
				} else {
					DataMigrateToUpdate.Status.Phase = common.PhaseComplete
				}
				DataMigrateToUpdate.Status.Duration = utils.CalculateDuration(DataMigrateToUpdate.CreationTimestamp.Time, jobCondition.LastTransitionTime.Time)

				if !reflect.DeepEqual(DataMigrateToUpdate.Status, DataMigrate.Status) {
					if err := r.Status().Update(ctx, DataMigrateToUpdate); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				log.Error(err, "can't update DataMigrate's phase status to Failed")
				return utils.RequeueIfError(err)
			}
			return utils.RequeueImmediately()
		}
	}

	log.V(1).Info("DataMigrate job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return utils.RequeueAfterInterval(20 * time.Second)
}

// reconcileCompleteDataMigrate reconciles DataMigrates that are in `DataMigratePhaseComplete` phase
func (r *DataMigrateReconcilerImplement) reconcileCompleteDataMigrate(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileCompleteDataMigrate")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, targetDataMigrate, ctx.Dataset)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. record event and no requeue
	log.Info("DataMigrate complete, no need to requeue")
	jobName := utils.GetDataMigrateJobName(utils.GetDataMigrateReleaseName(targetDataMigrate.Name))
	r.Recorder.Eventf(&targetDataMigrate, v1.EventTypeNormal, common.DataMigrateJobComplete, "DataMigrate job %s succeeded", jobName)
	return utils.NoRequeue()
}

// reconcileFailedDataMigrate reconciles DataMigrates that are in `DataMigratePhaseComplete` phase
func (r *DataMigrateReconcilerImplement) reconcileFailedDataMigrate(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailedDataMigrate")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, targetDataMigrate, ctx.Dataset)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. record event and no requeue
	log.Info("DataMigrate failed, won't requeue")
	jobName := utils.GetDataMigrateJobName(utils.GetDataMigrateReleaseName(targetDataMigrate.Name))
	r.Recorder.Eventf(&targetDataMigrate, v1.EventTypeNormal, common.DataMigrateJobFailed, "DataMigrate job %s failed", jobName)
	return utils.NoRequeue()
}

// releaseLockOnTargetDataset releases lock on target dataset if the lock currently belongs to reconciling DataMigrate.
// We use a key-value pair on the target dataset's status as the lock. To release the lock, we can simply set the value to empty.
func (r *DataMigrateReconcilerImplement) releaseLockOnTargetDataset(ctx cruntime.ReconcileRequestContext, targetDataMigrate datav1alpha1.DataMigrate, targetDataSet *datav1alpha1.Dataset) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if targetDataSet.Status.DataMigrateRef != utils.GetDataMigrateRef(targetDataMigrate.Name, targetDataMigrate.Namespace) {
			r.Log.Info("Found DataMigrateRef inconsistent with the reconciling DataMigrate, won't release this lock, ignore it", "DataMigrateRef", targetDataSet.Status.DataMigrateRef)
			return nil
		}
		datasetToUpdate := targetDataSet.DeepCopy()
		datasetToUpdate.Status.DataMigrateRef = ""
		datasetToUpdate.Status.Phase = datav1alpha1.BoundDatasetPhase
		if !reflect.DeepEqual(datasetToUpdate.Status, targetDataSet) {
			if err := r.Status().Update(ctx, datasetToUpdate); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
