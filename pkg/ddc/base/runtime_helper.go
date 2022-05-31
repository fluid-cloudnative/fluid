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
	"context"
	"fmt"
<<<<<<< HEAD
	"time"
=======
	"k8s.io/apimachinery/pkg/api/resource"
>>>>>>> 2fa66f87 (Support webhook mutation with fuse virtual device enabled)

	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
)

var (
	// datavolume-, volume-localtime for JindoFS
	// mem, ssd, hdd for Alluxio and GooseFS
	// cache-dir for JuiceFS
	cacheDirNames = []string{"datavolume-", "volume-localtime", "cache-dir", "mem", "ssd", "hdd"}

	// hostpath fuse mount point for Alluxio, JindoFS, GooseFS and JuiceFS
	hostMountNames = []string{"alluxio-fuse-mount", "jindofs-fuse-mount", "goosefs-fuse-mount", "juicefs-fuse-mount"}

	// fuse devices for Alluxio, JindoFS, GooseFS
	hostFuseDeviceNames = []string{"alluxio-fuse-device", "jindofs-fuse-device", "goosefs-fuse-device"}
)

// GetTemplateToInjectForFuse gets template for fuse injection
func (info *RuntimeInfo) GetTemplateToInjectForFuse(pvcName string, enableCacheDir bool) (template *common.FuseInjectionTemplate, err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "RuntimeInfo.GetTemplateToInjectForFuse",
			"pvc.name", pvcName, "pvc.namespace", info.GetNamespace())
	}
	// TODO: create fuse container
	ds, err := info.getFuseDaemonset()
	if err != nil {
		return template, err
	}

	dataset, err := utils.GetDataset(info.client, info.name, info.namespace)
	if err != nil {
		return template, err
	}

	ownerReference := metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}

	// 0. remove the cache dir if required
	if len(ds.Spec.Template.Spec.Containers) != 1 {
		return template, fmt.Errorf("the length of containers of fuse %s in namespace %s is not 1", ds.Name, ds.Namespace)
	}
	if !enableCacheDir {
		ds.Spec.Template.Spec.Containers[0].VolumeMounts = utils.TrimVolumeMounts(ds.Spec.Template.Spec.Containers[0].VolumeMounts, cacheDirNames)
		ds.Spec.Template.Spec.Volumes = utils.TrimVolumes(ds.Spec.Template.Spec.Volumes, cacheDirNames)
	}

	// 1. setup fuse sidecar container when enabling virtual fuse device
	if enableVirtFuseDev {
		// remove the fuse related volumes if using virtual fuse device
		ds.Spec.Template.Spec.Containers[0].VolumeMounts = utils.TrimVolumeMounts(ds.Spec.Template.Spec.Containers[0].VolumeMounts, hostMountNames)
		ds.Spec.Template.Spec.Volumes = utils.TrimVolumes(ds.Spec.Template.Spec.Volumes, hostMountNames)

		ds.Spec.Template.Spec.Containers[0].VolumeMounts = utils.TrimVolumeMounts(ds.Spec.Template.Spec.Containers[0].VolumeMounts, hostFuseDeviceNames)
		ds.Spec.Template.Spec.Volumes = utils.TrimVolumes(ds.Spec.Template.Spec.Volumes, hostFuseDeviceNames)

		// add virtual fuse device resource
		if ds.Spec.Template.Spec.Containers[0].Resources.Limits == nil {
			ds.Spec.Template.Spec.Containers[0].Resources.Limits = map[corev1.ResourceName]resource.Quantity{}
		}
		ds.Spec.Template.Spec.Containers[0].Resources.Limits["fluid.io/fuse"] = resource.MustParse("1")

		if ds.Spec.Template.Spec.Containers[0].Resources.Requests == nil {
			ds.Spec.Template.Spec.Containers[0].Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
		}
		ds.Spec.Template.Spec.Containers[0].Resources.Requests["fluid.io/fuse"] = resource.MustParse("1")

		// invalidate privileged fuse container
		privilegedContainer := false
		ds.Spec.Template.Spec.Containers[0].SecurityContext.Privileged = &privilegedContainer
		ds.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities.Add = utils.TrimCapabilities(ds.Spec.Template.Spec.Containers[0].SecurityContext.Capabilities.Add, []string{"SYS_ADMIN"})
	}

	// 2. set the pvc name
	template = &common.FuseInjectionTemplate{
		PVCName: pvcName,
	}

	// 3. set the fuse container
	template.FuseContainer = ds.Spec.Template.Spec.Containers[0]
	template.FuseContainer.Name = common.FuseContainerName

	// template.VolumesToAdd = ds.Spec.Template.Spec.Volumes
	// 4. inject the post start script for fuse container, if configmap doesn't exist, try to create it.
	// Post start script varies according to privileged or unprivileged sidecar.
	mountPath, mountType, err := kubeclient.GetMountInfoFromVolumeClaim(info.client, pvcName, info.namespace)
	if err != nil {
		return
	}

	mountPathInContainer := ""
	if !enableVirtFuseDev {
		volumeMountInContainer, err := kubeclient.GetFuseMountInContainer(mountType, template.FuseContainer)
		if err != nil {
			return template, err
		}
		mountPathInContainer = volumeMountInContainer.MountPath
	}

	gen := poststart.NewGenerator(types.NamespacedName{
		Name:      info.name,
		Namespace: info.namespace,
	}, mountPathInContainer, mountType, enableVirtFuseDev)
	cm := gen.BuildConfigmap(ownerReference)
	found, err := kubeclient.IsConfigMapExist(info.client, cm.Name, cm.Namespace)
	if err != nil {
		return template, err
	}

	if !found {
		err = info.client.Create(context.TODO(), cm)
		if err != nil {
			return template, err
		}
	}

	template.FuseContainer.VolumeMounts = append(template.FuseContainer.VolumeMounts, gen.GetVolumeMount())
	if template.FuseContainer.Lifecycle == nil {
		template.FuseContainer.Lifecycle = &corev1.Lifecycle{}
	}
	template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand()

	// 5. create a volume with pvcName with mountpath in pv, and add it to VolumesToUpdate
	template.VolumesToUpdate = []corev1.Volume{
		{
			Name: pvcName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: mountPath,
				},
			},
		},
	}
	template.VolumesToAdd = append(ds.Spec.Template.Spec.Volumes, gen.GetVolume())

	return
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
