package dataload

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
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

func (r *DataLoadReconcilerImplement) ReconcileDataLoadDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	//todo(xuzhihao) release targetDataset's lock if necessary
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) ReconcileDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("ReconcileDataLoad")
	log.V(1).Info("process the dataload", "dataload", ctx.DataLoad)

	if ctx.DataLoad.Status.Phase == common.DataLoadPhaseLoaded {
		return r.reconcileLoadedDataLoad(ctx)
	}

	if ctx.DataLoad.Status.Phase == common.DataLoadPhaseFailed {
		return r.reconcileFailedDataLoad(ctx)
	}

	if ctx.DataLoad.Status.Phase == common.DataLoadPhaseNone {
		return r.reconcileNoneDataLoad(ctx)
	}

	if ctx.DataLoad.Status.Phase == common.DataLoadPhasePending {
		return r.reconcilePendingDataLoad(ctx)
	}
	// ctx.DataLoad.Status.Phase == common.DataLoadPhaseLoading
	return r.reconcileLoadingDataLoad(ctx)
}

func (r *DataLoadReconcilerImplement) reconcileNoneDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileNoneDataLoad")
	dataloadToUpdate := ctx.DataLoad.DeepCopy()
	dataloadToUpdate.Status.Phase = common.DataLoadPhasePending
	if err := r.Status().Update(context.TODO(), dataloadToUpdate); err != nil {
		log.Error(err, "failed to update the dataload")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("Update phase of the dataload to Pending successfully")
	return utils.RequeueImmediately()
}

func (r *DataLoadReconcilerImplement) reconcilePendingDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcilePendingDataLoad")

	// 1. Check existence of the target dataset
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
	if targetDataset.Status.UfsTotal == "" || targetDataset.Status.UfsTotal == alluxio.METADATA_SYNC_NOT_DONE_MSG {
		log.V(1).Info("Target dataset not ready", "targetDataset", ctx.DataLoad.Spec.Dataset)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.TargetDatasetNotReady,
			"Target dataset(namespace: %s, name: %s) metadata sync not done",
			targetDataset.Namespace, targetDataset.Name)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 3. Check if there's any loading DataLoad jobs(conflict DataLoad)
	conflictDataLoadRef := targetDataset.Status.DataLoadRef
	myDataLoadRef := fmt.Sprintf("%s-%s", ctx.DataLoad.Namespace, ctx.DataLoad.Name)
	if len(conflictDataLoadRef) != 0 && conflictDataLoadRef != myDataLoadRef {
		log.V(1).Info("Found other DataLoads that is in Loading phase, will backoff", "other DataLoad", conflictDataLoadRef)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.DataLoadCollision,
			"Found other Dataload(%s) that is in Loading phase, will backoff",
			conflictDataLoadRef)
		return utils.RequeueAfterInterval(20 * time.Second)
	}

	// 4. Check if the bounded runtime is ready
	var ready bool
	var boundedRuntime v1alpha1.Runtime
	runtimes := targetDataset.Status.Runtimes
	for _, runtime := range runtimes {
		if runtime.Category != common.AccelerateCategory {
			continue
		}
		boundedRuntime = runtime
		switch runtime.Type {
		case common.ALLUXIO_RUNTIME:
			podName := fmt.Sprintf("%s-master-0", targetDataset.Name)
			containerName := "alluxio-master"
			fileUtils := operations.NewAlluxioFileUtils(podName, containerName, targetDataset.Namespace, ctx.Log)
			ready = fileUtils.Ready()
		default:
			log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported", "runtime", runtime)
		}
		// Assume there is at most one runtime with AccelerateCategory
		break
	}
	if !ready {
		log.V(1).Info("Bounded accelerate runtime not ready", "targetDataset", targetDataset)
		r.Recorder.Eventf(&ctx.DataLoad,
			v1.EventTypeNormal,
			common.RuntimeNotReady,
			"Bounded accelerate runtime not ready", "runtime", boundedRuntime)
	}

	// 5. lock the target dataset. Make sure only one DataLoad can win the lock and
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

	// 6. update phase to Loading
	// We offload the helm install logic to `reconcileLoadingDataLoad` to
	// avoid such a case that status.phase change successfully first but helm install failed,
	// where the DataLoad job will never start and all other DataLoads will be blocked forever.
	log.Info("Get lock on target dataset, try to update phase")
	dataLoadToUpdate := ctx.DataLoad.DeepCopy()
	dataLoadToUpdate.Status.Phase = common.DataLoadPhaseLoading
	if err = r.Client.Status().Update(context.TODO(), dataLoadToUpdate); err != nil {
		log.Error(err, "failed to update dataload's status to Loading, will retry")
		return utils.RequeueIfError(err)
	}
	log.V(1).Info("update dataload's status to Loading successfully")
	return utils.RequeueImmediately()
}

func (r *DataLoadReconcilerImplement) reconcileLoadingDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	//todo(xuzhihao): release lock either Failed or Loaded
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) reconcileLoadedDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	log := ctx.Log.WithName("reconcileLoadedDataLoad")
	log.Info("DataLoad Loaded, no need to requeue")
	return utils.NoRequeue()
}

func (r *DataLoadReconcilerImplement) reconcileFailedDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	//todo(xuzhihao): retry DataLoad after a period
	log := ctx.Log.WithName("reconcileFailedDataLoad")
	log.Info("DataLoad failed, won't requeue")
	return utils.NoRequeue()
}

func (r *DataLoadReconcilerImplement) listConflictDataLoads(ctx reconcileRequestContext) ([]v1alpha1.DataLoad, error) {
	var dataLoadList = &v1alpha1.DataLoadList{}
	err := r.Client.List(context.TODO(), dataLoadList)
	if err != nil {
		ctx.Log.Error(err, "fail to list all DataLoads")
		return nil, err
	}
	ret := []v1alpha1.DataLoad{}
	for _, item := range dataLoadList.Items {
		if item.Name == ctx.DataLoad.Name && item.Namespace == ctx.DataLoad.Namespace {
			continue
		}
		if item.Spec.Dataset.Name == ctx.DataLoad.Spec.Dataset.Name &&
			item.Spec.Dataset.Namespace == ctx.DataLoad.Spec.Dataset.Namespace &&
			item.Status.Phase == common.DataLoadPhaseLoading {
			ret = append(ret, item)
		}
	}
	return ret, nil
}
