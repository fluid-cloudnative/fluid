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
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		log.V(1).Info("Found other DataBackups that is in Executing phase, will backoff", "other DataBackup", conflictDataBackupRef)

		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Conditions = []v1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "conflictDataBackupRef",
				Message:            "Found other DataBackup that is in Executing phase",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}
		databackupToUpdate.Status.Phase = common.PhaseFailed

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
	if !strings.HasPrefix(ctx.DataBackup.Spec.BackupPath, common.PathScheme.String()) && !strings.HasPrefix(ctx.DataBackup.Spec.BackupPath, common.VolumeScheme.String()) {
		log.Error(fmt.Errorf("PathNotSupported"), "don't support path in this form", "path", ctx.DataBackup.Spec.BackupPath)
		databackupToUpdate := ctx.DataBackup.DeepCopy()
		databackupToUpdate.Status.Conditions = []v1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "PathNotSupported",
				Message:            "Only support pvc and local path now",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}
		databackupToUpdate.Status.Phase = common.PhaseFailed

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
	// 4. update phase to Executing
	log.Info("Get lock on target dataset, try to update phase")
	dataBackupToUpdate := ctx.DataBackup.DeepCopy()
	dataBackupToUpdate.Status.Phase = common.PhaseExecuting
	if err := r.Client.Status().Update(context.TODO(), dataBackupToUpdate); err != nil {
		log.Error(err, "failed to update cdatabackup's status to Executing, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update cdatabackup's status to Executing successfully")
	return utils.RequeueImmediately()
}

// reconcileExecutingDataBackup reconciles DataBackups that are in `Executing` phase
func (r *DataBackupReconcilerImplement) reconcileExecutingDataBackup(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileExecutingDataBackup")
	// 1. get the alluxio-master Pod
	// For HA Mode, need find the leading master pod name
	targetDataset := ctx.Dataset
	index, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)
	if index == -1 {
		log.Info("bounded runtime with Accelerate Category is not found on the target dataset", "targetDataset", targetDataset)
	}

	var podName string
	if boundedRuntime.MasterReplicas <= 1 {
		podName = ctx.Dataset.Name + "-master-0"
	} else {
		var err error
		switch boundedRuntime.Type {
		case common.ALLUXIO_RUNTIME:
			execPodName := fmt.Sprintf("%s-master-0", targetDataset.Name)
			containerName := "alluxio-master"
			fileUtils := operations.NewAlluxioFileUtils(execPodName, containerName, targetDataset.Namespace, ctx.Log)
			podName, err = fileUtils.MasterPodName()
		default:
			log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "runtime", boundedRuntime)
			r.Recorder.Eventf(&ctx.DataBackup,
				v1.EventTypeNormal,
				common.RuntimeNotReady,
				"Bounded accelerate runtime not supported")
		}
		if err != nil {
			log.Error(err, "failed to get master pod name, will retry")
			return utils.RequeueIfError(err)
		}
	}

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
		valueFileName, err := r.generateDataBackupValueFile(ctx, masterPod)
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
		databackupToUpdate.Status.Phase = common.PhaseComplete
		var successTime time.Time
		if len(backupPod.Status.Conditions) != 0 {
			successTime = backupPod.Status.Conditions[0].LastTransitionTime.Time
		} else {
			// fail to get successTime, use current time as default
			successTime = time.Now()
		}
		databackupToUpdate.Status.Duration = utils.CalculateDuration(databackupToUpdate.CreationTimestamp.Time, successTime)
		databackupToUpdate.Status.Conditions = []v1alpha1.Condition{
			{
				Type:               common.Complete,
				Status:             v1.ConditionTrue,
				Reason:             "BackupSuccessful",
				Message:            "Backup Pod exec successfully and finish",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(successTime),
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
		databackupToUpdate.Status.Phase = common.PhaseFailed
		var failedTime time.Time
		if len(backupPod.Status.Conditions) != 0 {
			failedTime = backupPod.Status.Conditions[0].LastTransitionTime.Time
		} else {
			// fail to get successTime, use current time as default
			failedTime = time.Now()
		}
		databackupToUpdate.Status.Duration = utils.CalculateDuration(databackupToUpdate.CreationTimestamp.Time, failedTime)
		databackupToUpdate.Status.Conditions = []v1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "BackupFailed",
				Message:            "Backup Pod exec failed and exit",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(failedTime),
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
func (r *DataBackupReconcilerImplement) generateDataBackupValueFile(ctx reconcileRequestContext, masterPod *v1.Pod) (valueFileName string, err error) {
	databackup := ctx.DataBackup
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

	dataBackupValue.InitUsers = common.InitUsers{
		Enabled: false,
	}

	var runtime v1alpha1.AlluxioRuntime
	var runAs *v1alpha1.User
	initUsers := common.ImageInfo{}

	// get the runAs and initUsers imageInfo from runtime
	err = r.Get(ctx, types.NamespacedName{
		Namespace: databackup.Namespace,
		Name:      databackup.Spec.Dataset,
	}, &runtime)
	if err == nil {
		runAs = runtime.Spec.RunAs
		initUsers = common.ImageInfo{
			Image:           runtime.Spec.InitUsers.Image,
			ImageTag:        runtime.Spec.InitUsers.ImageTag,
			ImagePullPolicy: runtime.Spec.InitUsers.ImagePullPolicy,
		}
	}
	// databackup.Spec.RunAs > runtime.Spec.RunAs > root
	if databackup.Spec.RunAs != nil {
		runAs = databackup.Spec.RunAs
	}

	if runAs != nil {
		dataBackupValue.UserInfo.User = int(*runAs.UID)
		dataBackupValue.UserInfo.Group = int(*runAs.GID)
		dataBackupValue.UserInfo.FSGroup = 0
		dataBackupValue.InitUsers = common.InitUsers{
			Enabled:  true,
			EnvUsers: utils.GetInitUserEnv(runAs),
			Dir:      utils.GetBackupUserDir(dataBackup.Namespace, dataBackup.Name),
		}
	}

	dataBackupValue.InitUsers.Image, dataBackupValue.InitUsers.ImageTag, dataBackupValue.InitUsers.ImagePullPolicy = docker.GetInitUserImage(initUsers)

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
