/*
Copyright 2022 The Fluid Authors.

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

package recover

import (
	"sync"
	"time"
)

const (
	// initialBackoff is the initial backoff duration after a recovery failure.
	// This value should be aligned with the default recovery period.
	initialBackoff = 5 * time.Second

	// maxBackoff is the maximum backoff duration to prevent excessively long waits.
	// After reaching this cap, retries occur at most every 5 minutes.
	maxBackoff = 5 * time.Minute

	// backoffMultiplier controls exponential backoff growth rate.
	// With multiplier of 2, backoff sequence is: 5s -> 10s -> 20s -> 40s -> 80s -> 160s -> 300s (capped)
	backoffMultiplier = 2.0
)

// MountState tracks the recovery state for a single mount point.
// This enables per-mount backoff and event deduplication without
// global locks affecting other mounts.
type MountState struct {
	// LastFailureTime records when the last failure occurred.
	// Used to calculate if backoff period has elapsed.
	LastFailureTime time.Time

	// ConsecutiveFailures counts sequential failures for backoff calculation.
	// Reset to 0 on successful recovery.
	ConsecutiveFailures int

	// LastEventReason stores the reason of the last emitted event to detect state changes.
	// Events are only emitted when this changes, preventing duplicate events.
	LastEventReason string

	// CurrentBackoff is the current backoff duration for this mount.
	// Increases exponentially with consecutive failures, capped at maxBackoff.
	CurrentBackoff time.Duration

	// IsHealthy tracks the health state for state-change event emission.
	// Used to emit recovery events only when transitioning from unhealthy to healthy.
	IsHealthy bool
}

// RecoverStateTracker manages per-mount recovery state with thread-safe access.
// It enables bounded retry behavior and event deduplication across the recover loop.
//
// Why this is needed:
// - Without backoff, a persistently broken mount causes retries every ~5s indefinitely
// - Without event deduplication, each retry emits a new Kubernetes event
// - This can cause 720+ events/hour per broken mount, overloading the API server
type RecoverStateTracker struct {
	mu     sync.RWMutex
	states map[string]*MountState // keyed by mount path
}

// NewRecoverStateTracker creates a new state tracker for mount recovery.
func NewRecoverStateTracker() *RecoverStateTracker {
	return &RecoverStateTracker{
		states: make(map[string]*MountState),
	}
}

// GetOrCreateState retrieves or initializes state for a mount point.
// Uses fine-grained locking to avoid blocking other mount operations.
func (t *RecoverStateTracker) GetOrCreateState(mountPath string) *MountState {
	t.mu.Lock()
	defer t.mu.Unlock()

	if state, exists := t.states[mountPath]; exists {
		return state
	}

	state := &MountState{
		CurrentBackoff: initialBackoff,
		IsHealthy:      true, // Assume healthy until proven otherwise
	}
	t.states[mountPath] = state
	return state
}

// RemoveState cleans up state for a mount that no longer exists.
// Should be called when a mount point is successfully unmounted or cleaned up.
func (t *RecoverStateTracker) RemoveState(mountPath string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.states, mountPath)
}

// ShouldAttemptRecovery checks if enough time has passed since the last
// failure to attempt recovery again. This implements exponential backoff
// to prevent tight retry loops on persistent failures.
//
// Why backoff is needed:
// - Persistent failures (e.g., unreachable storage, broken FUSE) won't self-heal quickly
// - Retrying every 5s wastes CPU/memory resources and floods logs/events
// - Exponential backoff reduces load while still enabling eventual recovery
// - After 6 consecutive failures, backoff reaches 5 minutes (capped)
func (t *RecoverStateTracker) ShouldAttemptRecovery(mountPath string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.states[mountPath]
	if !exists {
		return true // No state means first attempt
	}

	if state.ConsecutiveFailures == 0 {
		return true // No recent failures, attempt immediately
	}

	// Check if backoff period has elapsed
	nextAttemptTime := state.LastFailureTime.Add(state.CurrentBackoff)
	return time.Now().After(nextAttemptTime)
}

// RecordFailure updates state after a failed recovery attempt.
// Increases backoff exponentially up to maxBackoff.
//
// Backoff progression (with multiplier=2):
// - After failure 1: wait 10s
// - After failure 2: wait 20s
// - After failure 3: wait 40s
// - After failure 4: wait 80s
// - After failure 5: wait 160s
// - After failure 6+: wait 300s (5 min cap)
func (t *RecoverStateTracker) RecordFailure(mountPath string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.states[mountPath]
	if state == nil {
		state = &MountState{CurrentBackoff: initialBackoff}
		t.states[mountPath] = state
	}

	state.LastFailureTime = time.Now()
	state.ConsecutiveFailures++
	state.IsHealthy = false

	// Exponential backoff with cap
	// This bounds retry frequency: after ~6 failures, backoff reaches max
	newBackoff := time.Duration(float64(state.CurrentBackoff) * backoffMultiplier)
	if newBackoff > maxBackoff {
		newBackoff = maxBackoff
	}
	state.CurrentBackoff = newBackoff
}

// RecordSuccess resets state after successful recovery.
// This ensures quick retry if a future failure occurs.
func (t *RecoverStateTracker) RecordSuccess(mountPath string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.states[mountPath]
	if state == nil {
		return
	}

	wasUnhealthy := !state.IsHealthy
	state.ConsecutiveFailures = 0
	state.CurrentBackoff = initialBackoff
	state.IsHealthy = true
	// Only clear last event reason if we were unhealthy
	// This allows detecting the healthy->unhealthy transition next time
	if wasUnhealthy {
		state.LastEventReason = ""
	}
}

// ShouldEmitEvent determines if an event should be emitted for the given reason.
// Events are only emitted on state changes to prevent flooding.
//
// Why events are deduplicated:
// - Emitting the same "FuseRecoverFailed" event every 5s provides no new information
// - Event spam makes it hard to find meaningful signals in kubectl describe output
// - State-change events (healthy→broken, broken→recovered) are actionable
// - Operators can still see the issue from the single event + backoff behavior
func (t *RecoverStateTracker) ShouldEmitEvent(mountPath, eventReason string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.states[mountPath]
	if state == nil {
		// No state tracked yet, allow event emission
		return true
	}

	// Emit if event reason changed (state transition)
	if state.LastEventReason != eventReason {
		state.LastEventReason = eventReason
		return true
	}

	return false // Same state, suppress duplicate event
}

// GetBackoffInfo returns the current backoff state for observability/debugging.
// Returns consecutive failure count and current backoff duration.
func (t *RecoverStateTracker) GetBackoffInfo(mountPath string) (failures int, backoff time.Duration) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state := t.states[mountPath]
	if state == nil {
		return 0, initialBackoff
	}
	return state.ConsecutiveFailures, state.CurrentBackoff
}
