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

package alluxio

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
	"github.com/cloudnativefluid/fluid/pkg/utils"
	"github.com/cloudnativefluid/fluid/pkg/utils/helm"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	finalizer = "fluid-dataload-controller-finalizer"
)

type DataLoadValue struct {
	DataLoadInfo DataLoadInfo `yaml:"dataloader,omitempty"`
}

type DataLoadInfo struct {
	BackoffLimit int32  `yaml:"backoffLimit,omitempty"`
	Threads      int32  `yaml:"threads,omitempty"`
	MountPath    string `yaml:"mountPath,omitempty"`
	Image        string `yaml:"image"`
	NumWorker    int32  `yaml:"numWorker,omitempty"`
}

// DataLoadReconciler reconciles a AlluxioDataLoad object
type DataLoadReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type reconcileRequestContext struct {
	context.Context
	Log      logr.Logger
	DataLoad datav1alpha1.AlluxioDataLoad
	types.NamespacedName
}

//Reconcile reconciles the AlluxioDataLoad Object
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads/status,verbs=get;update;patch

func (r *DataLoadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := reconcileRequestContext{
		Context:        context.Background(),
		Log:            r.Log.WithValues("alluxiodataload", req.NamespacedName),
		NamespacedName: req.NamespacedName,
	}

	ctx.Log.V(1).Info("Reconciling dataload request", "request", req)

	/*
		1. Load the DataLoad resource object
	*/
	if err := r.Get(ctx, req.NamespacedName, &ctx.DataLoad); err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Dataload not found", "dataload", ctx.NamespacedName)
			return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get dataload info")
			return utils.RequeueIfError(errors.Wrap(err, "Failed to get dataload info"))
		}
	}
	ctx.Log.Info("Dataload found.", "dataload", ctx.DataLoad)

	/*
		2. delete dataload if necessary
	*/
	if utils.HasDeletionTimestamp(ctx.DataLoad.ObjectMeta) {
		return r.ReconcileDataloadDeletion(ctx)
	}

	/*
		3. Add finalizer
	*/
	if !utils.ContainsString(ctx.DataLoad.ObjectMeta.GetFinalizers(), finalizer) {
		return r.addFinalizerAndRequeue(ctx)
	}

	/*
		4. Do dataload reconciling
	*/
	return r.ReconcileDataload(ctx)
}

func (r *DataLoadReconciler) ReconcileDataloadDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	/*
		Delete helm release if exists
	*/
	datasetName := ctx.DataLoad.Spec.DatasetName
	releaseName := fmt.Sprintf("%s-load", datasetName)
	existed, err := helm.CheckRelease(releaseName, ctx.DataLoad.Namespace)
	if err != nil {
		ctx.Log.Error(err, "Helm check release error")
	} else if existed {
		if err := helm.DeleteRelease(releaseName, ctx.DataLoad.Namespace); err != nil {
			// TODO(xuzhihao) retry deleting
			ctx.Log.Error(err, "Helm can't uninstall release")
		}
		ctx.Log.Info("Helm release successfully deleted", "releaseName", releaseName)
	} else {
		ctx.Log.Info("Related Helm release not found, just delete dataload object")
	}

	/*
		Remove finalizer
	*/
	ctx.Log.Info("Start to clean up finalizer", "dataload", ctx.DataLoad)
	if !ctx.DataLoad.ObjectMeta.GetDeletionTimestamp().IsZero() {
		ctx.DataLoad.ObjectMeta.Finalizers = utils.RemoveString(ctx.DataLoad.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, &ctx.DataLoad); err != nil {
			ctx.Log.Error(err, "Failed to remove finalizer")
			return utils.RequeueIfError(err)
		}
		ctx.Log.Info("Finalizer is removed", "dataload", ctx.DataLoad)
	}

	return utils.NoRequeue()
}

func (r *DataLoadReconciler) addFinalizerAndRequeue(ctx reconcileRequestContext) (ctrl.Result, error) {
	ctx.DataLoad.ObjectMeta.Finalizers = append(ctx.DataLoad.ObjectMeta.Finalizers, finalizer)
	ctx.Log.Info("Add finalizer and Requeue", "finalizer", finalizer)
	prevGeneration := ctx.DataLoad.ObjectMeta.GetGeneration()
	if err := r.Update(ctx, &ctx.DataLoad); err != nil {
		ctx.Log.Error(err, "Failed to add finalizer to dataload", "StatusUpdateError", err)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, ctx.DataLoad.ObjectMeta.GetGeneration())
}

