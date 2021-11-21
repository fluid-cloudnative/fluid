package utils

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var timeLog logr.Logger

func init() {
	timeLog = ctrl.Log.WithName("utils")
}

// TimeTrack tracks the time cost for some process with some optional information.
// For example, to track the processing time of a function, just add the following code
// at the beginning of your function:
//   defer utils.TimeTrack(time.Now(), <func-name>, <keysAndValues>...)
func TimeTrack(start time.Time, processName string, keysAndValues ...interface{}) {
	elpased := time.Since(start)
	if checkLongTask(elpased) {
		timeLog.Info(fmt.Sprintf("Warning: %s took %s, it's a long task.", processName, elpased), keysAndValues...)
	} else {
		timeLog.V(1).Info(fmt.Sprintf("%s took %s", processName, elpased), keysAndValues...)
	}
}

// checkLongTask checks the time conusmes
func checkLongTask(elpased time.Duration) bool {
	var threshold time.Duration = 2 * time.Second
	return elpased >= threshold
}
