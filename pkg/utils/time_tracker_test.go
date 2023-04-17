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
