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

func mockTimeTrack(d time.Duration) {
	defer TimeTrack(time.Now(), "mockTimeTracker", "test", "test")
	time.Sleep(d)
}
