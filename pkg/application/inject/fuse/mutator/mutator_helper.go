package mutator

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type helperData struct {
	pvcName     string
	template    *common.FuseInjectionTemplate
	options     common.FuseSidecarInjectOption
	runtimeInfo base.RuntimeInfoInterface
	nameSuffix  string

	client client.Client
	log    logr.Logger

	Specs *MutatingPodSpecs
	ctx   mutatingContext
}

type mutatorHelper struct {
	prepareMutationFn            func(*helperData) error
	mutateDatasetVolumesFn       func(*helperData) error
	appendFuseContainerVolumesFn func(*helperData) error
	injectFuseContainerFn        func(*helperData) error
}

// doMutate organized core logic when mutating PodSpec. It takes the following steps:
// 1. prepare for mutation, where fuse container template will be transformed according to the serverless platform
// 2. mutate dataset volumes in the PodSpec, where user-specified volumes(usually a Fluid PVC volume) will be converted to a HostPath or EmptyDir.
// 3. append fuse container volumes to the PodSpec, where the volumes of the fuse container will be appended to the PodSpec
// 4. inject fuse container to the PodSpec, and fuse container will run as a sidecar container.
func (mh *mutatorHelper) doMutate(helperData *helperData) error {
	if mh.prepareMutationFn == nil ||
		mh.mutateDatasetVolumesFn == nil ||
		mh.appendFuseContainerVolumesFn == nil ||
		mh.injectFuseContainerFn == nil {
		return fmt.Errorf("mutatorHelper: helper functions cannot be nil, please recheck mutation logic")
	}

	if err := mh.prepareMutationFn(helperData); err != nil {
		return errors.Wrapf(err, "failed to prepare mutation for runtime \"%s/%s\"", helperData.runtimeInfo.GetNamespace(), helperData.runtimeInfo.GetName())
	}

	if err := mh.mutateDatasetVolumesFn(helperData); err != nil {
		return errors.Wrapf(err, "failed to mutate dataset volumes for runtime \"%s/%s\"", helperData.runtimeInfo.GetNamespace(), helperData.runtimeInfo.GetName())
	}

	if err := mh.appendFuseContainerVolumesFn(helperData); err != nil {
		return errors.Wrapf(err, "failed to append fuse container volumes for runtime \"%s/%s\"", helperData.runtimeInfo.GetNamespace(), helperData.runtimeInfo.GetName())
	}

	if err := mh.injectFuseContainerFn(helperData); err != nil {
		return errors.Wrapf(err, "failed to inject fuse container for runtime \"%s/%s\"", helperData.runtimeInfo.GetNamespace(), helperData.runtimeInfo.GetName())
	}

	return nil
}
