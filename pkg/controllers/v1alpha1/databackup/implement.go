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
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

// DataBackupReconcilerImplement implements the actual reconciliation logic of DataBackupReconciler
type DataBackupReconcilerImplement struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
}

// NewDataBackupReconcilerImplement returns a DataBackupReconcilerImplement
func NewDataBackupReconcilerImplement(client client.Client, log logr.Logger, recorder record.EventRecorder) *DataBackupReconcilerImplement {
	r := &DataBackupReconcilerImplement{
		Client:   client,
		Log:      log,
		Recorder: recorder,
	}
	return r
}

// ReconcileDataBackupDeletion reconciles the deletion of the DataBackup
func (r *DataBackupReconcilerImplement) ReconcileDataBackupDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataBackupDeletion")

	// 1. Release lock on target dataset if necessary
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataBackup.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. remove finalizer
	if utils.HasDeletionTimestamp(ctx.DataBackup.ObjectMeta) {
		ctx.DataBackup.ObjectMeta.Finalizers = utils.RemoveString(ctx.DataBackup.ObjectMeta.Finalizers, cdatabackup.FINALIZER)
		if err := r.Update(ctx, &ctx.DataBackup); err != nil {
			log.Error(err, "failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		log.Info("Finalizer is removed")
	}
	return utils.NoRequeue()
}

// ReconcileDataBackup reconciles the DataBackup according to its phase status
func (r *DataBackupReconcilerImplement) ReconcileDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataBackup")
	log.V(1).Info("process the cdatabackup", "cdatabackup", ctx.DataBackup)

	// DataBackup's phase transition: None -> Pending -> Backuping -> Complete or Failed
	switch ctx.DataBackup.Status.Phase {
	case cdatabackup.PhaseNone:
		return r.reconcileNoneDataBackup(ctx)
	case cdatabackup.PhasePending:
		return r.reconcilePendingDataBackup(ctx)
	case cdatabackup.PhaseBackuping:
		return r.reconcileBackupingDataBackup(ctx)
	case cdatabackup.PhaseComplete:
		return r.reconcileCompleteDataBackup(ctx)
	case cdatabackup.PhaseFailed:
		return r.reconcileFailedDataBackup(ctx)
	default:
		log.Info("Unknown DataBackup phase, won't reconcile it")
	}
	return utils.NoRequeue()
}

// reconcileNoneDataBackup reconciles DataBackups that are in `None` phase
func (r *DataBackupReconcilerImplement) reconcileNoneDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNoneDataBackup")
	databackupToUpdate := ctx.DataBackup.DeepCopy()
	databackupToUpdate.Status.Phase = cdatabackup.PhasePending
	if len(databackupToUpdate.Status.Conditions) == 0 {
		databackupToUpdate.Status.Conditions = []v1alpha1.DataBackupCondition{}
	}
	if err := r.Status().Update(context.TODO(), databackupToUpdate); err != nil {
		log.Error(err, "failed to update the databackup")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("Update phase of the databackup to Pending successfully")
	return utils.RequeueImmediately()
}

