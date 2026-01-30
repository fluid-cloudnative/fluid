/*
Copyright 2026 The Fluid Authors.

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

package dataclean

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	alluxioops "github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	jindoops "github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

const (
	controllerName                 string = "DataCleanController"
	alluxioCleanCacheTimeoutSeconds int32  = 60
)

type DataCleanReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func NewDataCleanReconciler(
	c client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *DataCleanReconciler {
	return &DataCleanReconciler{
		Client:   c,
		Log:      log,
		Scheme:   scheme,
		Recorder: recorder,
	}
}

// +kubebuilder:rbac:groups=data.fluid.io,resources=datacleans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=datacleans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch

func (r *DataCleanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("DataClean", req.NamespacedName)

	dataClean, err := utils.GetDataClean(ctx, r.Client, req.Name, req.Namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get DataClean")
		return utils.RequeueIfError(errors.Wrap(err, "failed to get DataClean info"))
	}

	if !dataClean.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if dataClean.Status.Phase == common.PhaseComplete || dataClean.Status.Phase == common.PhaseFailed {
		return r.reconcileTTL(ctx, dataClean, log)
	}

	datasetName := dataClean.Spec.Dataset.Name
	datasetNamespace := dataClean.Spec.Dataset.Namespace
	if datasetNamespace == "" {
		datasetNamespace = dataClean.Namespace
	}

	dataset, err := utils.GetDataset(r.Client, datasetName, datasetNamespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			log.Info("Target Dataset not found, will requeue", "dataset", fmt.Sprintf("%s/%s", datasetNamespace, datasetName))
			r.Recorder.Eventf(
				dataClean,
				v1.EventTypeWarning,
				common.TargetDatasetNotFound,
				"Target dataset %s/%s not found for DataClean",
				datasetNamespace,
				datasetName,
			)
			return utils.RequeueAfterInterval(20 * time.Second)
		}
		log.Error(err, "failed to get target Dataset")
		return utils.RequeueIfError(errors.Wrap(err, "unable to get target Dataset"))
	}

	if dataClean.Namespace != datasetNamespace {
		msg := "DataClean namespace must be the same as Dataset namespace"
		log.Info(msg, "dataCleanNamespace", dataClean.Namespace, "datasetNamespace", datasetNamespace)
		r.Recorder.Eventf(dataClean, v1.EventTypeWarning, common.TargetDatasetNamespaceNotSame, msg)
		now := metav1.Now()
		dataClean.Status.Phase = common.PhaseFailed
		dataClean.Status.Conditions = []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             common.TargetDatasetNamespaceNotSame,
				Message:            msg,
				LastProbeTime:      now,
				LastTransitionTime: now,
			},
		}
		if err = r.Status().Update(ctx, dataClean); err != nil {
			log.Error(err, "failed to update DataClean status for namespace mismatch")
			return utils.RequeueIfError(err)
		}
		return ctrl.Result{}, nil
	}

	index, boundedRuntime := utils.GetRuntimeByCategory(dataset.Status.Runtimes, common.AccelerateCategory)
	if index == -1 {
		log.Info("Bounded runtime with Accelerate Category is not found on the target dataset", "dataset", dataset.Name)
		r.Recorder.Eventf(
			dataClean,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready for dataset %s/%s",
			dataset.Namespace,
			dataset.Name,
		)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	start := time.Now()
	cleanErr := r.cleanCacheForRuntime(ctx, dataset, boundedRuntime.Type)
	duration := time.Since(start)

	now := metav1.Now()
	if cleanErr != nil {
		log.Error(cleanErr, "cache cleaning failed", "runtimeType", boundedRuntime.Type)
		r.Recorder.Eventf(
			dataClean,
			v1.EventTypeWarning,
			common.DataOperationExecutionFailed,
			"fail to execute DataClean on runtime %s: %v",
			boundedRuntime.Type,
			cleanErr,
		)
		dataClean.Status.Phase = common.PhaseFailed
		dataClean.Status.Duration = duration.String()
		dataClean.Status.Conditions = []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             common.DataOperationExecutionFailed,
				Message:            cleanErr.Error(),
				LastProbeTime:      now,
				LastTransitionTime: now,
			},
		}
	} else {
		log.Info("cache cleaning succeeded", "runtimeType", boundedRuntime.Type)
		r.Recorder.Eventf(
			dataClean,
			v1.EventTypeNormal,
			common.DataOperationSucceed,
			"DataClean %s succeeded on runtime %s",
			dataClean.Name,
			boundedRuntime.Type,
		)
		dataClean.Status.Phase = common.PhaseComplete
		dataClean.Status.Duration = duration.String()
		dataClean.Status.Conditions = []datav1alpha1.Condition{
			{
				Type:               common.Complete,
				Status:             v1.ConditionTrue,
				Reason:             common.DataOperationSucceed,
				Message:            "DataClean completed successfully",
				LastProbeTime:      now,
				LastTransitionTime: now,
			},
		}
	}

	if err = r.Status().Update(ctx, dataClean); err != nil {
		log.Error(err, "failed to update DataClean status after execution")
		return utils.RequeueIfError(err)
	}

	return r.reconcileTTL(ctx, dataClean, log)
}

func (r *DataCleanReconciler) reconcileTTL(ctx context.Context, dataClean *datav1alpha1.DataClean, log logr.Logger) (ctrl.Result, error) {
	if dataClean.Spec.TTLSecondsAfterFinished == nil {
		return ctrl.Result{}, nil
	}

	if len(dataClean.Status.Conditions) == 0 {
		return ctrl.Result{}, nil
	}

	var completionCond *datav1alpha1.Condition
	for i := range dataClean.Status.Conditions {
		cond := &dataClean.Status.Conditions[i]
		if cond.Type == common.Complete || cond.Type == common.Failed {
			completionCond = cond
			break
		}
	}

	if completionCond == nil {
		return ctrl.Result{}, nil
	}

	if completionCond.Type != common.Complete && completionCond.Type != common.Failed {
		return ctrl.Result{}, nil
	}

	ttlSeconds := *dataClean.Spec.TTLSecondsAfterFinished
	if ttlSeconds <= 0 {
		if err := r.Delete(ctx, dataClean); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "failed to delete DataClean for TTLSecondsAfterFinished=0")
			return utils.RequeueIfError(err)
		}
		log.Info("DataClean has been cleaned up immediately due to TTLSecondsAfterFinished=0")
		return ctrl.Result{}, nil
	}

	expireTime := completionCond.LastTransitionTime.Add(time.Duration(ttlSeconds) * time.Second)
	now := time.Now()
	if now.After(expireTime) {
		if err := r.Delete(ctx, dataClean); err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "failed to delete DataClean after TTL expiration")
			return utils.RequeueIfError(err)
		}
		log.Info("DataClean has been cleaned up after TTL expiration")
		return ctrl.Result{}, nil
	}

	remaining := expireTime.Sub(now)
	log.V(1).Info("requeue after remaining time to clean up DataClean", "timeToLive", remaining)
	return ctrl.Result{RequeueAfter: remaining}, nil
}

func (r *DataCleanReconciler) cleanCacheForRuntime(
	ctx context.Context,
	dataset *datav1alpha1.Dataset,
	runtimeType string,
) error {
	switch runtimeType {
	case common.AlluxioRuntime:
		return r.cleanAlluxioCache(ctx, dataset)
	case common.JindoRuntime:
		return r.cleanJindoCache(ctx, dataset)
	default:
		return fmt.Errorf("runtimeType %s is not supported for DataClean", runtimeType)
	}
}

func (r *DataCleanReconciler) cleanJindoCache(ctx context.Context, dataset *datav1alpha1.Dataset) error {
	namespace := dataset.Namespace
	name := dataset.Name

	masterName := fmt.Sprintf("%s-jindofs-master", name)
	master, err := kubeclient.GetStatefulSetWithContext(ctx, r.Client, masterName, namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			r.Log.Info("Jindo master not found when cleaning cache, skip", "master", masterName, "err", err.Error())
			return nil
		}
		return err
	}
	if master.Status.ReadyReplicas == 0 {
		r.Log.Info("Jindo master is not ready, skip cache cleaning", "master", masterName)
		return nil
	}

	podName := fmt.Sprintf("%s-jindofs-master-0", name)
	containerName := "jindofs-master"
	fileUtils := jindoops.NewJindoFileUtils(podName, containerName, namespace, r.Log)
	r.Log.Info("Cleaning Jindo cache via DataClean")
	return fileUtils.CleanCache()
}

func (r *DataCleanReconciler) cleanAlluxioCache(ctx context.Context, dataset *datav1alpha1.Dataset) error {
	namespace := dataset.Namespace
	name := dataset.Name

	masterName := fmt.Sprintf("%s-master", name)
	master, err := kubeclient.GetStatefulSetWithContext(ctx, r.Client, masterName, namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			r.Log.Info("Alluxio master not found when cleaning cache, skip", "master", masterName, "err", err.Error())
			return nil
		}
		return err
	}
	if master.Status.ReadyReplicas == 0 {
		r.Log.Info("Alluxio master is not ready, skip cache cleaning", "master", masterName)
		return nil
	}

	podName := fmt.Sprintf("%s-master-0", name)
	containerName := "alluxio-master"
	fileUtils := alluxioops.NewAlluxioFileUtils(podName, containerName, namespace, r.Log)
	r.Log.Info("Cleaning Alluxio cache via DataClean")
	return fileUtils.CleanCache("/", alluxioCleanCacheTimeoutSeconds)
}

func (r *DataCleanReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.DataClean{}).
		Complete(r)
}

func (r *DataCleanReconciler) ControllerName() string {
	return controllerName
}

