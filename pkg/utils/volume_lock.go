package utils

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

type VolumeLocks struct {
	locks sets.String
	mutex sync.Mutex
}

func NewVolumeLocks() *VolumeLocks {
	return &VolumeLocks{
		locks: sets.NewString(),
	}
}

// TryAcquire tries to acquire the lock for operating on resourceID and returns true if successful.
// If another operation is already using resourceID, returns false.
func (lock *VolumeLocks) TryAcquire(volumeID string) bool {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	if lock.locks.Has(volumeID) {
		return false
	}
	lock.locks.Insert(volumeID)
	return true
}

// Release releases lock in volume level
func (lock *VolumeLocks) Release(volumeID string) {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	lock.locks.Delete(volumeID)
}
