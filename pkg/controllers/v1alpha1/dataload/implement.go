package dataload

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	alluxioOperations "github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	jindoOperations "github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
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
func (r *DataLoadReconcilerImplement) ReconcileDataLoadDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataLoadDeletion")

	// 1. Delete release if exists
	releaseName := utils.GetDataLoadReleaseName(ctx.DataLoad.Name)
	err := helm.DeleteReleaseIfExists(releaseName, ctx.DataLoad.Namespace)
	if err != nil {
		log.Error(err, "can't delete release", "releaseName", releaseName)
		return utils.RequeueIfError(err)
	}

	// 2. Release lock on target dataset if necessary
	err = r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataLoad.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 3. remove finalizer
	if utils.HasDeletionTimestamp(ctx.DataLoad.ObjectMeta) {
		ctx.DataLoad.ObjectMeta.Finalizers = utils.RemoveString(ctx.DataLoad.ObjectMeta.Finalizers, cdataload.DATALOAD_FINALIZER)
		if err := r.Update(ctx, &ctx.DataLoad); err != nil {
			log.Error(err, "failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		log.Info("Finalizer is removed")
	}
	return utils.NoRequeue()
}

// ReconcileDataLoad reconciles the DataLoad according to its phase status
func (r *DataLoadReconcilerImplement) ReconcileDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataLoad")
	log.V(1).Info("process the cdataload", "cdataload", ctx.DataLoad)

	// DataLoad's phase transition: None -> Pending -> Executing -> Loaded or Failed
	switch ctx.DataLoad.Status.Phase {
	case common.PhaseNone:
		return r.reconcileNoneDataLoad(ctx)
	case common.PhasePending:
		return r.reconcilePendingDataLoad(ctx)
	case common.PhaseExecuting:
		return r.reconcileExecutingDataLoad(ctx)
	case common.PhaseComplete:
		return r.reconcileLoadedDataLoad(ctx)
	case common.PhaseFailed:
		return r.reconcileFailedDataLoad(ctx)
	default:
		log.Info("Unknown DataLoad phase, won't reconcile it")
	}
	return utils.NoRequeue()
}

// reconcileNoneDataLoad reconciles DataLoads that are in `DataLoadPhaseNone` phase
func (r *DataLoadReconcilerImplement) reconcileNoneDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNoneDataLoad")
	dataloadToUpdate := ctx.DataLoad.DeepCopy()
	dataloadToUpdate.Status.Phase = common.PhasePending
	if len(dataloadToUpdate.Status.Conditions) == 0 {
		dataloadToUpdate.Status.Conditions = []v1alpha1.Condition{}
	}
	dataloadToUpdate.Status.Duration = "Unfinished"
	if err := r.Status().Update(context.TODO(), dataloadToUpdate); err != nil {
		log.Error(err, "failed to update the cdataload")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("Update phase of the cdataload to Pending successfully")
	return utils.RequeueImmediately()
}

