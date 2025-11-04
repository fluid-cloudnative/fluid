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
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse/poststart"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

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

// TODO: DefaultMutator will be rewritten with polymorphism withe platform-specific mutation logic
type DefaultMutator struct {
	options common.FuseSidecarInjectOption
	client  client.Client
	log     logr.Logger
	Specs   *MutatingPodSpecs

	mutatorHelper mutatorHelper
}

func NewDefaultMutator(args MutatorBuildArgs) Mutator {
	var prepareMutationFn = defaultPrepareMutation
	var mutateDatasetVolumesFn = defaultMutateDatasetVolumes
	var appendFuseContainerVolumesFn = defaultAppendFuseContainerVolumes
	var injectFuseContainerFn = defaultInjectFuseContainer
	if args.Options.SidecarInjectionMode == common.SidecarInjectionMode_NativeSidecar {
		injectFuseContainerFn = defaultInjectFuseNativeSidecar
	}

	return &DefaultMutator{
		options: args.Options,
		client:  args.Client,
		log:     args.Log,
		Specs:   args.Specs,

		mutatorHelper: mutatorHelper{
			prepareMutationFn:            prepareMutationFn,
			mutateDatasetVolumesFn:       mutateDatasetVolumesFn,
			appendFuseContainerVolumesFn: appendFuseContainerVolumesFn,
			injectFuseContainerFn:        injectFuseContainerFn,
		},
	}
}

var _ Mutator = &DefaultMutator{}

func (mutator *DefaultMutator) MutateWithRuntimeInfo(pvcName string, runtimeInfo base.RuntimeInfoInterface, nameSuffix string) error {
	template, err := runtimeInfo.GetFuseContainerTemplate()
	if err != nil {
		return errors.Wrapf(err, "failed to get fuse container template for runtime \"%s/%s\"", runtimeInfo.GetNamespace(), runtimeInfo.GetName())
	}

	helperData := &helperData{
		pvcName:     pvcName,
		template:    template,
		options:     mutator.options,
		runtimeInfo: runtimeInfo,
		nameSuffix:  nameSuffix,
		client:      mutator.client,
		log:         mutator.log,
		Specs:       mutator.Specs,
		ctx:         mutatingContext{},
	}

	if err := mutator.mutatorHelper.doMutate(helperData); err != nil {
		return errors.Wrapf(err, "failed to mutate for runtime \"%s/%s\"", runtimeInfo.GetNamespace(), runtimeInfo.GetName())
	}

	if runtimeInfo.GetFuseMetricsScrapeTarget().Selected(base.SidecarMountMode) {
		enablePrometheusMetricsScrape(helperData)
	}

	return nil
}

func (mutator *DefaultMutator) GetMutatedPodSpecs() *MutatingPodSpecs {
	return mutator.Specs
}

func (mutator *DefaultMutator) PostMutate() error {
	return nil
}

// defaultPrepareMutation makes preparations for the later mutation. For example, the preparations may include dependent
// resources creation(e.g. post start script) and fuse container template modifications.
func defaultPrepareMutation(helper *helperData) error {
	if !helper.options.EnableCacheDir {
		transformTemplateWithCacheDirDisabled(helper)
	}

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

func defaultMutateDatasetVolumes(helper *helperData) (err error) {
	volumes := helper.Specs.Volumes
	mountPath := helper.template.FuseMountInfo.HostMountPath
	if common.HostPathMode(helper.Specs.MetaObj.Annotations[common.HostMountPathModeOnDefaultPlatformKey]) == common.HostPathModeRandomSuffix {
		mountPath, err = generateUniqueHostPath(helper, mountPath)
		if err != nil {
			return err
		}
	}

	if helper.template.FuseMountInfo.SubPath != "" {
		mountPath = mountPath + "/" + helper.template.FuseMountInfo.SubPath
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
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == helper.pvcName {
			name := volume.Name
			volumes[i] = mutatedDatasetVolume
			volumes[i].Name = name
			overriddenVolumeNames = append(overriddenVolumeNames, name)
		}
	}

	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

	helper.ctx.SetDatasetUsedInContainers(false)
	for _, container := range helper.Specs.Containers {
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(overriddenVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
				helper.ctx.SetDatasetUsedInContainers(true)
			}
		}
	}

	helper.ctx.SetDatasetUsedInInitContainers(false)
	for _, container := range helper.Specs.InitContainers {
		for i, volumeMount := range container.VolumeMounts {
			if utils.ContainsString(overriddenVolumeNames, volumeMount.Name) {
				container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
				helper.ctx.SetDatasetUsedInInitContainers(true)
			}
		}
	}

	return nil
}

