package dataload

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReconcilerImplement struct {
	client.Client
	Log             logr.Logger
	Recorder        record.EventRecorder
	DataLoaderImage string
}

// Return a new reconciler implement
func NewReconcilerImplement(client client.Client, log logr.Logger, recorder record.EventRecorder) *ReconcilerImplement {
	r := &ReconcilerImplement{
		Client:   client,
		Log:      log,
		Recorder: recorder,
	}
	return r
}

// Set up the reconciler
func (r *ReconcilerImplement) Setup() {
	if r.DataLoaderImage != "" {
		return
	}

	val, existed := os.LookupEnv(common.ENV_DATALOADER_IMG)
	if !existed {
		r.DataLoaderImage = common.DATALOAD_DEFAULT_IMAGE
	} else {
		r.DataLoaderImage = val
	}
	r.Log.Info("Setting up DataLoader Image", "Image", r.DataLoaderImage)
}

// Reconcile Dataload
func (r *ReconcilerImplement) ReconcileDataload(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	/*
		DataLoadPhase: Complete
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseComplete {
		return r.reconcileCompleteDataload(ctx)
	}

	/*
		DataLoadPhase: Failed
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseFailed {
		return r.reconcileFailedDataload(ctx)
	}

	/*
		DataLoadPhase: None -> Pending
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseNone {
		return r.reconcileNoneDataload(ctx)
	}

	/*
		DataLoadPhase: Pending -> Loading
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhasePending {
		return r.reconcilePendingDataload(ctx)
	}

	/*
		DataLoadPhase: Loading -> Complete/Failed
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseLoading {
		return r.reconcileLoadingDataload(ctx)
	}

	//unreachable in theory
	return ctrl.Result{}, nil
}

// Reconcile DataLoad deletion
func (r *ReconcilerImplement) ReconcileDataloadDeletion(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	/*
		Delete helm release if exists
	*/
	//datasetName := ctx.DataLoad.Spec.DatasetName
	//releaseName := common.GetReleaseName(datasetName)
	releaseName, existed := ctx.DataLoad.Labels["release"]
	if existed {
		existed, err := helm.CheckRelease(releaseName, ctx.DataLoad.Namespace)
		if err != nil {
			ctx.Log.Error(err, "Helm check release error")
		} else if existed {
			if err := helm.DeleteRelease(releaseName, ctx.DataLoad.Namespace); err != nil {
				ctx.Log.Error(err, "Helm can't uninstall release")
			}
			ctx.Log.Info("Helm release successfully deleted", "releaseName", releaseName)
		} else {
			ctx.Log.Info("Related Helm release not found, just delete dataload object")
		}
	}
	/*
		Remove finalizer
	*/
	ctx.Log.Info("Start to clean up finalizer", "dataload", ctx.DataLoad)
	if !ctx.DataLoad.ObjectMeta.GetDeletionTimestamp().IsZero() {
		ctx.DataLoad.ObjectMeta.Finalizers = utils.RemoveString(ctx.DataLoad.ObjectMeta.Finalizers, common.DATALOAD_FINALIZER)
		if err := r.Update(ctx, &ctx.DataLoad); err != nil {
			ctx.Log.Error(err, "Failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		ctx.Log.Info("Finalizer is removed", "dataload", ctx.DataLoad)
	}

	return utils.NoRequeue()
}

// Reconcile Dataload in `None` phase
func (r *ReconcilerImplement) reconcileNoneDataload(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	dataloadToUpdate := ctx.DataLoad.DeepCopy()
	dataloadToUpdate.Status.Phase = common.DataloadPhasePending
	if len(dataloadToUpdate.Status.Conditions) == 0 {
		dataloadToUpdate.Status.Conditions = []datav1alpha1.DataloadCondition{}
	}
	if err := r.Status().Update(ctx, dataloadToUpdate); err != nil {
		ctx.Log.Error(err, "Failed to update the dataload")
		return utils.RequeueIfError(err)
	} else {
		ctx.Log.V(1).Info("Update the status of the dataload successfully", "phase", ctx.DataLoad.Status.Phase)
	}
	// to sync ctx.Dataload for latest status update
	return utils.RequeueImmediately()
}

// Reconcile Dataload in `Pending` phase
func (r *ReconcilerImplement) reconcilePendingDataload(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	/*
		1. Check if related dataset ready (i.e. related dataset exists and bounded)
	*/
	dataset, ready, err := r.checkRelatedDatasetReady(ctx)
	if err != nil {
		ctx.Log.Error(err, "Failed to check related dataset status")
		return utils.RequeueIfError(err)
	} else if !ready {
		if dataset == nil {
			ctx.Log.Info("Related dataset not found", "datasetName", ctx.DataLoad.Spec.DatasetName, "namespace", ctx.DataLoad.Namespace)
			r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.DatasetNotReady, "Can't find dataset \"%s\"", ctx.DataLoad.Spec.DatasetName)
		} else {
			ctx.Log.Info("Related dataset not ready", "datasetName", dataset.Name, "namespace", dataset.Namespace)
			r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.DatasetNotReady, "Dataset \"%s\" is not bound", ctx.DataLoad.Spec.DatasetName)
		}
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}
	ctx.Log.Info("Related dataset ready", "datasetName", dataset.Name, "namespace", dataset.Namespace)

	/*
		2. Check if related AlluxioRuntime ready (i.e. related alluxioRuntime exists and setup done)
	*/
	alluxioRuntime, ready, err := r.checkRelatedRuntimeReady(ctx)
	if err != nil {
		ctx.Log.Error(err, "Failed to check related runtime status")
		return utils.RequeueIfError(err)
	} else if !ready {
		if alluxioRuntime == nil {
			ctx.Log.Info("Related alluxio runtime not found", "runtimeName", ctx.DataLoad.Spec.DatasetName, "namespace", ctx.DataLoad.Namespace)
			r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.RuntimeNotReady, "Cant't find Alluxio runtime named \"%s\"", ctx.DataLoad.Spec.DatasetName)
		} else {
			ctx.Log.Info("Related alluxio runtime not ready", "runtimeName", alluxioRuntime.Name, "namespace", alluxioRuntime.Namespace, "scheduledWorkers", alluxioRuntime.Status.CurrentWorkerNumberScheduled, "availableWorkers", alluxioRuntime.Status.WorkerNumberAvailable)
			r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.RuntimeNotReady, "Alluxio runtime is not ready")
		}
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}
	ctx.Log.Info("Related dataset ready", "runtimeName", alluxioRuntime.Name, "namespace", alluxioRuntime.Namespace)

	/*
		3. Check if there is any dataload with collision. Backoff if there exists
	*/
	dataloadWithCollision, err := r.findDataLoadWithCollision(ctx)
	if err != nil {
		ctx.Log.Error(err, "Failed to find other dataload")
		return utils.RequeueIfError(err)
	}

	if dataloadWithCollision != nil {
		ctx.Log.Info("Another Dataload is loading", "dataloadName", dataloadWithCollision.Name, "namespace", dataloadWithCollision.Namespace)
		r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.DataLoadCollision, "Another DataLoad(name: %s) is running a prefetch job", dataloadWithCollision.Name)
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}
	ctx.Log.Info("No other dataload loading, ready to prefetch...")

	/*
		4. Launch Prefetch Job
	*/
	installDone, err := r.launchPrefetchJob(ctx, alluxioRuntime.Status.WorkerNumberAvailable)
	if err != nil {
		return utils.RequeueIfError(err)
	} else if !installDone {
		ctx.Log.Info("Some other prefetch job running")
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}

	/*
		5. update dataload phase
		We have to make sure there is only one DataLoad object with phase `Loading` for the same datasetName at any time,
		which means the phase of the Dataload MUST be in consistency with a installed job.
		We retry to do updating, if fail, delete associated helm release and turn back to phase `Pending`.
	*/
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		dataload, err := utils.GetDataLoad(r, ctx.DataLoad.Name, ctx.Namespace)
		if err != nil {
			return err
		}
		dataloadToUpdate := dataload.DeepCopy()
		dataloadToUpdate.Status.Phase = common.DataloadPhaseLoading
		if !reflect.DeepEqual(dataload.Status, dataloadToUpdate.Status) {
			return r.Status().Update(ctx, dataloadToUpdate)
		}
		return nil
	})

	if err != nil {
		ctx.Log.Error(err, "Unable to update dataload's phase to `Loading`. Trying to uninstall related helm release")
		//releaseName := common.GetReleaseName(ctx.DataLoad.Spec.DatasetName)
		dataload, err := utils.GetDataLoad(r, ctx.DataLoad.Name, ctx.DataLoad.Namespace)
		if err == nil {
			releaseName, existed := dataload.Labels["release"]
			if existed {
				helmErr := helm.DeleteReleaseIfExists(releaseName, ctx.DataLoad.Namespace)
				if helmErr != nil {
					ctx.Log.Error(helmErr, "Unable to delete related release")
				}
			}
		} else {
			return utils.RequeueIfError(err)
		}

	} else {
		// Helm chart installed and phase updated with no error
		r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeNormal, common.PrefetchJobStarted, "Started Prefetch Job")
	}

	//dataloadToUpdate := ctx.DataLoad.DeepCopy()
	//dataloadToUpdate.Status.Phase = common.DataloadPhaseLoading
	//if err := r.Status().Update(ctx, dataloadToUpdate); err != nil {
	//	ctx.Log.Error(err, "Fail to update dataload phase", "Status update error", ctx)
	//	return utils.RequeueIfError(err)
	//}
	return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
}

