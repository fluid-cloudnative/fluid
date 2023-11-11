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