func defaultAppendFuseContainerVolumes(helper *helperData) (err error) {
	// collect all volumes' names
	var (
		volumeNames  = []string{}
		volumesToAdd = helper.template.VolumesToAdd
		nameSuffix   = helper.nameSuffix
	)
	for _, volume := range helper.Specs.Volumes {
		volumeNames = append(volumeNames, volume.Name)
	}

	if common.HostPathMode(helper.Specs.MetaObj.Annotations[common.HostMountPathModeOnDefaultPlatformKey]) == common.HostPathModeRandomSuffix {
		for index, volume := range volumesToAdd {
			if utils.IsVolumeNameHasPrefixes(volume, hostMountNames) {
				if volume.HostPath != nil {
					volume.HostPath.Path = fmt.Sprintf("%s/%s", volume.HostPath.Path, helper.ctx.generateUniqueHostMountPath)
					volumesToAdd[index] = volume
				}
			}
		}
	}

	// Append volumes
	ctxAppenedVolumeNames, err := helper.ctx.GetAppendedVolumeNames()
	if err != nil {
		return err
	}
	if len(volumesToAdd) > 0 {
		helper.log.V(1).Info("Before append volume", "original", helper.Specs.Volumes)
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
			helper.Specs.Volumes = append(helper.Specs.Volumes, volumeToAdd)
			if oldVolumeName != newVolumeName {
				ctxAppenedVolumeNames[oldVolumeName] = newVolumeName
			}
		}

		helper.log.V(1).Info("After append volume", "original", helper.Specs.Volumes)
	}
	helper.ctx.SetAppendedVolumeNames(ctxAppenedVolumeNames)

	return nil
}

func defaultInjectFuseContainer(helper *helperData) (err error) {
	used, err := helper.ctx.GetDatasetUsedInContainers()
	if err != nil {
		return err
	}

	if used {
		if err := prependFuseContainer(helper, false /* asInit */); err != nil {
			return err
		}
	}

	used, err = helper.ctx.GetDatasetUsedInInitContainers()
	if err != nil {
		return err
	}

	if used {
		if err := prependFuseContainer(helper, true /* asInit */); err != nil {
			return err
		}
	}

	return nil
}

func defaultInjectFuseNativeSidecar(helper *helperData) (err error) {
	usedInContainers, err := helper.ctx.GetDatasetUsedInContainers()
	if err != nil {
		return err
	}

	usedInInitContainers, err := helper.ctx.GetDatasetUsedInInitContainers()
	if err != nil {
		return err
	}

	if usedInContainers || usedInInitContainers {
		if err := prependFuseNativeSidecar(helper); err != nil {
			return err
		}
	}

	return nil
}

