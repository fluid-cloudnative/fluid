package base

import (
	"sync"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"time"
)

type StateMachineManager interface {
	GetCurrentState(operationID string) (OperationState, error)
	TransitionState(ctx cruntime.ReconcileRequestContext, operationID string, targetState OperationState, reason string) error
	CanTransition(currentState, targetState OperationState) bool
	GetStateHistory(operationID string) ([]StateTransition, error)
	RegisterStateHandler(state OperationState, handler StateHandler) error
}

type OperationState string

const (
	StateIdle         OperationState = "Idle"
	StateInitializing OperationState = "Initializing"
	StateExecuting    OperationState = "Executing"
	StateValidating   OperationState = "Validating"
	StateCompleted    OperationState = "Completed"
	StateFailed       OperationState = "Failed"
)

type StateTransition struct {
	FromState OperationState
	ToState   OperationState
	Reason    string
	Timestamp time.Time
}

type StateHandler interface {
	OnEnter(ctx cruntime.ReconcileRequestContext, operationID string) error
}

type DefaultStateMachine struct {
	mu     sync.RWMutex
	states map[string]OperationState
}

func NewDefaultStateMachine() *DefaultStateMachine {
	return &DefaultStateMachine{
		states: make(map[string]OperationState),
	}
}

func (sm *DefaultStateMachine) GetCurrentState(operationID string) (OperationState, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if state, ok := sm.states[operationID]; ok {
		return state, nil
	}
	return StateIdle, nil
}

func (sm *DefaultStateMachine) TransitionState(ctx cruntime.ReconcileRequestContext, operationID string, targetState OperationState, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.states[operationID] = targetState
	return nil
}

func (sm *DefaultStateMachine) CanTransition(current, target OperationState) bool {
	return true 
}

func (sm *DefaultStateMachine) GetStateHistory(operationID string) ([]StateTransition, error) {
	return nil, nil
}

func (sm *DefaultStateMachine) RegisterStateHandler(state OperationState, handler StateHandler) error {
	return nil
}