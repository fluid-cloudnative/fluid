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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecoverStateTracker", func() {
	var tracker *RecoverStateTracker
	const testMountPath = "/var/lib/kubelet/pods/test-pod/volumes/kubernetes.io~csi/test-pv/mount"

	BeforeEach(func() {
		tracker = NewRecoverStateTracker()
	})

	Describe("NewRecoverStateTracker", func() {
		It("should create a new tracker with empty state map", func() {
			Expect(tracker).NotTo(BeNil())
			Expect(tracker.states).NotTo(BeNil())
			Expect(tracker.states).To(BeEmpty())
		})
	})

	Describe("GetOrCreateState", func() {
		It("should create new state for unknown mount path", func() {
			state := tracker.GetOrCreateState(testMountPath)
			Expect(state).NotTo(BeNil())
			Expect(state.IsHealthy).To(BeTrue())
			Expect(state.ConsecutiveFailures).To(Equal(0))
			Expect(state.CurrentBackoff).To(Equal(initialBackoff))
		})

		It("should return existing state for known mount path", func() {
			state1 := tracker.GetOrCreateState(testMountPath)
			state1.ConsecutiveFailures = 5

			state2 := tracker.GetOrCreateState(testMountPath)
			Expect(state2.ConsecutiveFailures).To(Equal(5))
			Expect(state1).To(BeIdenticalTo(state2))
		})
	})

	Describe("ShouldAttemptRecovery", func() {
		Context("when mount has no tracked state", func() {
			It("should allow recovery attempt", func() {
				Expect(tracker.ShouldAttemptRecovery(testMountPath)).To(BeTrue())
			})
		})

		Context("when mount has no consecutive failures", func() {
			It("should allow recovery attempt", func() {
				tracker.GetOrCreateState(testMountPath)
				Expect(tracker.ShouldAttemptRecovery(testMountPath)).To(BeTrue())
			})
		})

		Context("when mount is in backoff period", func() {
			It("should not allow recovery attempt", func() {
				tracker.GetOrCreateState(testMountPath)
				tracker.RecordFailure(testMountPath)

				// Immediately after failure, should be in backoff
				Expect(tracker.ShouldAttemptRecovery(testMountPath)).To(BeFalse())
			})
		})

		Context("when backoff period has elapsed", func() {
			It("should allow recovery attempt", func() {
				state := tracker.GetOrCreateState(testMountPath)
				tracker.RecordFailure(testMountPath)

				// Manually set last failure time to past
				state.LastFailureTime = time.Now().Add(-20 * time.Second)

				Expect(tracker.ShouldAttemptRecovery(testMountPath)).To(BeTrue())
			})
		})
	})

	Describe("RecordFailure", func() {
		It("should increment consecutive failures", func() {
			tracker.GetOrCreateState(testMountPath)

			tracker.RecordFailure(testMountPath)
			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.ConsecutiveFailures).To(Equal(1))

			tracker.RecordFailure(testMountPath)
			Expect(state.ConsecutiveFailures).To(Equal(2))
		})

		It("should mark mount as unhealthy", func() {
			tracker.GetOrCreateState(testMountPath)
			tracker.RecordFailure(testMountPath)

			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.IsHealthy).To(BeFalse())
		})

		It("should increase backoff exponentially", func() {
			tracker.GetOrCreateState(testMountPath)

			// First failure: backoff doubles from initial
			tracker.RecordFailure(testMountPath)
			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.CurrentBackoff).To(Equal(initialBackoff * 2))

			// Second failure: backoff doubles again
			tracker.RecordFailure(testMountPath)
			Expect(state.CurrentBackoff).To(Equal(initialBackoff * 4))
		})

		It("should cap backoff at maxBackoff", func() {
			tracker.GetOrCreateState(testMountPath)

			// Simulate many failures to hit the cap
			for i := 0; i < 20; i++ {
				tracker.RecordFailure(testMountPath)
			}

			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.CurrentBackoff).To(Equal(maxBackoff))
		})

		It("should record failure time", func() {
			before := time.Now()
			tracker.GetOrCreateState(testMountPath)
			tracker.RecordFailure(testMountPath)
			after := time.Now()

			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.LastFailureTime).To(BeTemporally(">=", before))
			Expect(state.LastFailureTime).To(BeTemporally("<=", after))
		})
	})

	Describe("RecordSuccess", func() {
		It("should reset consecutive failures to zero", func() {
			tracker.GetOrCreateState(testMountPath)
			tracker.RecordFailure(testMountPath)
			tracker.RecordFailure(testMountPath)

			tracker.RecordSuccess(testMountPath)

			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.ConsecutiveFailures).To(Equal(0))
		})

		It("should reset backoff to initial value", func() {
			tracker.GetOrCreateState(testMountPath)
			tracker.RecordFailure(testMountPath)
			tracker.RecordFailure(testMountPath)

			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.CurrentBackoff).To(BeNumerically(">", initialBackoff))

			tracker.RecordSuccess(testMountPath)
			Expect(state.CurrentBackoff).To(Equal(initialBackoff))
		})

		It("should mark mount as healthy", func() {
			tracker.GetOrCreateState(testMountPath)
			tracker.RecordFailure(testMountPath)

			state := tracker.GetOrCreateState(testMountPath)
			Expect(state.IsHealthy).To(BeFalse())

			tracker.RecordSuccess(testMountPath)
			Expect(state.IsHealthy).To(BeTrue())
		})

		It("should handle success on non-existent state gracefully", func() {
			Expect(func() { tracker.RecordSuccess(testMountPath) }).NotTo(Panic())
		})
	})

	Describe("ShouldEmitEvent", func() {
		const (
			reasonFailed  = "FuseRecoverFailed"
			reasonSucceed = "FuseRecoverSucceed"
		)

		Context("when mount has no tracked state", func() {
			It("should allow event emission", func() {
				Expect(tracker.ShouldEmitEvent(testMountPath, reasonFailed)).To(BeTrue())
			})
		})

		Context("when event reason is different from last emitted", func() {
			It("should allow event emission", func() {
				tracker.GetOrCreateState(testMountPath)

				// First event should be allowed
				Expect(tracker.ShouldEmitEvent(testMountPath, reasonFailed)).To(BeTrue())

				// Different event should be allowed
				Expect(tracker.ShouldEmitEvent(testMountPath, reasonSucceed)).To(BeTrue())
			})
		})

		Context("when event reason is same as last emitted", func() {
			It("should suppress duplicate event", func() {
				tracker.GetOrCreateState(testMountPath)

				// First event should be allowed
				Expect(tracker.ShouldEmitEvent(testMountPath, reasonFailed)).To(BeTrue())

				// Same event should be suppressed
				Expect(tracker.ShouldEmitEvent(testMountPath, reasonFailed)).To(BeFalse())
				Expect(tracker.ShouldEmitEvent(testMountPath, reasonFailed)).To(BeFalse())
			})
		})
	})

	Describe("RemoveState", func() {
		It("should remove state for mount path", func() {
			tracker.GetOrCreateState(testMountPath)
			Expect(tracker.states).To(HaveLen(1))

			tracker.RemoveState(testMountPath)
			Expect(tracker.states).To(BeEmpty())
		})

		It("should handle removal of non-existent state gracefully", func() {
			Expect(func() { tracker.RemoveState(testMountPath) }).NotTo(Panic())
		})
	})

	Describe("GetBackoffInfo", func() {
		It("should return zero failures and initial backoff for unknown mount", func() {
			failures, backoff := tracker.GetBackoffInfo(testMountPath)
			Expect(failures).To(Equal(0))
			Expect(backoff).To(Equal(initialBackoff))
		})

		It("should return current failure count and backoff", func() {
			tracker.GetOrCreateState(testMountPath)
			tracker.RecordFailure(testMountPath)
			tracker.RecordFailure(testMountPath)

			failures, backoff := tracker.GetBackoffInfo(testMountPath)
			Expect(failures).To(Equal(2))
			Expect(backoff).To(Equal(initialBackoff * 4)) // doubled twice
		})
	})

	Describe("Integration: Backoff and Event Deduplication", func() {
		It("should demonstrate bounded behavior under persistent failures", func() {
			tracker.GetOrCreateState(testMountPath)

			// Simulate 10 consecutive failures
			for i := 0; i < 10; i++ {
				tracker.RecordFailure(testMountPath)
			}

			state := tracker.GetOrCreateState(testMountPath)

			// Backoff should be capped
			Expect(state.CurrentBackoff).To(Equal(maxBackoff))

			// Events should be deduplicated
			Expect(tracker.ShouldEmitEvent(testMountPath, "FuseRecoverFailed")).To(BeTrue())  // First
			Expect(tracker.ShouldEmitEvent(testMountPath, "FuseRecoverFailed")).To(BeFalse()) // Duplicate
			Expect(tracker.ShouldEmitEvent(testMountPath, "FuseRecoverFailed")).To(BeFalse()) // Duplicate

			// Recovery should reset state
			tracker.RecordSuccess(testMountPath)
			Expect(state.ConsecutiveFailures).To(Equal(0))
			Expect(state.CurrentBackoff).To(Equal(initialBackoff))
			Expect(state.IsHealthy).To(BeTrue())

			// Next failure should allow event again (state transition)
			tracker.RecordFailure(testMountPath)
			Expect(tracker.ShouldEmitEvent(testMountPath, "FuseRecoverFailed")).To(BeTrue())
		})
	})
})
