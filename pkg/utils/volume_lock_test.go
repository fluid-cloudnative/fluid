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

import "testing"

func TestTryAcquire(t *testing.T) {
	lock := NewVolumeLocks()
	testcases := []struct {
		name     string
		volumeId string
		expected bool
	}{
		{
			name:     "acquire success for test-vol-1",
			volumeId: "test-vol-1",
			expected: true,
		},
		{
			name:     "acquire success for test-vol-2",
			volumeId: "test-vol-2",
			expected: true,
		},
		{
			name:     "acquire failed for test-vol-1",
			volumeId: "test-vol-1",
			expected: false,
		},
	}
	for _, test := range testcases {
		acquire := lock.TryAcquire(test.volumeId)
		if acquire != test.expected {
			t.Errorf("%s failed, TryAcquire wants %v, got %v", test.name, test.expected, acquire)
		}
	}
}

func TestRelease(t *testing.T) {
	lock := NewVolumeLocks()
	vol1 := "test-vol-1"
	// release non-existing volume
	lock.Release(vol1)
	// release existing volume
	if acquire := lock.TryAcquire(vol1); !acquire {
		t.Errorf("TryAcquire failed on volume %s", vol1)
	}
	lock.Release(vol1)
	if len(lock.locks) != 0 {
		t.Errorf("volume %s has been released, expect 0 locks but got %v", vol1, len(lock.locks))
	}
	// repeat release
	lock.Release(vol1)
}
