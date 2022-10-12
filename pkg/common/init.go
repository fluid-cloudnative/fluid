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
	ImageInfo      `json:",inline"`
	EnvUsers       string `json:"envUsers"`
	Dir            string `json:"dir"`
	Enabled        bool   `json:"enabled,omitempty"`
	EnvTieredPaths string `json:"envTieredPaths"`
}

// InitPortCheck defines a init container reports port status usage
type InitPortCheck struct {
	ImageInfo    `json:",inline"`
	Enabled      bool   `json:"enabled,omitempty"`
	PortsToCheck string `json:"portsToCheck,omitempty"`
}

func PortCheckEnabled() bool {
	return runtimeOpts.PortCheckEnabled()
}
