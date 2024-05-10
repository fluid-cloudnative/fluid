/*
Copyright 2023 The Fluid Authors.

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

package utils

import (
	"sync"
)

type VolumeLocks struct {
	locks map[string]struct{} // An empty struct uses 0 bytes
	mutex sync.Mutex
}

func NewVolumeLocks() *VolumeLocks {
	return &VolumeLocks{
		locks: make(map[string]struct{}),
	}
}

// TryAcquire tries to acquire the lock for operating on resourceID and returns true if successful.
// If another operation is already using resourceID, returns false.
func (lock *VolumeLocks) TryAcquire(volumeID string) bool {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	if _, ok := lock.locks[volumeID]; ok {
		return false
	}
	lock.locks[volumeID] = struct{}{}
	return true
}

// Release releases lock in volume level
func (lock *VolumeLocks) Release(volumeID string) {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	delete(lock.locks, volumeID) // Delete is a no-op if the key doesn't exist, so it's safe to call it unconditionally
}
