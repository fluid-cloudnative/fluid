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
	"time"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// StateMachineManager defines interfaces for managing data operation lifecycle
// and state transitions using a state machine pattern
type StateMachineManager interface {
	// GetCurrentState returns the current state of the operation
	GetCurrentState(operationID string) (OperationState, error)

	// TransitionState attempts to transition to a new state
	TransitionState(ctx cruntime.ReconcileRequestContext, operationID string, targetState OperationState, reason string) error

	// CanTransition checks if a state transition is allowed
	CanTransition(currentState, targetState OperationState) bool

	// GetStateHistory returns the history of state transitions
	GetStateHistory(operationID string) ([]StateTransition, error)

	// RegisterStateHandler registers a handler for state transitions
	RegisterStateHandler(state OperationState, handler StateHandler) error
}

// OperationState represents a state in the operation lifecycle
type OperationState string

const (
	// Initial states
	StateIdle         OperationState = "Idle"
	StateInitializing OperationState = "Initializing"

	// Active states
	StatePreparing  OperationState = "Preparing"
	StateExecuting  OperationState = "Executing"
	StateValidating OperationState = "Validating"

	// Completion states
	StateCompleted OperationState = "Completed"
	StateFailed    OperationState = "Failed"
	StateCancelled OperationState = "Cancelled"

	// Recovery states
	StateRollingBack OperationState = "RollingBack"
	StateRecovering  OperationState = "Recovering"
)

// StateTransition represents a state transition event
type StateTransition struct {
	// FromState is the state before transition
	FromState OperationState

	// ToState is the state after transition
	ToState OperationState

	// Reason explains why the transition occurred
	Reason string

	// Timestamp records when the transition occurred
	Timestamp time.Time

	// Metadata contains additional information about the transition
	Metadata map[string]string
}

// StateHandler handles state transition events
type StateHandler interface {
	// OnEnter is called when entering a state
	OnEnter(ctx cruntime.ReconcileRequestContext, operationID string, fromState OperationState) error

	// OnExit is called when exiting a state
	OnExit(ctx cruntime.ReconcileRequestContext, operationID string, toState OperationState) error

	// OnTransition is called during state transition
	OnTransition(ctx cruntime.ReconcileRequestContext, operationID string, transition StateTransition) error
}

// DefaultStateMachine provides a basic implementation of StateMachineManager
type DefaultStateMachine struct {
	states             map[string]OperationState
	history            map[string][]StateTransition
	handlers           map[OperationState]StateHandler
	allowedTransitions map[OperationState][]OperationState
}

// NewDefaultStateMachine creates a new default state machine
func NewDefaultStateMachine() *DefaultStateMachine {
	sm := &DefaultStateMachine{
		states:             make(map[string]OperationState),
		history:            make(map[string][]StateTransition),
		handlers:           make(map[OperationState]StateHandler),
		allowedTransitions: make(map[OperationState][]OperationState),
	}

	// Define allowed state transitions
	sm.defineTransitions()

	return sm
}

// defineTransitions defines the allowed state transitions
func (sm *DefaultStateMachine) defineTransitions() {
	// From Idle
	sm.allowedTransitions[StateIdle] = []OperationState{StateInitializing}

	// From Initializing
	sm.allowedTransitions[StateInitializing] = []OperationState{StatePreparing, StateFailed}

	// From Preparing
	sm.allowedTransitions[StatePreparing] = []OperationState{StateExecuting, StateFailed, StateCancelled}

	// From Executing
	sm.allowedTransitions[StateExecuting] = []OperationState{StateValidating, StateFailed, StateCancelled, StateRollingBack}

	// From Validating
	sm.allowedTransitions[StateValidating] = []OperationState{StateCompleted, StateFailed, StateRollingBack}

	// From RollingBack
	sm.allowedTransitions[StateRollingBack] = []OperationState{StateFailed, StateRecovering}

	// From Recovering
	sm.allowedTransitions[StateRecovering] = []OperationState{StateIdle, StateFailed}
}

// GetCurrentState returns the current state of the operation
func (sm *DefaultStateMachine) GetCurrentState(operationID string) (OperationState, error) {
	state, exists := sm.states[operationID]
	if !exists {
		return StateIdle, nil // Default to Idle if not found
	}
	return state, nil
}

