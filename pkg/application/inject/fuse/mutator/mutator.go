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
