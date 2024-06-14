/*
Copyright 2023 The Fluid Authors.

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

package jindo

import (
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

const (
	engineTypeFromEnv = "JINDO_ENGINE_TYPE"

	defaultJindofsRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.8.0"

	defaultJindofsxRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:4.6.8"

	defaultJindoCacheRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:6.4.0"
)

// GetDefaultEngineImpl gets the runtime type for Jindo
func GetDefaultEngineImpl() (engine string) {
	engine = common.JindoCacheEngineImpl
	if env := os.Getenv(engineTypeFromEnv); env == common.JindoFSEngineImpl || env == common.JindoFSxEngineImpl {
		engine = env
	}
	return
}

// GetRuntimeImage gets the runtime of Jindo
func GetRuntimeImage() (image string) {
	if GetDefaultEngineImpl() == common.JindoFSxEngineImpl {
		image = defaultJindofsxRuntimeImage
	} else if GetDefaultEngineImpl() == common.JindoFSEngineImpl {
		image = defaultJindofsRuntimeImage
	} else if GetDefaultEngineImpl() == common.JindoCacheEngineImpl {
		image = defaultJindoCacheRuntimeImage
	}
	return
}
