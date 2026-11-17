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
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// SumQuotas parses a comma-separated quota list such as "10Gi,5Gi" and returns the total.
func SumQuotas(quotas string) (total resource.Quantity, err error) {
	for _, q := range strings.Split(quotas, ",") {
		parsed, parseErr := resource.ParseQuantity(strings.TrimSpace(q))
		if parseErr != nil {
			return total, fmt.Errorf("invalid quota %q in %q: %w", q, quotas, parseErr)
		}
		total.Add(parsed)
	}
	return total, nil
}

const (
	engineTypeFromEnv = "JINDO_ENGINE_TYPE"

	defaultJindofsRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.8.0"

	defaultJindofsxRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:4.6.8"

	defaultJindoCacheRuntimeImage = "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:6.2.0"
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
