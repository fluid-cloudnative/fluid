package dataload

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
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

func (r *DataLoadReconcilerImplement) ReconcileDataLoadDeletion(ctx reconcileRequestContext) (ctrl.Result, error) {
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) ReconcileDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	ctx.Log.WithName("ReconcileDataLoad").V(1).Info("process the dataload", "dataload", ctx.DataLoad)

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
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) reconcilePendingDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) reconcileLoadingDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) reconcileLoadedDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	panic("Not Implemented")
}

func (r *DataLoadReconcilerImplement) reconcileFailedDataLoad(ctx reconcileRequestContext) (ctrl.Result, error) {
	panic("Not Implemented")
}
