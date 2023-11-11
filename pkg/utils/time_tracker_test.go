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
	"testing"
	"time"
)

func TestTimeTrack(t *testing.T) {
	taskTimeThreshold = 5 * time.Millisecond

	d := 1 * time.Millisecond
	mockTimeTrack(d)

	check := checkLongTask(d)

	if check {
		t.Errorf("expected it's not a long task with %v", d)
	}

	d = 10 * time.Millisecond
	mockTimeTrack(d)

	check = checkLongTask(d)

	if !check {
		t.Errorf("expected it's a long task with %v", d)
	}
}

func TestIsTimeTrackerDebugEnabled(t *testing.T) {

	type testCases struct {
		name   string
		debug  bool
		wanted bool
	}

	tests := []testCases{
		{

			name:   "disable",
			debug:  false,
			wanted: false,
		}, {

			name:   "enable",
			debug:  true,
			wanted: true,
		},
	}

	for _, testCase := range tests {
		enableTimeTrackDebug = testCase.debug
		if IsTimeTrackerDebugEnabled() != testCase.wanted {
			t.Errorf("test case %s failed due to expect IsTimeTrackerDebugEnabled is %v, but got %v", testCase.name, testCase.wanted, IsTimeTrackerDebugEnabled())
		}

		if IsTimeTrackerEnabled() != testCase.wanted {
			t.Errorf("test case %s failed due to expect IsTimeTrackerEnabled is %v, but got %v", testCase.name, testCase.wanted, IsTimeTrackerEnabled())
		}
	}

}

func mockTimeTrack(d time.Duration) {
	defer TimeTrack(time.Now(), "mockTimeTracker", "test", "test")
	time.Sleep(d)
}
