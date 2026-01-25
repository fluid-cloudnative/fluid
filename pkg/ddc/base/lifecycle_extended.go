/*
Copyright 2024 The Fluid Authors.

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

package base

import (
	"fmt"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// DefaultExtendedLifecycleManager provides default implementations that return
// "not implemented" errors. Engines can embed this and override specific methods.
type DefaultExtendedLifecycleManager struct{}

// LoadData implements DataLifecycleManager interface
func (d *DefaultExtendedLifecycleManager) LoadData(ctx cruntime.ReconcileRequestContext, spec *DataLoadSpec) (*DataLoadResult, error) {
	return nil, fmt.Errorf("LoadData not implemented by this engine")
}

// ProcessData implements DataLifecycleManager interface
func (d *DefaultExtendedLifecycleManager) ProcessData(ctx cruntime.ReconcileRequestContext, spec *DataProcessSpec) (*DataProcessResult, error) {
	return nil, fmt.Errorf("ProcessData not implemented by this engine")
}

// MutateData implements DataLifecycleManager interface
func (d *DefaultExtendedLifecycleManager) MutateData(ctx cruntime.ReconcileRequestContext, spec *DataMutationSpec) (*DataMutationResult, error) {
	return nil, fmt.Errorf("MutateData not implemented by this engine")
}

// GetDataOperationStatus implements DataLifecycleManager interface
func (d *DefaultExtendedLifecycleManager) GetDataOperationStatus(operationID string) (*DataOperationStatus, error) {
	return nil, fmt.Errorf("GetDataOperationStatus not implemented by this engine")
}

// GetCurrentState implements StateMachineManager interface
func (d *DefaultExtendedLifecycleManager) GetCurrentState(operationID string) (OperationState, error) {
	return StateIdle, fmt.Errorf("GetCurrentState not implemented by this engine")
}

// TransitionState implements StateMachineManager interface
func (d *DefaultExtendedLifecycleManager) TransitionState(operationID string, targetState OperationState, reason string) error {
	return fmt.Errorf("TransitionState not implemented by this engine")
}

// CanTransition implements StateMachineManager interface
func (d *DefaultExtendedLifecycleManager) CanTransition(currentState, targetState OperationState) bool {
	return false
}

// GetStateHistory implements StateMachineManager interface
func (d *DefaultExtendedLifecycleManager) GetStateHistory(operationID string) ([]StateTransition, error) {
	return nil, fmt.Errorf("GetStateHistory not implemented by this engine")
}

// RegisterStateHandler implements StateMachineManager interface
func (d *DefaultExtendedLifecycleManager) RegisterStateHandler(state OperationState, handler StateHandler) error {
	return fmt.Errorf("RegisterStateHandler not implemented by this engine")
}