// Reconcile Dataload in `Loading` phase
func (r *ReconcilerImplement) reconcileLoadingDataload(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	dataloadToUpdate := ctx.DataLoad.DeepCopy()
	var needUpdate bool

	job := r.getRelatedJobIfNoInterrupts(ctx)

	if job == nil {
		// Some interruption occurs
		releaseName, existed := ctx.DataLoad.Labels["release"]
		if existed {
			_ = helm.DeleteReleaseIfExists(releaseName, ctx.DataLoad.Namespace)
		}
		dataloadToUpdate.Status.Phase = common.DataloadPhaseNone
		dataloadToUpdate.Status.Conditions = []datav1alpha1.DataloadCondition{}
		ctx.Log.Info("Dataload's Loading process has been interrupted")
		r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.PrefetchJobInterrupted, "Dataload's Loading process has been interrupted")
		needUpdate = true
	} else if job.Status.Failed > 0 || job.Status.Succeeded == *job.Spec.Completions {
		dataloadToUpdate.Status.Conditions = []datav1alpha1.DataloadCondition{
			{
				Type:               common.DataloadConditionType(job.Status.Conditions[0].Type),
				Status:             job.Status.Conditions[0].Status,
				Reason:             job.Status.Conditions[0].Reason,
				Message:            job.Status.Conditions[0].Message,
				LastProbeTime:      job.Status.Conditions[0].LastProbeTime,
				LastTransitionTime: job.Status.Conditions[0].LastTransitionTime,
			},
		}

		if job.Status.Failed > 0 {
			ctx.Log.Info("Related Job failed", "jobName", job.Name)
			dataloadToUpdate.Status.Phase = common.DataloadPhaseFailed
			r.Recorder.Eventf(dataloadToUpdate, v1.EventTypeNormal, common.PrefetchJobFailed, "Job \"%s\" Finished. Current status: Failed", job.Name)
		} else {
			ctx.Log.Info("Related Job complete", "jobName", job.Name)
			dataloadToUpdate.Status.Phase = common.DataloadPhaseComplete
			r.Recorder.Eventf(dataloadToUpdate, v1.EventTypeNormal, common.PrefetchJobComplete, "Job \"%s\" Finished. Current status: Complete", job.Name)
		}
		needUpdate = true
	}

	if needUpdate {
		if !reflect.DeepEqual(dataloadToUpdate.Status, ctx.DataLoad.Status) {
			if err := r.Status().Update(ctx, dataloadToUpdate); err != nil {
				ctx.Log.Error(err, "Failed to update status of a loading dataload", "dataloadName", ctx.DataLoad.Name)
				return utils.RequeueIfError(err)
			}
		}
	}

	return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
}

