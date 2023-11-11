/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

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
