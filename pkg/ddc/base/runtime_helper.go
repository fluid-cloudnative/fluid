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

package base

import (
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// GetFuseContainerTemplate collects the fuse container spec from the runtime's fuse daemonSet spec. The function summarizes fuse related information into
// the template and returns it. The template then can be freely modified according to need of the serverless platform.
func (info *RuntimeInfo) GetFuseContainerTemplate() (template *common.FuseInjectionTemplate, err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "RuntimeInfo.GetFuseContainerTemplate",
			"runtime.name", info.name, "runtime.namespace", info.namespace)
	}

	ds, err := info.getFuseDaemonset()
	if err != nil {
		return template, err
	}

	if len(ds.Spec.Template.Spec.Containers) <= 0 {
		return template, fmt.Errorf("the length of containers of fuse daemonset \"%s/%s\" should not be 0", ds.Namespace, ds.Name)
	}

	template = &common.FuseInjectionTemplate{
		FuseContainer: ds.Spec.Template.Spec.Containers[0],
		VolumesToAdd:  utils.FilterVolumesByVolumeMounts(ds.Spec.Template.Spec.Volumes, ds.Spec.Template.Spec.Containers[0].VolumeMounts),
	}

	template.FuseContainer.Name = common.FuseContainerName

	hostMountPath, mountType, subPath, err := kubeclient.GetMountInfoFromVolumeClaim(info.client, info.name, info.namespace)
	if err != nil {
		return template, errors.Wrapf(err, "failed get mount info from PVC \"%s/%s\"", info.namespace, info.name)
	}

	fuseVolMount, err := kubeclient.GetFuseMountInContainer(mountType, ds.Spec.Template.Spec.Containers[0])
	if err != nil {
		return template, errors.Wrapf(err, "failed to get fuse volume mount from container")
	}

	template.FuseMountInfo = common.FuseMountInfo{
		FsType:             mountType,
		HostMountPath:      hostMountPath,
		ContainerMountPath: fuseVolMount.MountPath,
		SubPath:            subPath,
	}

	return template, nil
}

func (info *RuntimeInfo) getFuseDaemonset() (ds *appsv1.DaemonSet, err error) {
	if info.client == nil {
		err = fmt.Errorf("client is not set")
		return
	}

	var fuseName string
	switch info.runtimeType {
	case common.JindoRuntime:
		fuseName = info.name + "-" + common.JindoChartName + "-fuse"
	default:
		fuseName = info.name + "-fuse"
	}
	return kubeclient.GetDaemonset(info.client, fuseName, info.GetNamespace())
}