func (r *DataLoadReconciler) ReconcilePendingDataload(ctx reconcileRequestContext) (ctrl.Result, error) {

	/*
		1. Check if related dataset ready (i.e. related dataset exists and bounded)
	*/
	dataset, ready, err := r.CheckRelatedDatasetReady(ctx)
	if err != nil {
		ctx.Log.Error(err, "Failed to check related dataset status")
		return utils.RequeueIfError(err)
	} else if !ready {
		if dataset == nil {
			ctx.Log.Info("Related dataset not found", "datasetName", ctx.DataLoad.Spec.DatasetName, "namespace", ctx.DataLoad.Namespace)
		} else {
			ctx.Log.Info("Related dataset not ready", "datasetName", dataset.Name, "namespace", dataset.Namespace)
		}
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}
	ctx.Log.Info("Related dataset ready", "datasetName", dataset.Name, "namespace", dataset.Namespace)

	/*
		2. Check if related AlluxioRuntime ready (i.e. related alluxioRuntime exists and setup done)
	*/
	alluxioRuntime, ready, err := r.CheckRelatedRuntimeReady(ctx)
	if err != nil {
		ctx.Log.Error(err, "Failed to check related runtime status")
		return utils.RequeueIfError(err)
	} else if !ready {
		if alluxioRuntime == nil {
			ctx.Log.Info("Related alluxio runtime not found", "runtimeName", ctx.DataLoad.Spec.DatasetName, "namespace", ctx.DataLoad.Namespace)
		} else {
			ctx.Log.Info("Related alluxio runtime not ready", "runtimeName", alluxioRuntime.Name, "namespace", alluxioRuntime.Namespace, "scheduledWorkers", alluxioRuntime.Status.CurrentWorkerNumberScheduled, "availableWorkers", alluxioRuntime.Status.WorkerNumberAvailable)
		}
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}
	ctx.Log.Info("Related dataset ready", "runtimeName", alluxioRuntime.Name, "namespace", alluxioRuntime.Namespace)

	/*
		3. Check if there is any dataload with collision. Backoff if there exists
	*/
	dataloadWithCollision, err := r.FindDataLoadWithCollision(ctx)
	if err != nil {
		ctx.Log.Error(err, "Failed to find other dataload")
		return utils.RequeueIfError(err)
	}

	if dataloadWithCollision != nil {
		ctx.Log.Info("Another Dataload is loading", "dataloadName", dataloadWithCollision.Name, "namespace", dataloadWithCollision.Namespace)
		//Backoff
		return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
	}
	ctx.Log.Info("No other dataload loading, ready to prefetch...")

	/*
		4. Launch Prefetch Job
	*/
	installDone, err := r.LaunchPrefetchJob(ctx, alluxioRuntime.Status.WorkerNumberAvailable)
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
	//TODO(xuzhihao) retry to do updating
	dataload := ctx.DataLoad.DeepCopy()
	dataload.Status.Phase = common.DataloadPhaseLoading
	if err := r.Status().Update(ctx, dataload); err != nil {
		ctx.Log.Error(err, "Fail to update dataload phase", "Status update error", ctx)
		return utils.RequeueIfError(err)
	}
	return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
}

func (r *DataLoadReconciler) ReconcileDataload(ctx reconcileRequestContext) (ctrl.Result, error) {

	/*
		DataLoadPhase: Complete/Failed
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseComplete ||
		ctx.DataLoad.Status.Phase == common.DataloadPhaseFailed {
		//TODO(xuzhihao): If failed, backoff and retry
		//No need to reconcile
		return utils.NoRequeue()
	}

	/*
		DataLoadPhase: None -> Pending
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseNone {
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

	/*
		DataLoadPhase: Pending -> Loading
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhasePending {
		return r.ReconcilePendingDataload(ctx)
	}

	/*
		DataLoadPhase: Loading -> Complete/Failed
	*/
	if ctx.DataLoad.Status.Phase == common.DataloadPhaseLoading {
		return r.ReconcileLoadingDataload(ctx)
	}

	//unreachable in theory
	return ctrl.Result{}, nil
}