// reconcilePendingDataLoad reconciles DataLoads that are in `DataLoadPhasePending` phase
func (r *DataLoadReconcilerImplement) reconcilePendingDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePendingDataLoad")

	// 1. Check dataload namespace and dataset namespace need to be same
	if ctx.DataLoad.Namespace != ctx.DataLoad.Spec.Dataset.Namespace {
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeWarning,
			common.TargetDatasetNamespaceNotSame,
			"Dataload(%s) namespace is not same as dataset",
			ctx.DataLoad.Name)

		// Update DataLoad's phase to Failed, and no requeue
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			dataload, err := utils.GetDataLoad(r.Client, ctx.Name, ctx.Namespace)
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

	// 3. Check existence of the target dataset
	targetDataset, err := utils.GetDataset(r.Client, ctx.DataLoad.Spec.Dataset.Name, ctx.DataLoad.Spec.Dataset.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			log.Info("Target dataset not found", "targetDataset", ctx.DataLoad.Spec.Dataset)
			r.Recorder.Eventf(&ctx.DataLoad,
				v1.EventTypeWarning,
				common.TargetDatasetNotFound,
				"Target dataset(namespace: %s, name: %s) not found",
				ctx.DataLoad.Spec.Dataset.Namespace, ctx.DataLoad.Spec.Dataset.Name)
		} else {
			log.Error(err, "can't get target dataset", "targetDataset", ctx.DataLoad.Spec.Dataset)
		}
		return utils.RequeueAfterInterval(20 * time.Second)
	}
	log.V(1).Info("get target dataset", "targetDataset", targetDataset)

	// 2. Check if the target dataset has synced metadata
	//if targetDataset.Status.UfsTotal == "" || targetDataset.Status.UfsTotal == alluxio.METADATA_SYNC_NOT_DONE_MSG {
	//	log.V(1).Info("Target dataset not ready", "targetDataset", ctx.DataLoad.Spec.Dataset)
	//	r.Recorder.Eventf(&ctx.DataLoad,
	//		v1.EventTypeNormal,
	//		common.TargetDatasetNotReady,
	//		"Target dataset(namespace: %s, name: %s) metadata sync not done",
	//		targetDataset.Namespace, targetDataset.Name)
	//	return utils.RequeueAfterInterval(20 * time.Second)
	//}

	// 3. Check if there's any Executing DataLoad jobs(conflict DataLoad)
	conflictDataLoadRef := targetDataset.Status.DataLoadRef
	myDataLoadRef := utils.GetDataLoadRef(ctx.DataLoad.Name, ctx.DataLoad.Namespace)
	if len(conflictDataLoadRef) != 0 && conflictDataLoadRef != myDataLoadRef {
		log.V(1).Info("Found other DataLoads that is in Executing phase, will backoff", "other DataLoad", conflictDataLoadRef)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.DataLoadCollision,
			"Found other Dataload(%s) that is in Executing phase, will backoff",
			conflictDataLoadRef)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 4. Check if the bounded runtime is ready

	//runtimeConditions := targetDataset.Status.Conditions
	//ready := len(runtimeConditions) != 0 && runtimeConditions[len(runtimeConditions)-1].Status == v1.ConditionTrue

	var ready bool
	index, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)
	if index == -1 {
		log.Info("bounded runtime with Accelerate Category is not found on the target dataset", "targetDataset", targetDataset)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}
	switch boundedRuntime.Type {
	case common.ALLUXIO_RUNTIME:
		podName := fmt.Sprintf("%s-master-0", targetDataset.Name)
		containerName := "alluxio-master"
		fileUtils := alluxioOperations.NewAlluxioFileUtils(podName, containerName, targetDataset.Namespace, ctx.Log)
		ready = fileUtils.Ready()
	case common.JINDO_RUNTIME:
		podName := fmt.Sprintf("%s-jindofs-master-0", targetDataset.Name)
		containerName := "jindofs-master"
		fileUtils := jindoOperations.NewJindoFileUtils(podName, containerName, targetDataset.Namespace, ctx.Log)
		ready = fileUtils.Ready()
	default:
		log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "runtime", boundedRuntime)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not supported")
	}

	if !ready {
		log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataset)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready")
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 5. Check existence of the targetPath in alluxio
	var notExisted bool
	switch boundedRuntime.Type {
	case common.ALLUXIO_RUNTIME:
		podName := fmt.Sprintf("%s-master-0", targetDataset.Name)
		containerName := "alluxio-master"
		fileUtils := alluxioOperations.NewAlluxioFileUtils(podName, containerName, targetDataset.Namespace, ctx.Log)
		for _, target := range ctx.DataLoad.Spec.Target {
			isExist, err := fileUtils.IsExist(target.Path)
			if err != nil {
				return utils.RequeueAfterInterval(20 * time.Second)
			}
			if !isExist {
				notExisted = true
			}
		}
	case common.JINDO_RUNTIME:
		podName := fmt.Sprintf("%s-jindofs-master-0", targetDataset.Name)
		containerName := "jindofs-master"
		fileUtils := jindoOperations.NewJindoFileUtils(podName, containerName, targetDataset.Namespace, ctx.Log)
		for _, target := range ctx.DataLoad.Spec.Target {
			isExist, err := fileUtils.IsExist(target.Path)
			if err != nil {
				return utils.RequeueAfterInterval(20 * time.Second)
			}
			if !isExist {
				notExisted = true
			}
		}
	default:
		log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "runtime", boundedRuntime)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not supported")
	}

	if notExisted {
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeWarning,
			common.TargetDatasetPathNotFound,
			"Dataload target dataset's path is not existed")

		// Update DataLoad's phase to Failed, and no requeue
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			dataload, err := utils.GetDataLoad(r.Client, ctx.Name, ctx.Namespace)
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

	// 6. lock the target dataset. Make sure only one DataLoad can win the lock and
	// the losers have to requeue and go through the whole reconciliation loop.
	log.Info("No conflicts detected, try to lock the target dataset")
	datasetToUpdate := targetDataset.DeepCopy()
	datasetToUpdate.Status.DataLoadRef = myDataLoadRef
	if !reflect.DeepEqual(targetDataset.Status, datasetToUpdate.Status) {
		if err = r.Client.Status().Update(context.TODO(), datasetToUpdate); err != nil {
			log.V(1).Info("fail to get target dataset's lock, will requeue")
			//todo(xuzhihao): random backoff
			return utils.RequeueAfterInterval(20 * time.Second)
		}
	}

	// 7. update phase to Executing
	// We offload the helm install logic to `reconcileExecutingDataLoad` to
	// avoid such a case that status.phase change successfully first but helm install failed,
	// where the DataLoad job will never start and all other DataLoads will be blocked forever.
	log.Info("Get lock on target dataset, try to update phase")
	dataLoadToUpdate := ctx.DataLoad.DeepCopy()
	dataLoadToUpdate.Status.Phase = common.PhaseExecuting
	if err = r.Client.Status().Update(context.TODO(), dataLoadToUpdate); err != nil {
		log.Error(err, "failed to update cdataload's status to Executing, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update cdataload's status to Executing successfully")
	return utils.RequeueImmediately()
}

