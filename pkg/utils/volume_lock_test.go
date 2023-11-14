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
