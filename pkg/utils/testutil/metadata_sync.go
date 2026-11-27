/*
Copyright 2026 The Fluid Authors.

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

package testutil

import (
	"errors"
	"fmt"
	"time"
)

// VerifyMetadataSyncChannelClosedOnce verifies that an async metadata sync initializes
// its result channel on the first call, receives one result and clears it on a follow-up
// call, and closes the channel cleanly without blocking on a duplicate send.
func VerifyMetadataSyncChannelClosedOnce[T any](syncFn func() error, getCh func() <-chan T) error {
	if err := syncFn(); err != nil {
		return fmt.Errorf("first sync call failed: %w", err)
	}
	ch := getCh()
	if ch == nil {
		return errors.New("expected metadata sync channel to be initialized")
	}

	deadline := time.Now().Add(5 * time.Second)
	for getCh() != nil {
		if err := syncFn(); err != nil {
			return fmt.Errorf("follow-up sync call failed: %w", err)
		}
		if time.Now().After(deadline) {
			return errors.New("timed out waiting for metadata sync channel to be cleared")
		}
	}

	select {
	case _, ok := <-ch:
		if ok {
			return errors.New("unexpected second value sent to metadata sync channel")
		}
		return nil
	case <-time.After(time.Second):
		return errors.New("timed out waiting for metadata sync channel to close; goroutine likely blocked on duplicate send")
	}
}
