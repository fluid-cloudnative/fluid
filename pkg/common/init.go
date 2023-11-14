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
	runtimeOpts "github.com/fluid-cloudnative/fluid/pkg/utils/runtimes/options"
)

const (
	DefaultInitImage = "registry.cn-hangzhou.aliyuncs.com/fluid/init-users:v0.3.0-1467caa"
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

func PortCheckEnabled() bool {
	return runtimeOpts.PortCheckEnabled()
}
