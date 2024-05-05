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

package mutator

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Mutator is the fuse sidecar mutator for platform-specific mutation logic.
type Mutator interface {
	MutateWithRuntimeInfo(pvcName string, runtimeInfo base.RuntimeInfoInterface, nameSuffix string) error

	PostMutate() error

	GetMutatedPodSpecs() *MutatingPodSpecs
}

type MutatorBuildOpts struct {
	Options common.FuseSidecarInjectOption
	Client  client.Client
	Log     logr.Logger
	Specs   *MutatingPodSpecs
}

var mutatorBuildFn map[string]func(MutatorBuildOpts) Mutator = map[string]func(MutatorBuildOpts) Mutator{
	utils.PlatformDefault:      NewDefaultMutator,
	utils.PlatformUnprivileged: NewUnprivilegedMutator,
}

func BuildMutator(opts MutatorBuildOpts, platform string) (Mutator, error) {
	if fn, ok := mutatorBuildFn[platform]; ok {
		return fn(opts), nil
	}

	return nil, fmt.Errorf("fuse sidecar mutator cannot be found for platform %s", platform)
}
