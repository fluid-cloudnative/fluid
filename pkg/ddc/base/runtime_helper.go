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
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
)

var fuseDeviceResourceName string

var (
	// datavolume-, volume-localtime for JindoFS
	// mem, ssd, hdd for Alluxio and GooseFS
	// cache-dir for JuiceFS
	cacheDirNames = []string{"datavolume-", "volume-localtime", "cache-dir", "mem", "ssd", "hdd"}

	// hostpath fuse mount point for Alluxio, JindoFS, GooseFS and JuiceFS
	hostMountNames = []string{"alluxio-fuse-mount", "jindofs-fuse-mount", "goosefs-fuse-mount", "juicefs-fuse-mount", "thin-fuse-mount", "efc-fuse-mount", "efc-sock"}

	// fuse devices for Alluxio, JindoFS, GooseFS
	hostFuseDeviceNames = []string{"alluxio-fuse-device", "jindofs-fuse-device", "goosefs-fuse-device", "thin-fuse-device"}
)

func init() {
	fuseDeviceResourceName = utils.GetStringValueFromEnv(common.EnvFuseDeviceResourceName, common.DefaultFuseDeviceResourceName)
}

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

	mountPath, mountType, subPath, err := kubeclient.GetMountInfoFromVolumeClaim(info.client, info.name, info.namespace)
	if err != nil {
		return template, errors.Wrapf(err, "failed get mount info from PVC \"%s/%s\"", info.namespace, info.name)
	}

	template.FuseMountInfo = common.FuseMountInfo{
		MountPath: mountPath,
		FsType:    mountType,
		SubPath:   subPath,
	}

	return template, nil
}

// GetTemplateToInjectForFuse gets template for fuse injection
func (info *RuntimeInfo) GetTemplateToInjectForFuse(pvcName string, pvcNamespace string, option common.FuseSidecarInjectOption) (template *common.FuseInjectionTemplate, err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "RuntimeInfo.GetTemplateToInjectForFuse",
			"pvc.name", pvcName, "pvc.namespace", pvcNamespace)
	}
	// TODO: create fuse container
	ds, err := info.getFuseDaemonset()
	if err != nil {
		return template, err
	}

	if len(ds.Spec.Template.Spec.Containers) <= 0 {
		return template, fmt.Errorf("the length of containers of fuse daemonset \"%s/%s\" should not be 0", ds.Namespace, ds.Name)
	}

	// 1. set the pvc name
	template = &common.FuseInjectionTemplate{
		PVCName:       pvcName,
		FuseContainer: ds.Spec.Template.Spec.Containers[0],
		// only add volumes that the Fuse container needs
		VolumesToAdd: utils.FilterVolumesByVolumeMounts(ds.Spec.Template.Spec.Volumes, ds.Spec.Template.Spec.Containers[0].VolumeMounts),
	}

	// 2. Inject cache dir to enable short-circuit read if needed
	if !option.EnableCacheDir {
		info.transformTemplateWithCacheDirDisabled(template)
	}

	// 3. Transform fuse sidecar container when injecting an unprivileged sidecar
	if option.EnableUnprivilegedSidecar {
		info.transformTemplateWithUnprivilegedSidecarEnabled(template)
	}

	// 4. set the fuse container name
	template.FuseContainer.Name = common.FuseContainerName

	// get the pv attribute, mountPath is with prefix "/runtime-mnt/..."
	mountPath, mountType, subPath, err := kubeclient.GetMountInfoFromVolumeClaim(info.client, info.name, info.namespace)
	if err != nil {
		return template, err
	}

	// 5. Inject FUSE sidecar post start script, script varies according to privileged or unprivileged sidecar.
	if !option.SkipSidecarPostStartInject {
		if err := info.injectFuseContainerPostStartScript(template, mountType, subPath, option); err != nil {
			return template, err
		}
	}

	// 6. Update PVC Volume to HostPath
	if subPath != "" {
		mountPath = mountPath + "/" + subPath
	}
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

func (info *RuntimeInfo) transformTemplateWithUnprivilegedSidecarEnabled(template *common.FuseInjectionTemplate) {
	// remove the fuse related volumes if using virtual fuse device
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostMountNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostMountNames)

	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostFuseDeviceNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostFuseDeviceNames)

	// add virtual fuse device resource
	if template.FuseContainer.Resources.Limits == nil {
		template.FuseContainer.Resources.Limits = map[corev1.ResourceName]resource.Quantity{}
	}
	template.FuseContainer.Resources.Limits[corev1.ResourceName(getFuseDeviceResourceName())] = resource.MustParse("1")

	if template.FuseContainer.Resources.Requests == nil {
		template.FuseContainer.Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
	}
	template.FuseContainer.Resources.Requests[corev1.ResourceName(getFuseDeviceResourceName())] = resource.MustParse("1")

	// invalidate privileged fuse container
	if template.FuseContainer.SecurityContext != nil {
		privilegedContainer := false
		template.FuseContainer.SecurityContext.Privileged = &privilegedContainer
		if template.FuseContainer.SecurityContext.Capabilities != nil {
			template.FuseContainer.SecurityContext.Capabilities.Add = utils.TrimCapabilities(template.FuseContainer.SecurityContext.Capabilities.Add, []string{"SYS_ADMIN"})
		}
	}
}

func (info *RuntimeInfo) transformTemplateWithCacheDirDisabled(template *common.FuseInjectionTemplate) {
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, cacheDirNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, cacheDirNames)
}

func (info *RuntimeInfo) injectFuseContainerPostStartScript(template *common.FuseInjectionTemplate, mountType string, subPath string, option common.FuseSidecarInjectOption) error {
	// 4. inject the post start script for fuse container, if configmap doesn't exist, try to create it.
	// Post start script varies according to privileged or unprivileged sidecar.

	dataset, err := utils.GetDataset(info.client, info.name, info.namespace)
	if err != nil {
		return err
	}

	ownerReference := metav1.OwnerReference{
		APIVersion: dataset.APIVersion,
		Kind:       dataset.Kind,
		Name:       dataset.Name,
		UID:        dataset.UID,
	}

	// the mountPathInContainer is the parent dir of fuse mount path in the container
	mountPathInContainer := ""
	if !option.EnableUnprivilegedSidecar {
		volumeMountInContainer, err := kubeclient.GetFuseMountInContainer(mountType, template.FuseContainer)
		if err != nil {
			return err
		}
		mountPathInContainer = volumeMountInContainer.MountPath
	}

	// Fluid assumes pvc name is the same with runtime's name
	gen := poststart.NewGenerator(types.NamespacedName{
		Name:      info.name,
		Namespace: info.namespace,
	}, mountPathInContainer, mountType, subPath, option)
	cm := gen.BuildConfigmap(ownerReference)
	found, err := kubeclient.IsConfigMapExist(info.client, cm.Name, cm.Namespace)
	if err != nil {
		return err
	}

	if !found {
		err = info.client.Create(context.TODO(), cm)
		if err != nil {
			// If ConfigMap creation succeeds concurrently, continue to mutate
			if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
				return err
			}
		}
	}

	template.FuseContainer.VolumeMounts = append(template.FuseContainer.VolumeMounts, gen.GetVolumeMount())
	if template.FuseContainer.Lifecycle == nil {
		template.FuseContainer.Lifecycle = &corev1.Lifecycle{}
	}
	template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand()
	template.VolumesToAdd = append(template.VolumesToAdd, gen.GetVolume())

	return nil
}

func getFuseDeviceResourceName() string {
	return fuseDeviceResourceName
}
