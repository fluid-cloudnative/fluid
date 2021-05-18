package utils

import (
	"fmt"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
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
	timeLog.Info(fmt.Sprintf("%s took %s", processName, elpased), keysAndValues...)
}
