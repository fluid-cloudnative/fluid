/*
Copyright 2025 The Fluid Authors.

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

package options

import (
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const (
	EnvControllerSkipSyncingRuntime = "CONTROLLER_SKIP_SYNCING_RUNTIME"
)

var controllerSkipSyncingRuntime bool

func setControllerSkipSyncingRuntime() {
	if strVal, found := os.LookupEnv(EnvControllerSkipSyncingRuntime); found {
		boolVal, err := strconv.ParseBool(strVal)
		if err != nil {
			panic(errors.Wrapf(err, "can't parse env %s to bool", EnvControllerSkipSyncingRuntime))
		}
		controllerSkipSyncingRuntime = boolVal
	}

	log.Info("ControllerSkipSyncingRuntime", "value", controllerSkipSyncingRuntime)
}

func ShouldSkipSyncingRuntime() bool {
	return controllerSkipSyncingRuntime
}