// reconcileExecutingDataLoad reconciles DataLoads that are in `Executing` phase
func (r *DataLoadReconcilerImplement) reconcileExecutingDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileExecutingDataLoad")

	// 1. Check if the helm release already exists
	releaseName := utils.GetDataLoadReleaseName(ctx.Name)
	jobName := utils.GetDataLoadJobName(releaseName)
	existed, err := helm.CheckRelease(releaseName, ctx.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseName, "namespace", ctx.Namespace)
		return utils.RequeueIfError(err)
	}

	// 2. install the helm chart if not exists and requeue
	if !existed {
		log.Info("DataLoad job helm chart not installed yet, will install")

		targetDataset, err := utils.GetDataset(r.Client, ctx.DataLoad.Spec.Dataset.Name, ctx.DataLoad.Spec.Dataset.Namespace)
		if err != nil {
			log.Error(err, "targetDataset not exists", "targetDataset", releaseName, "namespace", ctx.DataLoad.Spec.Dataset.Name)
		}

		_, boundedRuntime := utils.GetRuntimeByCategory(targetDataset.Status.Runtimes, common.AccelerateCategory)

		chartName := ""
		valueFileName := ""
		switch boundedRuntime.Type {
		case common.ALLUXIO_RUNTIME:
			valueFileName, err = r.generateDataLoadValueFile(ctx.DataLoad)
			if err != nil {
				log.Error(err, "failed to generate dataload chart's value file")
				return utils.RequeueIfError(err)
			}
			chartName = utils.GetChartsDirectory() + "/" + cdataload.DATALOAD_CHART + "/" + common.ALLUXIO_RUNTIME
		case common.JINDO_RUNTIME:
			valueFileName, err = r.generateJindoDataLoadValueFile(ctx.DataLoad)
			if err != nil {
				log.Error(err, "failed to generate dataload chart's value file")
				return utils.RequeueIfError(err)
			}
			chartName = utils.GetChartsDirectory() + "/" + cdataload.DATALOAD_CHART + "/" + common.JINDO_RUNTIME
		default:
			log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "runtime", boundedRuntime)
			r.Recorder.Eventf(&ctx.DataLoad,
				v1.EventTypeNormal,
				common.RuntimeNotReady,
				"Bounded accelerate runtime not supported")
		}
		err = helm.InstallRelease(releaseName, ctx.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install dataload chart")
			return utils.RequeueIfError(err)
		}
		log.Info("DataLoad job helm chart successfullly installed", "namespace", ctx.Namespace, "releaseName", releaseName)
		r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeNormal, common.DataLoadJobStarted, "The DataLoad job %s started", jobName)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 3. Check running status of the DataLoad job
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
				dataload, err := utils.GetDataLoad(r.Client, ctx.Name, ctx.Namespace)
				if err != nil {
					return err
				}
				dataloadToUpdate := dataload.DeepCopy()
				jobCondition := job.Status.Conditions[0]
				dataloadToUpdate.Status.Conditions = []v1alpha1.Condition{
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
func (r *DataLoadReconcilerImplement) reconcileLoadedDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileLoadedDataLoad")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataLoad.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. record event and no requeue
	log.Info("DataLoad Loaded, no need to requeue")
	jobName := utils.GetDataLoadJobName(utils.GetDataLoadReleaseName(ctx.Name))
	r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeNormal, common.DataLoadJobComplete, "DataLoad job %s succeeded", jobName)
	return utils.NoRequeue()
}

