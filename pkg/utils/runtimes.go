/*

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

package utils

import (
	data "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
)

// NewRuntimeCondition creates a new Cache condition.
func NewRuntime(name, namespace string, category common.Category, runtimeType string) data.Runtime {
	return data.Runtime{
		Name:      name,
		Namespace: namespace,
		Category:  category,
		// Engine:    engine,
		Type: runtimeType,
	}
}

// AddRuntimesIfNotExist
func AddRuntimesIfNotExist(runtimes []data.Runtime, newRuntime data.Runtime) (updatedRuntimes []data.Runtime) {
	catergoryMap := map[common.Category]bool{}
	for _, runtime := range runtimes {
		catergoryMap[runtime.Category] = true
	}

	if _, found := catergoryMap[newRuntime.Category]; !found {
		updatedRuntimes = append(runtimes, newRuntime)
	} else {
		updatedRuntimes = runtimes
		log.V(1).Info("No need to add new runtime to dataset", "type", newRuntime.Category)
	}

	return updatedRuntimes
}
