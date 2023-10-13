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

package fuse

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/scripts/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// TODO: Mutator will be rewritten with polymorphism withe platform-specific mutation logic
type Mutator struct {
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

// prepareMutation makes preparations for the later mutation. For example, the preparations may include dependent
// resources creation(e.g. post start script) and fuse container template modifications.
func (mutator *Mutator) PrepareMutation() error {
	if !mutator.options.EnableCacheDir {
		mutator.transformTemplateWithCacheDirDisabled()
	}

	if mutator.options.EnableUnprivilegedSidecar {
		mutator.transformTemplateWithUnprivilegedSidecarEnabled()
	}

	if !mutator.options.SkipSidecarPostStartInject {
		if err := mutator.prepareFuseContainerPostStartScript(); err != nil {
			return err
		}
	}

	return nil
}

func (mutator *Mutator) MutateAndReturn() (*MutatingPodSpecs, error) {
	if err := mutator.mutateDatasetVolumes(); err != nil {
		return nil, err
	}

	if err := mutator.appendFuseContainerVolumes(); err != nil {
		return nil, err
	}

	used, err := mutator.ctx.GetDatsetUsedInContainers()
	if err != nil {
		return nil, err
	}

	if used {
		if err := mutator.prependFuseContainer(false /* asInit */); err != nil {
			return nil, err
		}
	}

	used, err = mutator.ctx.GetDatasetUsedInInitContainers()
	if err != nil {
		return nil, err
	}

	if used {
		if err := mutator.prependFuseContainer(true /* asInit */); err != nil {
			return nil, err
		}
	}

	return mutator.Specs, nil
}

func (mutator *Mutator) mutateDatasetVolumes() error {
	volumes := mutator.Specs.Volumes

	mountPath := mutator.template.FuseMountInfo.MountPath
	if mutator.template.FuseMountInfo.SubPath != "" {
		mountPath = mountPath + "/" + mutator.template.FuseMountInfo.SubPath
	}

	mutatedDatasetVolume := corev1.Volume{
		Name: "",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: mountPath,
			},
		},
	}

	var overriddenVolumeNames []string
	for i, volume := range volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == mutator.pvcName {
			name := volume.Name
			volumes[i] = mutatedDatasetVolume
			volumes[i].Name = name
			overriddenVolumeNames = append(overriddenVolumeNames, name)
		}
	}

	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

	mutator.ctx.SetDatasetUsedInContainers(false)
	for _, container := range mutator.Specs.Containers {
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(overriddenVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
				mutator.ctx.SetDatasetUsedInContainers(true)
			}
		}
	}

	mutator.ctx.SetDatasetUsedInInitContainers(false)
	for _, container := range mutator.Specs.InitContainers {
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(overriddenVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
				mutator.ctx.SetDatasetUsedInInitContainers(true)
			}
		}
	}

	return nil
}

func (mutator *Mutator) appendFuseContainerVolumes() (err error) {
	// collect all volumes' names
	var (
		volumeNames  = []string{}
		volumesToAdd = mutator.template.VolumesToAdd
		nameSuffix   = mutator.nameSuffix
	)
	for _, volume := range mutator.Specs.Volumes {
		volumeNames = append(volumeNames, volume.Name)
	}

	// Append volumes
	ctxAppenedVolumeNames, err := mutator.ctx.GetAppendedVolumeNames()
	if err != nil {
		return err
	}
	if len(volumesToAdd) > 0 {
		mutator.log.V(1).Info("Before append volume", "original", mutator.Specs.Volumes)
		// volumes = append(volumes, template.VolumesToAdd...)
		for _, volumeToAdd := range volumesToAdd {
			// nameSuffix would be like: "-0", "-1", "-2", "-3", ...
			oldVolumeName := volumeToAdd.Name
			newVolumeName := volumeToAdd.Name + nameSuffix
			if utils.ContainsString(volumeNames, newVolumeName) {
				newVolumeName, err = randomizeNewVolumeName(newVolumeName, volumeNames)
				if err != nil {
					return err
				}
			}
			volumeToAdd.Name = newVolumeName
			volumeNames = append(volumeNames, newVolumeName)
			mutator.Specs.Volumes = append(mutator.Specs.Volumes, volumeToAdd)
			if oldVolumeName != newVolumeName {
				ctxAppenedVolumeNames[oldVolumeName] = newVolumeName
			}
		}

		mutator.log.V(1).Info("After append volume", "original", mutator.Specs.Volumes)
	}
	mutator.ctx.SetAppendedVolumeNames(ctxAppenedVolumeNames)

	return nil
}