func (r *DataLoadReconciler) ReconcileLoadingDataload(ctx reconcileRequestContext) (ctrl.Result, error) {
	jobName := types.NamespacedName{
		Namespace: ctx.DataLoad.Namespace,
		Name:      fmt.Sprintf("%s-loader", ctx.DataLoad.Spec.DatasetName),
	}

	job := &batchv1.Job{}
	dataloadToUpdate := ctx.DataLoad.DeepCopy()
	var needUpdate bool
	if err := r.Get(ctx, jobName, job); err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related job not found, turn phase to `None`", "dataload", ctx.DataLoad)
			dataloadToUpdate.Status.Phase = common.DataloadPhaseNone
			needUpdate = true
		} else {
			ctx.Log.Error(err, "Fail to get related job", "jobName", jobName)
			return utils.RequeueIfError(err)
		}
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
		} else {
			ctx.Log.Info("Related Job complete", "jobName", job.Name)
			dataloadToUpdate.Status.Phase = common.DataloadPhaseComplete
		}
		needUpdate = true
	}

	if needUpdate {
		//TODO(xuzhihao) retry
		if err := r.Status().Update(ctx, dataloadToUpdate); err != nil {
			ctx.Log.Error(err, "Failed to update status with a loading dataload", "dataloadName", ctx.DataLoad.Name)
			return utils.RequeueIfError(err)
		}
	}

	return utils.RequeueAfterInterval(time.Duration(10 * time.Second))
}

//SetupWithManager setups the manager with AlluxioDataLoad
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.AlluxioDataLoad{}).
		Complete(r)
}

func (r *DataLoadReconciler) CheckRelatedRuntimeReady(ctx reconcileRequestContext) (alluxioRuntime *datav1alpha1.AlluxioRuntime, ready bool, err error) {
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

func (r *DataLoadReconciler) CheckRelatedDatasetReady(ctx reconcileRequestContext) (dataset *datav1alpha1.Dataset, ready bool, err error) {
	datasetName := ctx.DataLoad.Spec.DatasetName
	dataset, err = utils.GetDataset(r, datasetName, ctx.DataLoad.Namespace)
	if err != nil || dataset == nil {
		// Ignore not found error and take it as not ready
		err = utils.IgnoreNotFound(err)
		return nil, false, err
	}
	return dataset, dataset.Status.Phase == datav1alpha1.BoundDatasetPhase, nil
}

func (r *DataLoadReconciler) FindDataLoadWithCollision(ctx reconcileRequestContext) (dataLoad *datav1alpha1.AlluxioDataLoad, err error) {
	collisionFunc := func(dl datav1alpha1.AlluxioDataLoad) bool {
		// A dataload with collision is another dataload object with same DatasetName and loading phase
		return dl.Name != ctx.DataLoad.Name && dl.Spec.DatasetName == ctx.DataLoad.Spec.DatasetName && dl.Status.Phase == common.DataloadPhaseLoading
	}
	return utils.FindDataLoadWithPredicate(r, ctx.DataLoad.Namespace, collisionFunc)
}

func (r *DataLoadReconciler) LaunchPrefetchJob(ctx reconcileRequestContext, numWorker int32) (done bool, err error) {
	/*
		1. Check Helm releases
	*/
	releaseName := fmt.Sprintf("%s-load", ctx.DataLoad.Spec.DatasetName)
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

	const dataload_chart = "fluid-dataloader"
	var chartName = utils.GetChartsDirectory() + "/" + dataload_chart

	/*
		3. Install release with generated value file
	*/

	if err := helm.InstallRelease(releaseName, ctx.DataLoad.Namespace, valueFileName, chartName); err != nil {
		ctx.Log.Error(err, "Fail to install helm chart")
		//TODO(xuzhihao) inform error like `configuration not right` to events
		return false, err
	}

	ctx.Log.Info("Helm chart installed", "releaseName", releaseName)
	return true, nil
}

func (r *DataLoadReconciler) generateValueFile(dataload datav1alpha1.AlluxioDataLoad, numWorker int32) (valueFileName string, err error) {

	const image = "registry.cn-hangzhou.aliyuncs.com/fluid-namespace/coco-perf"

	dataloadConfig := DataLoadInfo{
		BackoffLimit: 6,
		Image:        image,
		NumWorker:    numWorker,
	}

	if dataload.Spec.SlotsPerNode != nil {
		dataloadConfig.Threads = *dataload.Spec.SlotsPerNode
	}

	if dataload.Spec.Path != "" {
		dataloadConfig.MountPath = dataload.Spec.Path
	}

	value := &DataLoadValue{DataLoadInfo: dataloadConfig}
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
