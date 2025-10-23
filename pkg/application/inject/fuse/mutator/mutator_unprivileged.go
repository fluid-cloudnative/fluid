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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

type UnprivilegedMutator struct {
	// UnprivilegedMutator inherits from DefaultMutator
	DefaultMutator
}

var _ Mutator = &UnprivilegedMutator{}

func NewUnprivilegedMutator(opts MutatorBuildArgs) Mutator {
	var prepareMutationFn = unprivilegedPrepareMutation
	var mutateDatasetVolumesFn = defaultMutateDatasetVolumes
	var appendFuseContainerVolumesFn = defaultAppendFuseContainerVolumes
	var injectFuseContainerFn = defaultInjectFuseContainer
	if opts.Options.SidecarInjectionMode == common.SidecarInjectionMode_NativeSidecar {
		injectFuseContainerFn = defaultInjectFuseNativeSidecar
	}

	return &UnprivilegedMutator{
		DefaultMutator: DefaultMutator{
			options: opts.Options,
			client:  opts.Client,
			log:     opts.Log,
			Specs:   opts.Specs,

			mutatorHelper: mutatorHelper{
				prepareMutationFn:            prepareMutationFn,
				mutateDatasetVolumesFn:       mutateDatasetVolumesFn,
				appendFuseContainerVolumesFn: appendFuseContainerVolumesFn,
				injectFuseContainerFn:        injectFuseContainerFn,
			},
		},
	}
}

// unprivilegedPrepareMutation extends the func defaultPrepareMutations and insert transformation logic for a unprivileged sidecar
func unprivilegedPrepareMutation(helper *helperData) error {
	if !helper.options.EnableCacheDir {
		transformTemplateWithCacheDirDisabled(helper)
	}

	transformTemplateWithUnprivilegedSidecarEnabled(helper)

	if !helper.options.SkipSidecarPostStartInject {
		if err := prepareFuseContainerPostStartScript(helper); err != nil {
			return err
		}
	}

	if !helper.runtimeInfo.GetFuseMetricsScrapeTarget().Selected(base.SidecarMountMode) {
		removeFuseMetricsContainerPort(helper)
	}

	return nil
}

func transformTemplateWithUnprivilegedSidecarEnabled(helper *helperData) {
	// remove the fuse related volumes if using virtual fuse device
	template := helper.template
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostMountNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostMountNames)

	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostFuseDeviceNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostFuseDeviceNames)

	// invalidate privileged fuse container
	if template.FuseContainer.SecurityContext != nil {
		privilegedContainer := false
		template.FuseContainer.SecurityContext.Privileged = &privilegedContainer
		if template.FuseContainer.SecurityContext.Capabilities != nil {
			template.FuseContainer.SecurityContext.Capabilities.Add = utils.TrimCapabilities(template.FuseContainer.SecurityContext.Capabilities.Add, []string{"SYS_ADMIN"})
		}
	}
}