// Reconcile Dataload in `Complete` phase
func (r *ReconcilerImplement) reconcileCompleteDataload(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	//No need to reconcile
	ctx.Log.Info("Dataload prefetch job done", "dataloadName", ctx.DataLoad.Name, "currentPhase", ctx.DataLoad.Status.Phase)
	return utils.NoRequeue()
}

// Reconcile Dataload in `Failed` phase
func (r *ReconcilerImplement) reconcileFailedDataload(ctx cdataload.ReconcileRequestContext) (ctrl.Result, error) {
	//No need to reconcile
	ctx.Log.Info("Dataload prefetch job done", "dataloadName", ctx.DataLoad.Name, "currentPhase", ctx.DataLoad.Status.Phase)
	return utils.NoRequeue()
}

/*
	Check existence and status of the runtime related to the dataload.
	If the runtime exists, return a pointer to that runtime.
	If the runtime is ready, which means all its scheduled workers are available, `ready`= true, otherwise false.
*/
func (r *ReconcilerImplement) checkRelatedRuntimeReady(ctx cdataload.ReconcileRequestContext) (alluxioRuntime *datav1alpha1.AlluxioRuntime, ready bool, err error) {
	runtimeName := ctx.DataLoad.Spec.DatasetName
	alluxioRuntime, err = utils.GetAlluxioRuntime(r, runtimeName, ctx.DataLoad.Namespace)
	if err != nil || alluxioRuntime == nil {
		// Ignore not found error and take it as not ready
		err = utils.IgnoreNotFound(err)
		return nil, false, err
	}

	// Check if all scheduled workers have been ready
	if alluxioRuntime.Status.WorkerNumberAvailable == alluxioRuntime.Status.CurrentWorkerNumberScheduled &&
		alluxioRuntime.Status.FuseNumberAvailable == alluxioRuntime.Status.CurrentFuseNumberScheduled &&
		alluxioRuntime.Status.WorkerNumberAvailable != 0 && alluxioRuntime.Status.FuseNumberAvailable != 0 {

		ready = true
	}
	return
}