// reconcilePendingDataBackup reconciles DataBackups that are in `Pending` phase
func (r *DataBackupReconcilerImplement) reconcilePendingDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePendingDataBackup")
	targetDataset := ctx.Dataset
	// 1. Check if there's any Backinging pods(conflict DataBackup)
	conflictDataBackupRef := targetDataset.Status.DataBackupRef
	myDataBackupRef := utils.GetDataBackupRef(ctx.DataBackup.Name, ctx.DataBackup.Namespace)
	if len(conflictDataBackupRef) != 0 && conflictDataBackupRef != myDataBackupRef {
		log.V(1).Info("Found other DataBackups that is in Backuping phase, will backoff", "other DataBackup", conflictDataBackupRef)

		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Conditions = []v1alpha1.DataBackupCondition{
			{
				Type:               cdatabackup.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "conflictDataBackupRef",
				Message:            "Found other Databackup that is in Backinging phase",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}
		databackupToUpdate.Status.Phase = cdatabackup.PhaseFailed

		if err := r.Status().Update(ctx, databackupToUpdate); err != nil {
			return utils.RequeueIfError(err)
		}
		return utils.RequeueImmediately()
	}

	// 2. Check if the bounded runtime is ready
	var ready bool
	index, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)
	if index == -1 {
		log.Info("bounded runtime with Accelerate Category is not found on the target dataset", "targetDataset", targetDataset)
	}
	switch boundedRuntime.Type {
	case common.ALLUXIO_RUNTIME:
		podName := fmt.Sprintf("%s-master-0", targetDataset.Name)
		containerName := "alluxio-master"
		fileUtils := operations.NewAlluxioFileUtils(podName, containerName, targetDataset.Namespace, ctx.Log)
		ready = fileUtils.Ready()
	default:
		log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "runtime", boundedRuntime)
		r.Recorder.Eventf(&ctx.DataBackup,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not supported")
	}

	if !ready {
		log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataset)
		r.Recorder.Eventf(&ctx.DataBackup,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 3. check the path
	if !strings.HasPrefix(ctx.DataBackup.Spec.BackupPath, common.PathScheme) && !strings.HasPrefix(ctx.DataBackup.Spec.BackupPath, common.VolumeScheme) {
		log.Error(fmt.Errorf("PathNotSupported"), "don't support path in this form", "path", ctx.DataBackup.Spec.BackupPath)
		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Conditions = []v1alpha1.DataBackupCondition{
			{
				Type:               cdatabackup.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "PathNotSupported",
				Message:            "Only support pvc and local path now",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}
		databackupToUpdate.Status.Phase = cdatabackup.PhaseFailed

		if err := r.Status().Update(ctx, databackupToUpdate); err != nil {
			return utils.RequeueIfError(err)
		}
		return utils.RequeueImmediately()
	}

	// 3. lock the target dataset
	// only one Databackup can win the lock
	// the losers not need to backup again
	log.Info("No conflicts detected, try to lock the target dataset")
	datasetToUpdate := targetDataset.DeepCopy()
	datasetToUpdate.Status.DataBackupRef = myDataBackupRef
	if !reflect.DeepEqual(targetDataset.Status, datasetToUpdate.Status) {
		if err := r.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
			log.V(1).Info("fail to get target dataset's lock, will requeue")
			return utils.RequeueAfterInterval(20 * time.Second)
		}
	}
	// 4. update phase to Backuping
	log.Info("Get lock on target dataset, try to update phase")
	dataBackupToUpdate := ctx.DataBackup.DeepCopy()
	dataBackupToUpdate.Status.Phase = cdatabackup.PhaseBackuping
	if err := r.Client.Status().Update(context.TODO(), dataBackupToUpdate); err != nil {
		log.Error(err, "failed to update cdatabackup's status to Backuping, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update cdatabackup's status to Backuping successfully")
	return utils.RequeueImmediately()
}

// reconcileBackupingDataBackup reconciles DataBackups that are in `Backuping` phase
func (r *DataBackupReconcilerImplement) reconcileBackupingDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileBackupingDataBackup")
	// 1. get the alluxio-master Pod
	podName := ctx.Dataset.Name + "-master-0"
	masterPod, err := kubeclient.GetPodByName(r.Client, podName, ctx.Namespace)
	if err != nil{
		log.Error(err, "Failed to get alluxio-master")
		return utils.RequeueIfError(err)
	}
	// 2. create backup Pod if not exist
	backupPod, _ := kubeclient.GetPodByName(r.Client, ctx.DataBackup.Name, ctx.Namespace)
	if backupPod == nil {
		err = utils.CreateBackupPod(r.Client, masterPod, ctx.Dataset, &ctx.DataBackup)
		if err != nil{
			log.Error(err, "Failed to create backup Pod")
			return utils.RequeueIfError(err)
		}
	}

	// 3. Check running status of the DataBackup Pod
	backupPod, err = kubeclient.GetPodByName(r.Client, ctx.DataBackup.Name, ctx.Namespace)
	if kubeclient.IsSucceededPod(backupPod) {
		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Phase = cdatabackup.PhaseComplete
		if err := r.Status().Update(context.TODO(), databackupToUpdate); err != nil {
			log.Error(err, "the backup pod has completd, but failed to update the databackup")
			return utils.RequeueIfError(err)
		}
		log.V(1).Info("Update phase of the databackup to Complete successfully")
		return utils.RequeueImmediately()
	} else if kubeclient.IsFailedPod(backupPod){
		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Phase = cdatabackup.PhaseFailed
		if err := r.Status().Update(context.TODO(), databackupToUpdate); err != nil {
			log.Error(err, "the backup pod has failed, but failed to update the databackup")
			return utils.RequeueIfError(err)
		}
		log.V(1).Info("Update phase of the databackup to Failed successfully")
		return utils.RequeueImmediately()
	}
	return utils.RequeueAfterInterval(20 * time.Second)
}

// reconcileCompleteDataBackup reconciles DataBackups that are in `Complete` phase
func (r *DataBackupReconcilerImplement) reconcileCompleteDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileCompleteDataBackup")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataBackup.Spec.Dataset)
		return utils.RequeueIfError(err)
	}
	// 2. record and no requeue
	log.Info("DataBackup success, no need to requeue")
	return utils.NoRequeue()
}

// reconcileFailedDataBackup reconciles DataBackups that are in `Failed` phase
func (r *DataBackupReconcilerImplement) reconcileFailedDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailedDatabackup")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataBackup.Spec.Dataset)
		return utils.RequeueIfError(err)
	}
	// 2. record and no requeue
	log.Info("DataBackup failed, won't requeue")
	return utils.NoRequeue()
}

// releaseLockOnTargetDataset releases lock on target dataset if the lock currently belongs to reconciling DataLoad.
// We use a key-value pair on the target dataset's status as the lock. To release the lock, we can simply set the value to empty.
func (r *DataBackupReconcilerImplement) releaseLockOnTargetDataset(ctx reconcileRequestContext, log logr.Logger) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(r.Client, ctx.DataBackup.Spec.Dataset, ctx.DataBackup.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				log.Info("can't find target dataset, won't release lock", "targetDataset", ctx.DataBackup.Spec.Dataset)
				return nil
			}
			// other error
			return err
		}
		if dataset.Status.DataBackupRef != utils.GetDataBackupRef(ctx.DataBackup.Name, ctx.DataBackup.Namespace) {
			log.Info("Found DataBackuRef inconsistent with the reconciling DataBack, won't release this lock, ignore it", "DataLoadRef", dataset.Status.DataLoadRef)
			return nil
		}
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.Status.DataBackupRef = ""
		if !reflect.DeepEqual(datasetToUpdate.Status, dataset) {
			if err := r.Status().Update(ctx, datasetToUpdate); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
