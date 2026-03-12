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
	"testing"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func TestDefaultStateMachine_GetCurrentState(t *testing.T) {
	sm := NewDefaultStateMachine()

	// Test default state (Idle)
	state, err := sm.GetCurrentState("test-op-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if state != StateIdle {
		t.Errorf("Expected StateIdle, got %s", state)
	}
}

func TestDefaultStateMachine_CanTransition(t *testing.T) {
	sm := NewDefaultStateMachine()

	tests := []struct {
		name        string
		from        OperationState
		to          OperationState
		wantAllowed bool
	}{
		{"Idle to Initializing", StateIdle, StateInitializing, true},
		{"Initializing to Preparing", StateInitializing, StatePreparing, true},
		{"Preparing to Executing", StatePreparing, StateExecuting, true},
		{"Executing to Validating", StateExecuting, StateValidating, true},
		{"Validating to Completed", StateValidating, StateCompleted, true},
		{"Executing to Failed", StateExecuting, StateFailed, true},
		{"Executing to RollingBack", StateExecuting, StateRollingBack, true},
		{"RollingBack to Recovering", StateRollingBack, StateRecovering, true},
		{"Idle to Completed (invalid)", StateIdle, StateCompleted, false},
		{"Completed to Executing (invalid)", StateCompleted, StateExecuting, false},
		{"Self-transition", StateExecuting, StateExecuting, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := sm.CanTransition(tt.from, tt.to)
			if allowed != tt.wantAllowed {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, allowed, tt.wantAllowed)
			}
		})
	}
}

func TestDefaultStateMachine_TransitionState(t *testing.T) {
	sm := NewDefaultStateMachine()
	operationID := "test-op-1"
	ctx := cruntime.ReconcileRequestContext{}

	// Test valid transition
	err := sm.TransitionState(ctx, operationID, StateInitializing, "Starting operation")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	state, _ := sm.GetCurrentState(operationID)
	if state != StateInitializing {
		t.Errorf("Expected StateInitializing, got %s", state)
	}

	// Test invalid transition
	err = sm.TransitionState(ctx, operationID, StateCompleted, "Invalid transition")
	if err == nil {
		t.Errorf("Expected error for invalid transition, got nil")
	}
	if _, ok := err.(*InvalidStateTransitionError); !ok {
		t.Errorf("Expected InvalidStateTransitionError, got %T", err)
	}

	// Verify state didn't change
	state, _ = sm.GetCurrentState(operationID)
	if state != StateInitializing {
		t.Errorf("Expected state to remain StateInitializing, got %s", state)
	}
}

