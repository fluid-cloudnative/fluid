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