// reconcileFailedDataLoad reconciles DataLoads that are in `DataLoadPhaseComplete` phase
func (r *DataLoadReconcilerImplement) reconcileFailedDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileFailedDataLoad")
	// 1. release lock on target dataset
	err := r.releaseLockOnTargetDataset(ctx, log)
	if err != nil {
		log.Error(err, "can't release lock on target dataset", "targetDataset", ctx.DataLoad.Spec.Dataset)
		return utils.RequeueIfError(err)
	}

	// 2. record event and no requeue
	log.Info("DataLoad failed, won't requeue")
	jobName := utils.GetDataLoadJobName(utils.GetDataLoadReleaseName(ctx.Name))
	r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeNormal, common.DataLoadJobFailed, "DataLoad job %s failed", jobName)
	return utils.NoRequeue()
}

// generateDataLoadValueFile builds a DataLoadValue by extracted specifications from the given DataLoad, and
// marshals the DataLoadValue to a temporary yaml file where stores values that'll be used by fluid dataloader helm chart
func (r *DataLoadReconcilerImplement) generateDataLoadValueFile(dataload v1alpha1.DataLoad) (valueFileName string, err error) {
	targetDataset, err := utils.GetDataset(r.Client, dataload.Spec.Dataset.Name, dataload.Spec.Dataset.Namespace)
	if err != nil {
		return "", err
	}

	imageName, imageTag := docker.GetWorkerImage(r.Client, dataload.Spec.Dataset.Name, "alluxio", dataload.Spec.Dataset.Namespace)
	image := fmt.Sprintf("%s:%s", imageName, imageTag)

	dataloadInfo := cdataload.DataLoadInfo{
		BackoffLimit:  3,
		TargetDataset: dataload.Spec.Dataset.Name,
		LoadMetadata:  dataload.Spec.LoadMetadata,
		Image:         image,
	}

	targetPaths := []cdataload.TargetPath{}
	for _, target := range dataload.Spec.Target {
		fluidNative := isTargetPathUnderFluidNativeMounts(target.Path, *targetDataset)
		targetPaths = append(targetPaths, cdataload.TargetPath{
			Path:        target.Path,
			Replicas:    target.Replicas,
			FluidNative: fluidNative,
		})
	}
	dataloadInfo.TargetPaths = targetPaths
	dataLoadValue := cdataload.DataLoadValue{DataLoadInfo: dataloadInfo}
	data, err := yaml.Marshal(dataLoadValue)
	if err != nil {
		return
	}

	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-loader-values.yaml", dataload.Namespace, dataload.Name))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(valueFile.Name(), data, 0400)
	if err != nil {
		return
	}
	return valueFile.Name(), nil
}

// isTargetPathUnderFluidNativeMounts checks if targetPath is a subpath under some given native mount point.
// We check this for the reason that native mount points need extra metadata sync alluxioOperations.
func isTargetPathUnderFluidNativeMounts(targetPath string, dataset v1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {
		mountPointOnDDCEngine := fmt.Sprintf("/%s", mount.Name)
		if mount.Path != "" {
			mountPointOnDDCEngine = mount.Path
		}

		//todo(xuzhihao): HasPrefix is not enough.
		if strings.HasPrefix(targetPath, mountPointOnDDCEngine) &&
			(strings.HasPrefix(mount.MountPoint, common.PathScheme.String()) || strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String())) {
			return true
		}
	}
	return false
}

// releaseLockOnTargetDataset releases lock on target dataset if the lock currently belongs to reconciling DataLoad.
// We use a key-value pair on the target dataset's status as the lock. To release the lock, we can simply set the value to empty.
func (r *DataLoadReconcilerImplement) releaseLockOnTargetDataset(ctx reconcileRequestContext, log logr.Logger) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(r.Client, ctx.DataLoad.Spec.Dataset.Name, ctx.DataLoad.Spec.Dataset.Namespace)
		if err != nil {
			if utils.IgnoreNotFound(err) == nil {
				log.Info("can't find target dataset, won't release lock", "targetDataset", ctx.DataLoad.Spec.Dataset.Name)
				return nil
			}
			// other error
			return err
		}
		if dataset.Status.DataLoadRef != utils.GetDataLoadRef(ctx.DataLoad.Name, ctx.DataLoad.Namespace) {
			log.Info("Found DataLoadRef inconsistent with the reconciling DataLoad, won't release this lock, ignore it", "DataLoadRef", dataset.Status.DataLoadRef)
			return nil
		}
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.Status.DataLoadRef = ""
		if !reflect.DeepEqual(datasetToUpdate.Status, dataset) {
			if err := r.Status().Update(ctx, datasetToUpdate); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
