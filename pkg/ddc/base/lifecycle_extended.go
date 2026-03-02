package base

import (
	"fmt"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// DefaultExtendedLifecycleManager provides default "not implemented" behavior
// for engines that don't need the full feature set yet.
type DefaultExtendedLifecycleManager struct{}

// --- DataLifecycleManager Implementation ---

func (d *DefaultExtendedLifecycleManager) LoadData(ctx cruntime.ReconcileRequestContext, spec *DataLoadSpec) (*DataLoadResult, error) {
	return nil, fmt.Errorf("LoadData not implemented by this engine")
}

func (d *DefaultExtendedLifecycleManager) GetDataOperationStatus(operationID string) (*DataOperationStatus, error) {
	return nil, fmt.Errorf("GetDataOperationStatus not implemented by this engine")
}

func (d *DefaultExtendedLifecycleManager) ProcessData(ctx cruntime.ReconcileRequestContext, spec *DataProcessSpec) (*DataProcessResult, error) {
	return nil, fmt.Errorf("ProcessData not implemented by this engine")
}

func (d *DefaultExtendedLifecycleManager) MutateData(ctx cruntime.ReconcileRequestContext, spec *DataMutationSpec) (*DataMutationResult, error) {
	return nil, fmt.Errorf("MutateData not implemented by this engine")
}

// --- StateMachineManager Implementation (Stubs) ---

func (d *DefaultExtendedLifecycleManager) GetCurrentState(operationID string) (OperationState, error) {
	return StateIdle, fmt.Errorf("GetCurrentState not implemented by this engine")
}

func (d *DefaultExtendedLifecycleManager) TransitionState(ctx cruntime.ReconcileRequestContext, operationID string, targetState OperationState, reason string) error {
	return fmt.Errorf("TransitionState not implemented by this engine")
}

func (d *DefaultExtendedLifecycleManager) CanTransition(currentState, targetState OperationState) bool {
	return false
}

func (d *DefaultExtendedLifecycleManager) GetStateHistory(operationID string) ([]StateTransition, error) {
	return nil, fmt.Errorf("GetStateHistory not implemented by this engine")
}

func (d *DefaultExtendedLifecycleManager) RegisterStateHandler(state OperationState, handler StateHandler) error {
	return fmt.Errorf("RegisterStateHandler not implemented by this engine")
}