func TestDefaultStateMachine_GetStateHistory(t *testing.T) {
	sm := NewDefaultStateMachine()
	operationID := "test-op-2"

	// No history initially
	history, err := sm.GetStateHistory(operationID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d transitions", len(history))
	}

	// Perform transitions
	ctx := cruntime.ReconcileRequestContext{}
	sm.TransitionState(ctx, operationID, StateInitializing, "Start")
	sm.TransitionState(ctx, operationID, StatePreparing, "Prepare")
	sm.TransitionState(ctx, operationID, StateExecuting, "Execute")

	// Check history
	history, err = sm.GetStateHistory(operationID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(history) != 3 {
		t.Errorf("Expected 3 transitions, got %d", len(history))
	}

	// Verify transition details
	if history[0].FromState != StateIdle {
		t.Errorf("Expected first transition from StateIdle, got %s", history[0].FromState)
	}
	if history[0].ToState != StateInitializing {
		t.Errorf("Expected first transition to StateInitializing, got %s", history[0].ToState)
	}
	if history[2].ToState != StateExecuting {
		t.Errorf("Expected last transition to StateExecuting, got %s", history[2].ToState)
	}
}

func TestDefaultStateMachine_RegisterStateHandler(t *testing.T) {
	sm := NewDefaultStateMachine()
	handler := &testStateHandler{}

	err := sm.RegisterStateHandler(StateExecuting, handler)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify handler is registered by performing a transition that should trigger it
	operationID := "test-op-3"
	ctx := cruntime.ReconcileRequestContext{}
	sm.TransitionState(ctx, operationID, StateInitializing, "Start")
	sm.TransitionState(ctx, operationID, StatePreparing, "Prepare")
	sm.TransitionState(ctx, operationID, StateExecuting, "Execute")

	// Handler should have been called
	if handler.enterCount == 0 {
		t.Errorf("Expected OnEnter to be called, but it wasn't")
	}
}

func TestDefaultStateMachine_CompleteWorkflow(t *testing.T) {
	sm := NewDefaultStateMachine()
	operationID := "workflow-op"
	ctx := cruntime.ReconcileRequestContext{}

	// Test complete workflow: Idle -> Initializing -> Preparing -> Executing -> Validating -> Completed
	workflow := []struct {
		targetState OperationState
		reason      string
	}{
		{StateInitializing, "Start operation"},
		{StatePreparing, "Prepare resources"},
		{StateExecuting, "Execute operation"},
		{StateValidating, "Validate results"},
		{StateCompleted, "Operation completed"},
	}

	for _, step := range workflow {
		err := sm.TransitionState(ctx, operationID, step.targetState, step.reason)
		if err != nil {
			t.Errorf("Failed to transition to %s: %v", step.targetState, err)
		}

		state, _ := sm.GetCurrentState(operationID)
		if state != step.targetState {
			t.Errorf("Expected state %s, got %s", step.targetState, state)
		}
	}

	// Verify history
	history, _ := sm.GetStateHistory(operationID)
	if len(history) != len(workflow) {
		t.Errorf("Expected %d transitions in history, got %d", len(workflow), len(history))
	}

	// Verify final state
	finalState, _ := sm.GetCurrentState(operationID)
	if finalState != StateCompleted {
		t.Errorf("Expected final state Completed, got %s", finalState)
	}
}

func TestDefaultStateMachine_RollbackWorkflow(t *testing.T) {
	sm := NewDefaultStateMachine()
	operationID := "rollback-op"
	ctx := cruntime.ReconcileRequestContext{}

	// Simulate failure and rollback: Idle -> Initializing -> Executing -> RollingBack -> Failed
	sm.TransitionState(ctx, operationID, StateInitializing, "Start")
	sm.TransitionState(ctx, operationID, StatePreparing, "Prepare")
	sm.TransitionState(ctx, operationID, StateExecuting, "Execute")

	// Simulate error and rollback
	err := sm.TransitionState(ctx, operationID, StateRollingBack, "Error occurred")
	if err != nil {
		t.Errorf("Failed to transition to RollingBack: %v", err)
	}

	err = sm.TransitionState(ctx, operationID, StateFailed, "Rollback completed")
	if err != nil {
		t.Errorf("Failed to transition to Failed: %v", err)
	}

	// Verify final state
	state, _ := sm.GetCurrentState(operationID)
	if state != StateFailed {
		t.Errorf("Expected final state Failed, got %s", state)
	}
}

func TestDefaultExtendedLifecycleManager_StateMachineMethods(t *testing.T) {
	manager := &DefaultExtendedLifecycleManager{}

	// Test GetCurrentState
	state, err := manager.GetCurrentState("test-op")
	if err == nil {
		t.Errorf("Expected error for GetCurrentState, got nil")
	}
	if state != StateIdle {
		t.Errorf("Expected StateIdle as default, got %s", state)
	}

	// Test TransitionState
	ctx := cruntime.ReconcileRequestContext{}
	err = manager.TransitionState(ctx, "test-op", StateExecuting, "test")
	if err == nil {
		t.Errorf("Expected error for TransitionState, got nil")
	}

	// Test CanTransition
	allowed := manager.CanTransition(StateIdle, StateExecuting)
	if allowed {
		t.Errorf("Expected CanTransition to return false, got true")
	}

	// Test GetStateHistory
	history, err := manager.GetStateHistory("test-op")
	if err == nil {
		t.Errorf("Expected error for GetStateHistory, got nil")
	}
	if history != nil {
		t.Errorf("Expected nil history, got %v", history)
	}

	// Test RegisterStateHandler
	err = manager.RegisterStateHandler(StateExecuting, nil)
	if err == nil {
		t.Errorf("Expected error for RegisterStateHandler, got nil")
	}
}

func TestInvalidStateTransitionError(t *testing.T) {
	err := &InvalidStateTransitionError{
		From:   StateIdle,
		To:     StateCompleted,
		Reason: "invalid transition",
	}

	expectedMsg := "invalid state transition from Idle to Completed: invalid transition"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestDefaultStateMachine_HandlerErrorHandling(t *testing.T) {
	sm := NewDefaultStateMachine()
	operationID := "error-test-op"
	ctx := cruntime.ReconcileRequestContext{}

	// Test OnExit handler failure
	exitHandler := &testStateHandler{failOnExit: true}
	sm.RegisterStateHandler(StateIdle, exitHandler)

	err := sm.TransitionState(ctx, operationID, StateInitializing, "Start")
	if err == nil {
		t.Errorf("Expected error when OnExit handler fails, got nil")
	}
	if _, ok := err.(*StateTransitionError); !ok {
		t.Errorf("Expected StateTransitionError, got %T", err)
	}

	// Verify state didn't change
	state, _ := sm.GetCurrentState(operationID)
	if state != StateIdle {
		t.Errorf("Expected state to remain Idle after OnExit failure, got %s", state)
	}

	// Reset and test OnEnter handler failure
	sm = NewDefaultStateMachine()
	enterHandler := &testStateHandler{failOnEnter: true}
	sm.RegisterStateHandler(StateInitializing, enterHandler)

	err = sm.TransitionState(ctx, operationID, StateInitializing, "Start")
	if err == nil {
		t.Errorf("Expected error when OnEnter handler fails, got nil")
	}
	if _, ok := err.(*StateTransitionError); !ok {
		t.Errorf("Expected StateTransitionError, got %T", err)
	}

	// Verify state was rolled back
	state, _ = sm.GetCurrentState(operationID)
	if state != StateIdle {
		t.Errorf("Expected state to be rolled back to Idle after OnEnter failure, got %s", state)
	}
}

func TestDefaultStateMachine_OnTransitionHandler(t *testing.T) {
	sm := NewDefaultStateMachine()
	operationID := "transition-test-op"
	ctx := cruntime.ReconcileRequestContext{}

	handler := &testStateHandler{}
	// Register handler for a different state to test OnTransition is called for all handlers
	sm.RegisterStateHandler(StateCompleted, handler)

	// Perform transitions
	sm.TransitionState(ctx, operationID, StateInitializing, "Start")
	sm.TransitionState(ctx, operationID, StatePreparing, "Prepare")
	sm.TransitionState(ctx, operationID, StateExecuting, "Execute")
	sm.TransitionState(ctx, operationID, StateValidating, "Validate")
	sm.TransitionState(ctx, operationID, StateCompleted, "Complete")

	// OnTransition should have been called for each transition
	if handler.transitionCount != 5 {
		t.Errorf("Expected OnTransition to be called 5 times, got %d", handler.transitionCount)
	}
}

// testStateHandler is a test implementation of StateHandler
type testStateHandler struct {
	enterCount      int
	exitCount       int
	transitionCount int
	shouldFail      bool // For testing error handling
	failOnEnter     bool
	failOnExit      bool
}

func (h *testStateHandler) OnEnter(ctx cruntime.ReconcileRequestContext, operationID string, fromState OperationState) error {
	h.enterCount++
	if h.failOnEnter {
		return fmt.Errorf("OnEnter handler failed")
	}
	return nil
}

func (h *testStateHandler) OnExit(ctx cruntime.ReconcileRequestContext, operationID string, toState OperationState) error {
	h.exitCount++
	if h.failOnExit {
		return fmt.Errorf("OnExit handler failed")
	}
	return nil
}

func (h *testStateHandler) OnTransition(ctx cruntime.ReconcileRequestContext, operationID string, transition StateTransition) error {
	h.transitionCount++
	return nil
}