// TransitionState attempts to transition to a new state
// The method ensures state consistency by:
// 1. Validating the transition is allowed
// 2. Calling OnExit handler (if fails, abort transition)
// 3. Updating state
// 4. Calling OnEnter handler (if fails, rollback state)
// 5. Calling OnTransition handler (for all registered handlers)
func (sm *DefaultStateMachine) TransitionState(ctx cruntime.ReconcileRequestContext, operationID string, targetState OperationState, reason string) error {
	currentState, err := sm.GetCurrentState(operationID)
	if err != nil {
		return err
	}

	if !sm.CanTransition(currentState, targetState) {
		return &InvalidStateTransitionError{
			From:   currentState,
			To:     targetState,
			Reason: "transition not allowed",
		}
	}

	// Prepare transition record
	transition := StateTransition{
		FromState: currentState,
		ToState:   targetState,
		Reason:    reason,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}

	// Step 1: Call OnExit handler for current state (if exists)
	if handler, exists := sm.handlers[currentState]; exists {
		if err := handler.OnExit(ctx, operationID, targetState); err != nil {
			return &StateTransitionError{
				OperationID: operationID,
				From:        currentState,
				To:          targetState,
				HandlerType: "OnExit",
				Err:         err,
			}
		}
	}

	// Step 2: Update state
	sm.states[operationID] = targetState
	sm.history[operationID] = append(sm.history[operationID], transition)

	// Step 3: Call OnEnter handler for target state (if exists)
	if handler, exists := sm.handlers[targetState]; exists {
		if err := handler.OnEnter(ctx, operationID, currentState); err != nil {
			// Rollback state on failure
			sm.states[operationID] = currentState
			// Remove the failed transition from history
			if len(sm.history[operationID]) > 0 {
				sm.history[operationID] = sm.history[operationID][:len(sm.history[operationID])-1]
			}
			return &StateTransitionError{
				OperationID: operationID,
				From:        currentState,
				To:          targetState,
				HandlerType: "OnEnter",
				Err:         err,
			}
		}
	}

	// Step 4: Call OnTransition handler for all registered handlers
	var transitionErrors []error
	for state, handler := range sm.handlers {
		if err := handler.OnTransition(ctx, operationID, transition); err != nil {
			transitionErrors = append(transitionErrors, fmt.Errorf("OnTransition error in handler for state %s: %w", state, err))
		}
	}

	if len(transitionErrors) > 0 {
		return &StateTransitionError{
			OperationID: operationID,
			From:        currentState,
			To:          targetState,
			HandlerType: "OnTransition",
			Err:         fmt.Errorf("multiple OnTransition errors: %v", transitionErrors),
		}
	}

	return nil
}

// CanTransition checks if a state transition is allowed
func (sm *DefaultStateMachine) CanTransition(currentState, targetState OperationState) bool {
	if currentState == targetState {
		return true // Self-transition is allowed
	}

	allowed, exists := sm.allowedTransitions[currentState]
	if !exists {
		return false
	}

	for _, state := range allowed {
		if state == targetState {
			return true
		}
	}

	return false
}

// GetStateHistory returns the history of state transitions
func (sm *DefaultStateMachine) GetStateHistory(operationID string) ([]StateTransition, error) {
	history, exists := sm.history[operationID]
	if !exists {
		return []StateTransition{}, nil
	}
	return history, nil
}

// RegisterStateHandler registers a handler for state transitions
func (sm *DefaultStateMachine) RegisterStateHandler(state OperationState, handler StateHandler) error {
	sm.handlers[state] = handler
	return nil
}

// InvalidStateTransitionError represents an invalid state transition error
type InvalidStateTransitionError struct {
	From   OperationState
	To     OperationState
	Reason string
}

func (e *InvalidStateTransitionError) Error() string {
	return "invalid state transition from " + string(e.From) + " to " + string(e.To) + ": " + e.Reason
}

// StateTransitionError represents an error during state transition handler execution
type StateTransitionError struct {
	OperationID string
	From        OperationState
	To          OperationState
	HandlerType string // "OnExit", "OnEnter", or "OnTransition"
	Err         error
}

func (e *StateTransitionError) Error() string {
	return "state transition error in " + e.HandlerType + " handler for operation " + e.OperationID +
		" (from " + string(e.From) + " to " + string(e.To) + "): " + e.Err.Error()
}

func (e *StateTransitionError) Unwrap() error {
	return e.Err
}
