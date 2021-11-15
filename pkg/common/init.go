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

package common

import (
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
)

const (
	DefaultInitImage    = "registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.3.0-1467caa"
	EnvPortCheckEnabled = "INIT_PORT_CHECK_ENABLED"
)

// The InitContainer to init the users for other Containers
type InitUsers struct {
	ImageInfo      `yaml:",inline"`
	EnvUsers       string `yaml:"envUsers"`
	Dir            string `yaml:"dir"`
	Enabled        bool   `yaml:"enabled,omitempty"`
	EnvTieredPaths string `yaml:"envTieredPaths"`
}

// InitPortCheck defines a init container reports port status usage
type InitPortCheck struct {
	ImageInfo    `yaml:",inline"`
	Enabled      bool   `yaml:"enabled,omitempty"`
	PortsToCheck string `yaml:"portsToCheck,omitempty"`
}

var initPortCheckEnabled = false

func init() {
	if strVal, exist := os.LookupEnv(EnvPortCheckEnabled); exist {
		if boolVal, err := strconv.ParseBool(strVal); err != nil {
			panic(errors.Wrapf(err, "can't parse %s to bool", EnvPortCheckEnabled))
		} else {
			initPortCheckEnabled = boolVal
		}
	}
	log.Printf("Using %s = %v\n", EnvPortCheckEnabled, initPortCheckEnabled)
}

func PortCheckEnabled() bool {
	return initPortCheckEnabled
}
