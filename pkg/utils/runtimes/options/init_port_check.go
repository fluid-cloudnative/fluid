/*
Copyright 2021 The Fluid Authors.

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
	EnvPortCheckEnabled = "INIT_PORT_CHECK_ENABLED"
)

var initPortCheckEnabled = false

func setPortCheckOption() {
	if strVal, found := os.LookupEnv(EnvPortCheckEnabled); found {
		if boolVal, err := strconv.ParseBool(strVal); err != nil {
			panic(errors.Wrapf(err, "can't parse %s to bool", EnvPortCheckEnabled))
		} else {
			initPortCheckEnabled = boolVal
		}
	}
	// log.Printf("Using %s = %v\n", EnvPortCheckEnabled, initPortCheckEnabled)
	log.Info("setPortCheckOption", "EnvPortCheckEnabled",
		EnvPortCheckEnabled,
		"initPortCheckEnabled",
		initPortCheckEnabled)
}

func PortCheckEnabled() bool {
	return initPortCheckEnabled
}
