package alluxio

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func (e *AlluxioEngine) LoadData(ctx cruntime.ReconcileRequestContext, spec *base.DataLoadSpec) (*base.DataLoadResult, error) {
	opID := fmt.Sprintf("Load-%s-%d", ctx.Name, spec.Priority)
	
	_ = e.TransitionState(ctx, opID, base.StateInitializing, "Starting")
	e.Log.Info("Alluxio Load Data", "Paths", spec.Paths)
	_ = e.TransitionState(ctx, opID, base.StateCompleted, "Done")

	return &base.DataLoadResult{
		OperationID: opID,
		Status: base.DataOperationStatus{
			Phase: base.DataOperationPhaseCompleted,
		},
	}, nil
}

func (e *AlluxioEngine) GetDataOperationStatus(operationID string) (*base.DataOperationStatus, error) {
	state, _ := e.GetCurrentState(operationID)
	return &base.DataOperationStatus{
		Message: string(state),
	}, nil
}