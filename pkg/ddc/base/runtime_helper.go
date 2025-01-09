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

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
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

	hostMountPath, mountType, subPath, err := info.getMountInfo()
	if err != nil {
		return template, errors.Wrapf(err, "failed to get mount info by runtimeInfo %s/%s", info.namespace, info.name)
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

func (info *RuntimeInfo) getMountInfo() (path, mountType, subpath string, err error) {
	pv, err := kubeclient.GetPersistentVolume(info.client, info.GetPersistentVolumeName())
	if err != nil {
		err = errors.Wrapf(err, "cannot find pvc \"%s/%s\"'s bounded PV", info.namespace, info.name)
		return
	}

	if pv.Spec.CSI != nil && len(pv.Spec.CSI.VolumeAttributes) > 0 {
		path = pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidPath]
		mountType = pv.Spec.CSI.VolumeAttributes[common.VolumeAttrMountType]
		subpath = pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath]
	} else {
		err = fmt.Errorf("the pv %s is not created by fluid", pv.Name)
	}

	return
}
