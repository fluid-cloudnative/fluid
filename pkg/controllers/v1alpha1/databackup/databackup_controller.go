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
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
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
	r.DataBackupReconcilerImplement = NewDataBackupReconcilerImplement(client, log, recorder, r)
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
	ctx.Log.Info("DataBackup found", "detail", databackup)

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

	// DataBackup's phase transition: None -> Pending -> Backuping -> Complete or Failed
	switch databackup.Status.Phase {
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
		ctx.Log.Info("Unknown DataBackup phase, won't reconcile it", "databackup", databackup)
	}
	return utils.NoRequeue()

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

// reconcileNoneDataBackup reconciles DataBackups that are in `None` phase
func (r *DataBackupReconciler) reconcileNoneDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
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

// reconcileCompleteDataBackup reconciles DataBackups that are in `Complete` phase
func (r *DataBackupReconciler) reconcileCompleteDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileCompleteDataBackup")
	// 1. Update BackupPath of the databackup
	databackupToUpdate := ctx.DataBackup.DeepCopy()
	databackupToUpdate.Status.BackupLocation.Path = databackupToUpdate.Spec.BackupPath
	if strings.HasPrefix(databackupToUpdate.Spec.BackupPath, common.PathScheme) {
		podName := databackupToUpdate.Name + "-pod"
		backupPod, err := kubeclient.GetPodByName(r.Client, podName, ctx.Namespace)
		if err != nil {
			log.Error(err, "Failed to get backup pod")
			return utils.RequeueIfError(err)
		}
		databackupToUpdate.Status.BackupLocation.NodeName = backupPod.Spec.NodeName
	} else {
		databackupToUpdate.Status.BackupLocation.NodeName = "NA"
	}
	if err := r.Status().Update(context.TODO(), databackupToUpdate); err != nil {
		log.Error(err, "the backup pod has completd, but failed to  update the databackup")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("Update BackupPath of the databackup  successfully")
	// 2. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataBackup.Spec.Dataset)
		return utils.RequeueIfError(err)
	}
	// 3. record and no requeue
	log.Info("DataBackup success, no need to requeue")
	r.Recorder.Eventf(&ctx.DataBackup, v1.EventTypeNormal, common.DataBackupFailed, "DataBackup %s failed", ctx.DataBackup.Name)
	return utils.NoRequeue()
}

// reconcileFailedDataBackup reconciles DataBackups that are in `Failed` phase
func (r *DataBackupReconciler) reconcileFailedDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailedDatabackup")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataBackup.Spec.Dataset)
		return utils.RequeueIfError(err)
	}
	// 2. record and no requeue
	log.Info("DataBackup failed, won't requeue")
	r.Recorder.Eventf(&ctx.DataBackup, v1.EventTypeNormal, common.DataBackupComplete, "DataBackup %s succeeded", ctx.DataBackup.Name)
	return utils.NoRequeue()
}

// releaseLockOnTargetDataset releases lock on target dataset if the lock currently belongs to reconciling DataLoad.
// We use a key-value pair on the target dataset's status as the lock. To release the lock, we can simply set the value to empty.
func (r *DataBackupReconciler) releaseLockOnTargetDataset(ctx reconcileRequestContext, log logr.Logger) error {
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

// SetupWithManager sets up the controller with the given controller manager
func (r *DataBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.DataBackup{}).
		Complete(r)
}