func prepareFuseContainerPostStartScript(helper *helperData) error {
	// 4. inject the post start script for fuse container, if configmap doesn't exist, try to create it.
	// Post start script varies according to privileged or unprivileged sidecar.
	var (
		info             = helper.runtimeInfo
		template         = helper.template
		datasetName      = info.GetName()
		datasetNamespace = info.GetNamespace()
	)

	dataset, err := utils.GetDataset(helper.client, datasetName, datasetNamespace)
	if err != nil {
		return err
	}

	// Fluid assumes pvc name is the same with runtime's name
	gen := poststart.NewDefaultPostStartScriptGenerator()
	cmKey := gen.GetNamespacedConfigMapKey(types.NamespacedName{Namespace: datasetNamespace, Name: datasetName}, template.FuseMountInfo.FsType)
	found, err := kubeclient.IsConfigMapExist(helper.client, cmKey.Name, cmKey.Namespace)
	if err != nil {
		return err
	}

	if !found {
		cm := gen.BuildConfigMap(dataset, cmKey)
		err = helper.client.Create(context.TODO(), cm)
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
	template.FuseContainer.Lifecycle.PostStart = gen.GetPostStartCommand(template.FuseMountInfo.ContainerMountPath, template.FuseMountInfo.FsType, template.FuseMountInfo.SubPath)
	template.VolumesToAdd = append(template.VolumesToAdd, gen.GetVolume(cmKey))

	return nil
}

func transformTemplateWithCacheDirDisabled(helper *helperData) {
	template := helper.template
	template.FuseContainer.VolumeMounts = utils.TrimVolumeMounts(template.FuseContainer.VolumeMounts, cacheDirNames)
	template.VolumesToAdd = utils.TrimVolumes(template.VolumesToAdd, cacheDirNames)
}
func enablePrometheusMetricsScrape(helper *helperData) {
	if helper.Specs.MetaObj.Annotations == nil {
		helper.Specs.MetaObj.Annotations = map[string]string{}
	}

	if _, exists := helper.Specs.MetaObj.Annotations[common.AnnotationPrometheusFuseMetricsScrapeKey]; !exists {
		helper.Specs.MetaObj.Annotations[common.AnnotationPrometheusFuseMetricsScrapeKey] = "true"
	}
}

func generateUniqueHostPath(helper *helperData, originalMountPath string) (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	uniqueDirName := hex.EncodeToString(bytes)[:8]

	name := helper.Specs.MetaObj.Name
	if helper.Specs.MetaObj.Name == "" {
		name = helper.Specs.MetaObj.GenerateName + "--generate-name"
	}
	generateUniqueHostMountPath := fmt.Sprintf("%s/%d-%s", name, time.Now().UnixMicro(), uniqueDirName)
	helper.ctx.generateUniqueHostMountPath = generateUniqueHostMountPath

	mountPathParts := strings.Split(originalMountPath, string(filepath.Separator))
	if len(mountPathParts) < 2 {
		return "", fmt.Errorf("unsupported host mount path format: %s", originalMountPath)
	}
	baseComponents := mountPathParts[:len(mountPathParts)-1]
	lastComponent := mountPathParts[len(mountPathParts)-1]
	newPathComponents := append(baseComponents, generateUniqueHostMountPath, lastComponent)
	return fmt.Sprintf("/%s", filepath.Join(newPathComponents...)), nil
}

func removeFuseMetricsContainerPort(helper *helperData) {
	if len(helper.template.FuseContainer.Ports) == 0 {
		return
	}

	containerPorts := []corev1.ContainerPort{}
	for _, containerPort := range helper.template.FuseContainer.Ports {
		if strings.HasSuffix(containerPort.Name, "-metrics") {
			continue
		}
		containerPorts = append(containerPorts, containerPort)
	}

	helper.template.FuseContainer.Ports = containerPorts
}

func prependFuseContainer(helper *helperData, asInit bool) error {
	fuseContainer := helper.template.FuseContainer
	if !asInit {
		fuseContainer.Name = common.FuseContainerName + helper.nameSuffix
	} else {
		fuseContainer.Name = common.InitFuseContainerName + helper.nameSuffix
	}

	if asInit {
		fuseContainer.Lifecycle = nil
		fuseContainer.Command = []string{"sleep"}
		fuseContainer.Args = []string{"2s"}
	}

	ctxAppenedVolumeNames, err := helper.ctx.GetAppendedVolumeNames()
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
		helper.Specs.Containers = append([]corev1.Container{fuseContainer}, helper.Specs.Containers...)
	} else {
		helper.Specs.InitContainers = append([]corev1.Container{fuseContainer}, helper.Specs.InitContainers...)
	}

	// TODO: move this to annotation because label has a max length limit for both key and value
	containerDatasetMappingLabelKey := common.LabelContainerDatasetMappingKeyPrefix + fuseContainer.Name
	if helper.Specs.MetaObj.Labels == nil {
		helper.Specs.MetaObj.Labels = map[string]string{}
	}
	helper.Specs.MetaObj.Labels[containerDatasetMappingLabelKey] = fmt.Sprintf("%s_%s", helper.runtimeInfo.GetNamespace(), helper.runtimeInfo.GetName())
	return nil
}

func prependFuseNativeSidecar(helper *helperData) error {
	fuseContainer := helper.template.FuseContainer
	fuseContainer.Name = common.FuseContainerName + helper.nameSuffix

	ctxAppenedVolumeNames, err := helper.ctx.GetAppendedVolumeNames()
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

	containerRestartPolicyAlways := corev1.ContainerRestartPolicyAlways
	fuseContainer.RestartPolicy = &containerRestartPolicyAlways
	helper.Specs.InitContainers = append([]corev1.Container{fuseContainer}, helper.Specs.InitContainers...)

	// TODO: move this to annotation because label has a max length limit for both key and value
	containerDatasetMappingLabelKey := common.LabelContainerDatasetMappingKeyPrefix + fuseContainer.Name
	if helper.Specs.MetaObj.Labels == nil {
		helper.Specs.MetaObj.Labels = map[string]string{}
	}
	helper.Specs.MetaObj.Labels[containerDatasetMappingLabelKey] = fmt.Sprintf("%s_%s", helper.runtimeInfo.GetNamespace(), helper.runtimeInfo.GetName())
	return nil
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