func (mutator *Mutator) prependFuseContainer(asInit bool) error {
	fuseContainer := mutator.template.FuseContainer
	if !asInit {
		fuseContainer.Name = common.FuseContainerName + mutator.nameSuffix
	} else {
		fuseContainer.Name = common.InitFuseContainerName + mutator.nameSuffix
	}

	if asInit {
		fuseContainer.Lifecycle = nil
		fuseContainer.Command = []string{"sleep"}
		fuseContainer.Args = []string{"2s"}
	}

	ctxAppenedVolumeNames, err := mutator.ctx.GetAppendedVolumeNames()
	if err != nil {
		return err
	}
	for oldName, newName := range ctxAppenedVolumeNames {
		for i, volumeMount := range fuseContainer.VolumeMounts {
			if volumeMount.Name == oldName {
				fuseContainer.VolumeMounts[i].Name = newName
			}
		}
	}

	if !asInit {
		mutator.Specs.Containers = append([]corev1.Container{fuseContainer}, mutator.Specs.Containers...)
	} else {
		mutator.Specs.InitContainers = append([]corev1.Container{fuseContainer}, mutator.Specs.InitContainers...)
	}
	return nil
}

func (mutator *Mutator) prepareFuseContainerPostStartScript() error {
	// 4. inject the post start script for fuse container, if configmap doesn't exist, try to create it.
	// Post start script varies according to privileged or unprivileged sidecar.
	var (
		info             = mutator.runtimeInfo
		template         = mutator.template
		datasetName      = info.GetName()
		datasetNamespace = info.GetNamespace()
	)

	dataset, err := utils.GetDataset(mutator.client, datasetName, datasetNamespace)
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
	if !mutator.options.EnableUnprivilegedSidecar {
		volumeMountInContainer, err := kubeclient.GetFuseMountInContainer(template.FuseMountInfo.FsType, template.FuseContainer)
		if err != nil {
			return err
		}
		mountPathInContainer = volumeMountInContainer.MountPath
	}

	// Fluid assumes pvc name is the same with runtime's name
	gen := poststart.NewGenerator(types.NamespacedName{
		Name:      datasetName,
		Namespace: datasetNamespace,
	}, mountPathInContainer, template.FuseMountInfo.FsType, template.FuseMountInfo.SubPath, mutator.options)
	cm := gen.BuildConfigmap(ownerReference)
	found, err := kubeclient.IsConfigMapExist(mutator.client, cm.Name, cm.Namespace)
	if err != nil {
		return err
	}

	if !found {
		err = mutator.client.Create(context.TODO(), cm)
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

func (mutator *Mutator) transformTemplateWithUnprivilegedSidecarEnabled() {
	// remove the fuse related volumes if using virtual fuse device
	template := mutator.template
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostMountNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostMountNames)

	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, hostFuseDeviceNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, hostFuseDeviceNames)

	// add virtual fuse device resource
	if template.FuseContainer.Resources.Limits == nil {
		template.FuseContainer.Resources.Limits = map[corev1.ResourceName]resource.Quantity{}
	}
	template.FuseContainer.Resources.Limits[corev1.ResourceName(fuseDeviceResourceName)] = resource.MustParse("1")

	if template.FuseContainer.Resources.Requests == nil {
		template.FuseContainer.Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
	}
	template.FuseContainer.Resources.Requests[corev1.ResourceName(fuseDeviceResourceName)] = resource.MustParse("1")

	// invalidate privileged fuse container
	if template.FuseContainer.SecurityContext != nil {
		privilegedContainer := false
		template.FuseContainer.SecurityContext.Privileged = &privilegedContainer
		if template.FuseContainer.SecurityContext.Capabilities != nil {
			template.FuseContainer.SecurityContext.Capabilities.Add = utils.TrimCapabilities(template.FuseContainer.SecurityContext.Capabilities.Add, []string{"SYS_ADMIN"})
		}
	}
}

func (mutator *Mutator) transformTemplateWithCacheDirDisabled() {
	template := mutator.template
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, cacheDirNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, cacheDirNames)
}

func randomizeNewVolumeName(origName string, existingNames []string) (string, error) {
	i := 0
	newVolumeName := utils.ReplacePrefix(origName, common.Fluid)
	for {
		if !utils.ContainsString(existingNames, newVolumeName) {
			break
		} else {
			if i > 100 {
				return "", fmt.Errorf("retry  the volume name %v because duplicate name more than 100 times, then give up", newVolumeName)
			}
			suffix := common.Fluid + "-" + utils.RandomAlphaNumberString(3)
			newVolumeName = utils.ReplacePrefix(origName, suffix)
			i++
		}
	}

	return newVolumeName, nil
}
