package mutator

import "github.com/fluid-cloudnative/fluid/pkg/common"

type FluidObjectMutator interface {
	Mutate(pod *common.FluidObject, template common.FuseInjectionTemplate, options common.FuseSidecarInjectOption) error
}
