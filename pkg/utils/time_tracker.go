/*
Copyright 2022 The Fluid Authors.

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
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// the default task elapsed
var (
	taskTimeThreshold    time.Duration = 30 * time.Second
	timeLog              logr.Logger
	enableTimeTrack      bool
	enableTimeTrackDebug bool
)

func init() {
	timeLog = ctrl.Log.WithName("utils.time")
	enableTimeTrack = GetBoolValueFromEnv(common.EnvTimeTrack, false)
	enableTimeTrackDebug = GetBoolValueFromEnv(common.EnvTimeTrackDebug, false)
}

// TimeTrack tracks the time cost for some process with some optional information.
// For example, to track the processing time of a function, just add the following code
// at the beginning of your function:
//
//	defer utils.TimeTrack(time.Now(), <func-name>, <keysAndValues>...)
func TimeTrack(start time.Time, processName string, keysAndValues ...interface{}) {
	elpased := time.Since(start)
	if IsTimeTrackerEnabled() {
		timeLog.Info(fmt.Sprintf("%s took %s", processName, elpased), keysAndValues...)
	} else if checkLongTask(elpased) {
		timeLog.Info(fmt.Sprintf("Warning: %s took %s , it's a long task.", processName, elpased), keysAndValues...)
	} else {
		timeLog.V(1).Info(fmt.Sprintf("%s took %s", processName, elpased), keysAndValues...)
	}
}

// checkLongTask checks the time conusmes
func checkLongTask(elpased time.Duration) bool {
	return elpased >= taskTimeThreshold
}

func IsTimeTrackerEnabled() bool {
	return enableTimeTrack || enableTimeTrackDebug
}

func IsTimeTrackerDebugEnabled() bool {
	return enableTimeTrackDebug
}