/*
	Check existence and status of the dataset related to the dataload.
	If the dataset exists, return a pointer to that dataset.
	If the dataset has a phase `Bound`, `ready`=true, otherwise false.
*/
func (r *ReconcilerImplement) checkRelatedDatasetReady(ctx cdataload.ReconcileRequestContext) (dataset *datav1alpha1.Dataset, ready bool, err error) {
	datasetName := ctx.DataLoad.Spec.DatasetName
	dataset, err = utils.GetDataset(r, datasetName, ctx.DataLoad.Namespace)
	if err != nil || dataset == nil {
		// Ignore not found error and take it as not ready
		err = utils.IgnoreNotFound(err)
		return nil, false, err
	}
	return dataset, dataset.Status.Phase == datav1alpha1.BoundDatasetPhase, nil
}

/*
	Check out if there's any dataload collides with the one reconciling. Return a pointer to it if exists.
	A dataload collides with another one means that they have the same `datasetName` and some of them in phase `Loading`
*/
func (r *ReconcilerImplement) findDataLoadWithCollision(ctx cdataload.ReconcileRequestContext) (dataLoad *datav1alpha1.AlluxioDataLoad, err error) {
	collisionFunc := func(dl datav1alpha1.AlluxioDataLoad) bool {
		// A dataload with collision is another dataload object with same DatasetName and loading phase
		return dl.Name != ctx.DataLoad.Name && dl.Spec.DatasetName == ctx.DataLoad.Spec.DatasetName && dl.Status.Phase == common.DataloadPhaseLoading
	}
	//TODO: Using label Selector instead of using predicate
	return utils.FindDataLoadWithPredicate(r, ctx.DataLoad.Namespace, collisionFunc)
}

