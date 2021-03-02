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
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

// DataBackupReconcilerImplement implements the actual reconciliation logic of DataBackupReconciler
type DataBackupReconcilerImplement struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	*DataBackupReconciler
}

// NewDataBackupReconcilerImplement returns a DataBackupReconcilerImplement
func NewDataBackupReconcilerImplement(client client.Client, log logr.Logger, recorder record.EventRecorder, databackupReconciler *DataBackupReconciler) *DataBackupReconcilerImplement {
	r := &DataBackupReconcilerImplement{
		Client:               client,
		Log:                  log,
		Recorder:             recorder,
		DataBackupReconciler: databackupReconciler,
	}
	return r
}

// ReconcileDataBackupDeletion reconciles the deletion of the DataBackup
func (r *DataBackupReconcilerImplement) ReconcileDataBackupDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataBackupDeletion")

	// 1. Delete release if exists
	releaseName := utils.GetDataBackupReleaseName(ctx.DataBackup.Name)
	err := helm.DeleteReleaseIfExists(releaseName, ctx.DataBackup.Namespace)
	if err != nil {
		log.Error(err, "can't delete release", "releaseName", releaseName)
		return utils.RequeueIfError(err)
	}

	// 2. Release lock on target dataset if necessary
	err = r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataBackup.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 3. remove finalizer
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

// reconcilePendingDataBackup reconciles DataBackups that are in `Pending` phase
func (r *DataBackupReconcilerImplement) reconcilePendingDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePendingDataBackup")
	targetDataset := ctx.Dataset
	// 1. Check if there's any Backuping pods(conflict DataBackup)
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
	if err != nil {
		log.Error(err, "Failed to get alluxio-master")
		return utils.RequeueIfError(err)
	}

	// 2. create backup Pod if not exist
	releaseName := utils.GetDataBackupReleaseName(ctx.DataBackup.Name)
	existed, err := helm.CheckRelease(releaseName, ctx.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseName, "namespace", ctx.Namespace)
		return utils.RequeueIfError(err)
	}
	// 2. install the helm chart if not exists and requeue
	if !existed {
		log.Info("DataBackup helm chart not installed yet, will install")
		valueFileName, err := r.generateDataBackupValueFile(ctx.DataBackup, masterPod)
		if err != nil {
			log.Error(err, "failed to generate databackup chart's value file")
			return utils.RequeueIfError(err)
		}
		chartName := utils.GetChartsDirectory() + "/" + cdatabackup.DATABACKUP_CHART
		err = helm.InstallRelease(releaseName, ctx.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install databackup chart")
			return utils.RequeueIfError(err)
		}
		log.Info("DataBackup helm chart successfullly installed", "namespace", ctx.Namespace, "releaseName", releaseName)

		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 3. Check running status of the DataBackup Pod
	backupPodName := utils.GetDataBackupPodName(ctx.DataBackup.Name)
	backupPod, err := kubeclient.GetPodByName(r.Client, backupPodName, ctx.Namespace)
	if err != nil {
		log.Error(err, "Failed to get databackup-pod")
		return utils.RequeueIfError(err)
	}
	if kubeclient.IsSucceededPod(backupPod) {
		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Phase = cdatabackup.PhaseComplete
		databackupToUpdate.Status.DurationTime = time.Since(databackupToUpdate.CreationTimestamp.Time).Round(time.Second).String()
		databackupToUpdate.Status.Conditions = []v1alpha1.DataBackupCondition{
			{
				Type:               cdatabackup.Complete,
				Status:             v1.ConditionTrue,
				Reason:             "BackupSuccessful",
				Message:            "Backup Pod exec successfully and finish",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}
		if err := r.Status().Update(context.TODO(), databackupToUpdate); err != nil {
			log.Error(err, "the backup pod has completd, but failed to update the databackup")
			return utils.RequeueIfError(err)
		}
		log.V(1).Info("Update phase of the databackup to Complete successfully")
		return utils.RequeueImmediately()
	} else if kubeclient.IsFailedPod(backupPod) {
		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Phase = cdatabackup.PhaseFailed
		databackupToUpdate.Status.DurationTime = time.Since(databackupToUpdate.CreationTimestamp.Time).Round(time.Second).String()
		databackupToUpdate.Status.Conditions = []v1alpha1.DataBackupCondition{
			{
				Type:               cdatabackup.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "BackupFailed",
				Message:            "Backup Pod exec failed and exit",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}
		if err := r.Status().Update(context.TODO(), databackupToUpdate); err != nil {
			log.Error(err, "the backup pod has failed, but failed to update the databackup")
			return utils.RequeueIfError(err)
		}
		log.V(1).Info("Update phase of the databackup to Failed successfully")
		return utils.RequeueImmediately()
	}
	return utils.RequeueAfterInterval(20 * time.Second)
}

// generateDataBackupValueFile builds a DataBackupValueFile by extracted specifications from the given DataBackup, and
// marshals the DataBackup to a temporary yaml file where stores values that'll be used by fluid dataBackup helm chart
func (r *DataBackupReconcilerImplement) generateDataBackupValueFile(databackup v1alpha1.DataBackup, masterPod *v1.Pod) (valueFileName string, err error) {
	nodeName, ip, rpcPort := utils.GetAddressOfMaster(masterPod)

	imageName, imageTag := docker.GetWorkerImage(r.Client, databackup.Spec.Dataset, "alluxio", databackup.Namespace)
	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	workdir := os.Getenv("FLUID_WORKDIR")
	if workdir == "" {
		workdir = "/tmp"
	}

	dataBackup := cdatabackup.DataBackup{
		Namespace: databackup.Namespace,
		Dataset:   databackup.Spec.Dataset,
		Name:      databackup.Name,
		NodeName:  nodeName,
		Image:     image,
		JavaEnv:   "-Dalluxio.master.hostname=" + ip + " -Dalluxio.master.rpc.port=" + strconv.Itoa(int(rpcPort)),
		Workdir:   workdir,
	}
	pvcName, path, err := utils.ParseBackupRestorePath(databackup.Spec.BackupPath)
	if err != nil {
		return
	}
	dataBackup.PVCName = pvcName
	dataBackup.Path = path

	dataBackupValue := cdatabackup.DataBackupValue{DataBackup: dataBackup}
	data, err := yaml.Marshal(dataBackupValue)
	if err != nil {
		return
	}

	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-backuper-values.yaml", databackup.Namespace, databackup.Name))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(valueFile.Name(), data, 0400)
	if err != nil {
		return
	}
	return valueFile.Name(), nil
}
