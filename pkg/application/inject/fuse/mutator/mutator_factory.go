package mutator

import (
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServerlessPlatform string

const (
	PlatformDefault      ServerlessPlatform = "Default"
	PlatformUnprivileged ServerlessPlatform = "Unprivileged"
)

func BuildMutator(ctx MutatingContext, client client.Client, log logr.Logger, platform string) (FluidObjectMutator, error) {
	switch ServerlessPlatform(platform) {
	case PlatformDefault:
		mutator := NewDefaultMutator(ctx, client, log)
		return mutator, nil
	case PlatformUnprivileged:
		mutator := NewUnprivilegedMutator(ctx, client, log)
		return mutator, nil
	default:
		return nil, fmt.Errorf("%s platform is not supported for fuse sidecar injection", platform)
	}
}
