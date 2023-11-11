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