/*
	Launch a batch job to do prefetching
*/
func (r *ReconcilerImplement) launchPrefetchJob(ctx cdataload.ReconcileRequestContext, numWorker int32) (done bool, err error) {
	/*
		1. Check Helm releases
	*/
	//releaseName := fmt.Sprintf("%s-load", ctx.DataLoad.Spec.DatasetName)
	releaseName := utils.NewReleaseName(ctx.DataLoad.Spec.DatasetName)
	existed, err := helm.CheckRelease(releaseName, ctx.DataLoad.Namespace)
	if err != nil {
		ctx.Log.Error(err, "Fail to check helm releases")
		return false, err
	}
	if existed {
		ctx.Log.Info("A helm release with same name and namespace has already existed", "releaseName", releaseName, "namespace", ctx.DataLoad.Namespace)
		return false, nil
	}

	ctx.Log.Info("Check Helm releases: No conflicts found")

	/*
		2. generate value file
	*/
	valueFileName, err := r.generateValueFile(ctx.DataLoad, numWorker)
	if err != nil {
		return false, err
	}

	var chartName = utils.GetChartsDirectory() + "/" + common.DATALOAD_CHART

	/*
		3. Label dataload with release name
	*/
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		dataload, err := utils.GetDataLoad(r, ctx.DataLoad.Name, ctx.DataLoad.Namespace)
		if err != nil {
			return err
		}
		dataloadToUpdate := dataload.DeepCopy()
		if dataloadToUpdate.Labels == nil {
			dataloadToUpdate.Labels = map[string]string{}
		}
		dataloadToUpdate.Labels["release"] = releaseName
		return r.Update(ctx, dataloadToUpdate)
	})

	if err != nil {
		return false, err
	}

	/*
		4. Install release with generated value file
	*/
	if err := helm.InstallRelease(releaseName, ctx.DataLoad.Namespace, valueFileName, chartName); err != nil {
		ctx.Log.Error(err, "Fail to install helm chart")
		r.Recorder.Eventf(&ctx.DataLoad, v1.EventTypeWarning, common.ErrorHelmInstall, "Can't install Helm release due to %v", err)
		return false, err
	}

	ctx.Log.Info("Helm chart installed", "releaseName", releaseName)
	return true, nil
}

/*
	Generate value file used to config batch job
*/
func (r *ReconcilerImplement) generateValueFile(dataload datav1alpha1.AlluxioDataLoad, numWorker int32) (valueFileName string, err error) {
	dataloadConfig := cdataload.DataLoadInfo{
		BackoffLimit: 6,
		Image:        r.DataLoaderImage,
		NumWorker:    numWorker,
	}

	if dataload.Spec.SlotsPerNode != nil {
		dataloadConfig.Threads = *dataload.Spec.SlotsPerNode
	}

	if dataload.Spec.Path != "" {
		dataloadConfig.MountPath = dataload.Spec.Path
	}

	value := &cdataload.DataLoadValue{DataLoadInfo: dataloadConfig}
	valueFileName = ""
	data, err := yaml.Marshal(value)
	if err != nil {
		return
	}

	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-load-values.yaml", dataload.Namespace, dataload.Name))
	if err != nil {
		return
	}
	err = ioutil.WriteFile(valueFile.Name(), data, 0400)
	if err != nil {
		return
	}
	return valueFile.Name(), nil
}

/*
	If there's any change of other related resources that interrupted loading process, return nil.
	Return a pointer to the related job otherwise.
	An interruption can be something like related runtime or dataset becomes not ready, related job not found, etc.

*/
func (r *ReconcilerImplement) getRelatedJobIfNoInterrupts(ctx cdataload.ReconcileRequestContext) *batchv1.Job {
	dataset, ready, err := r.checkRelatedDatasetReady(ctx)
	if err != nil || dataset == nil || !ready {
		return nil
	}

	runtime, ready, err := r.checkRelatedRuntimeReady(ctx)
	if err != nil || runtime == nil || !ready {
		return nil
	}

	releaseName, existed := ctx.DataLoad.Labels["release"]
	if !existed {
		return nil
	}

	key := types.NamespacedName{
		Namespace: ctx.DataLoad.Namespace,
		Name:      utils.GetJobNameFromReleaseName(releaseName),
	}
	job := &batchv1.Job{}
	if err := r.Get(ctx, key, job); err != nil {
		return nil
	}

	return job
}
