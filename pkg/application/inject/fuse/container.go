/*
Copyright 2022 The Fluid Authors.

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
	"errors"
	"regexp"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	initcopy "github.com/fluid-cloudnative/fluid/pkg/scripts/init-copy"
	corev1 "k8s.io/api/core/v1"
	utilpointer "k8s.io/utils/pointer"
)

func injectFuseContainerToFirst(containers []corev1.Container, fuseContainerName string,
	template *common.FuseInjectionTemplate,
	volumeNamesConflict map[string]string) []corev1.Container {
	fuseContainer := template.FuseContainer
	fuseContainer.Name = fuseContainerName

	for oldName, newName := range volumeNamesConflict {
		for i, volumeMount := range fuseContainer.VolumeMounts {
			if volumeMount.Name == oldName {
				fuseContainer.VolumeMounts[i].Name = newName
			}
		}
	}

	containers = append([]corev1.Container{fuseContainer}, containers...)
	return containers
}

func collectAllContainerNames(pod common.FluidObject) ([]string, error) {
	var allContainerNames []string

	containers, err := pod.GetContainers()
	if err != nil {
		return allContainerNames, err
	}

	for _, c := range containers {
		allContainerNames = append(allContainerNames, c.Name)
	}

	initContainers, err := pod.GetInitContainers()
	if err != nil {
		return allContainerNames, err
	}

	for _, c := range initContainers {
		allContainerNames = append(allContainerNames, c.Name)
	}

	return allContainerNames, nil
}

// changeForInitFuse change the original fuse template for init fuse
func (s *Injector) changeForInitFuse(runtimeInfo base.RuntimeInfoInterface, template *common.FuseInjectionTemplate, pvcName, files string) error {
	// 1. check if the files string contain shell command
	if err := checkShellCommand(files); err != nil {
		return err
	}

	fuseContainer := template.FuseContainer
	mountPath := "/" + common.InitPrefix + pvcName

	// 2. add volumemounts of emptyDir uesd in init phase
	emptyDir := corev1.VolumeMount{
		Name:      common.InitPrefix + pvcName,
		MountPath: mountPath,
	}

	// 3. add volumemounts of copy configmap
	copyConfigMap := corev1.VolumeMount{
		Name:      initcopy.CopyVolName,
		MountPath: initcopy.CopyScriptPath,
		ReadOnly:  true,
		SubPath:   initcopy.CopyScriptName,
	}

	fuseContainer.VolumeMounts = append(fuseContainer.VolumeMounts, emptyDir, copyConfigMap)

	// 4. add volumes of copy configmap
	var mode int32 = 0755
	template.VolumesToAdd = []corev1.Volume{
		{
			Name: initcopy.CopyVolName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: initcopy.CopyConfigMapName,
					},
					DefaultMode: utilpointer.Int32Ptr(mode),
				},
			},
		},
	}

	if err := runtimeInfo.GetOrCreateCopyConfigMap(); err != nil {
		return err
	}

	// 5. get and set the shell command
	var mountShell, mountArgs, checkShell, copyShell, umountShell, dsMountPoint string

	if len(fuseContainer.Command) > 1 {
		mountShell = fuseContainer.Command[len(fuseContainer.Command)-1]
	} else {
		mountShell = fuseContainer.Command[0]
	}

	for _, line := range fuseContainer.Args {
		mountArgs = mountArgs + " " + line
	}

	mountShell = "nohup " + mountShell + mountArgs + " & "

	lifycycle := fuseContainer.Lifecycle

	// PostStart is fixed: bash -c time /check-mount.sh <mount-path> <type>  >> /proc/1/fd/1
	checkShell = lifycycle.PostStart.Exec.Command[2]
	lifycycle.PreStop.ProtoMessage()

	dsMountPoint = strings.Split(checkShell, " ")[2]

	if !strings.HasSuffix(dsMountPoint, "-fuse") {
		dsMountPoint += "/" + runtimeInfo.GetRuntimeType() + "-fuse"
	}

	copyShell = "; " + initcopy.CopyScriptPath + " " + dsMountPoint + " " + mountPath + " " + files

	umountShell = "; umount " + dsMountPoint

	fuseContainer.Command = []string{"/bin/bash", "-c"}
	fuseContainer.Args = []string{mountShell + checkShell + copyShell + umountShell}

	fuseContainer.Lifecycle = nil
	fuseContainer.ReadinessProbe = nil

	// 6. set the fuse container
	template.FuseContainer = fuseContainer

	return nil
}

func checkShellCommand(filesStr string) error {
	files := strings.Split(filesStr, ",")
	compile, _ := regexp.Compile(`\A[\w\s\\\/\.\-]+\z`)

	for _, file := range files {
		if matched := compile.MatchString(file); !matched {
			return errors.New("Annotations <dataset>.init.fluid.io may contain shell command! " + filesStr)
		}
	}
	return nil
}

// generateInitDatasetMap generates the map (dataset name to files needed to use in init phase)
func (s *Injector) generateInitDatasetMap(annotations map[string]string) (dsName2SourceFiles map[string]string) {
	if annotations == nil {
		return
	}
	dsName2SourceFiles = map[string]string{}
	for key, value := range annotations {
		if strings.HasSuffix(key, common.AnnotationSuffix) {
			datasetName := strings.ReplaceAll(key, common.AnnotationSuffix, "")
			dsName2SourceFiles[datasetName] = value
		}
	}
	return dsName2SourceFiles
}
