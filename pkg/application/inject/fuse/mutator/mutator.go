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
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Mutator is the fuse sidecar mutator for platform-specific mutation logic.
type Mutator interface {
	MutateWithRuntimeInfo(pvcName string, runtimeInfo base.RuntimeInfoInterface, nameSuffix string) error

	PostMutate() error

	GetMutatedPodSpecs() *MutatingPodSpecs
}

type MutatorBuildArgs struct {
	Client    client.Client
	Log       logr.Logger
	Specs     *MutatingPodSpecs
	Options   common.FuseSidecarInjectOption
	ExtraArgs map[string]string
}

func (args MutatorBuildArgs) String() string {
	return fmt.Sprintf("{options: %v, extraArgs: %v}", args.Options, args.ExtraArgs)
}

var mutatorBuildFn map[string]func(MutatorBuildArgs) Mutator = map[string]func(MutatorBuildArgs) Mutator{
	utils.PlatformDefault:      NewDefaultMutator,
	utils.PlatformUnprivileged: NewUnprivilegedMutator,
}

func BuildMutator(args MutatorBuildArgs, platform string) (Mutator, error) {
	if fn, ok := mutatorBuildFn[platform]; ok {
		return fn(args), nil
	}

	return nil, fmt.Errorf("fuse sidecar mutator cannot be found for platform %s", platform)
}

// FindExtraArgsFromMetadata tries to get extra build args for a given mutator from a metaObj.
// For any platform-specific mutator, its extra args should be key-values and defined in the format of "{platform}.fluid.io/{key}={value}" in metaObj.annotaions.
func FindExtraArgsFromMetadata(metaObj metav1.ObjectMeta, platform string) (extraArgs map[string]string) {
	if len(metaObj.Annotations) == 0 || len(platform) == 0 {
		return
	}

	extraArgs = make(map[string]string)
	platformPrefix := fmt.Sprintf("%s.%s", platform, common.LabelAnnotationPrefix)
	for key, value := range metaObj.Annotations {
		if strings.HasPrefix(key, platformPrefix) {
			extraArgs[strings.TrimPrefix(key, platformPrefix)] = value
		}
	}

	return
}
