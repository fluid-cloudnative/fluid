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

package ctrl

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The common part of the engine which can be reused
type Helper struct {
	runtimeInfo base.RuntimeInfoInterface

	client client.Client
}

func BuildHelper(runtime base.RuntimeInfoInterface, client client.Client) *Helper {
	return &Helper{
		runtimeInfo: runtime,
		client:      client,
	}
}

// SetupWorkers builds workers affinity if it doesn't have
func (e *Helper) SetupWorkers(runtime client.Object) (err error) {

}
