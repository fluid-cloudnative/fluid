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

import "os"

const (
	engineTypeFromEnv = "JINDO_ENGINE_TYPE"

	jindoEngine = "jindo"

	jindofsxEngine = "jindofsx"

	jindocacheEngine = "jindocache"

	defaultJindofsxRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:4.6.8"

	defaultJindofsRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.8.0"

	defaultJindoCacheRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:5.0.0"
)

// GetRuntimeType gets the runtime type for Jindo
func GetRuntimeType() (engine string) {
	engine = jindofsxEngine
	if env := os.Getenv(engineTypeFromEnv); env == jindoEngine || env == jindocacheEngine {
		engine = env
	}
	return
}

// GetRuntimeImage gets the runtime of Jindo
func GetRuntimeImage() (image string) {
	if GetRuntimeType() == jindofsxEngine {
		image = defaultJindofsxRuntimeImage
	} else if GetRuntimeType() == jindoEngine {
		image = defaultJindofsRuntimeImage
	} else if GetRuntimeType() == jindocacheEngine {
		image = defaultJindoCacheRuntimeImage
	}
	return
}
