/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package dataprocess

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type Processor interface {
	ValidateDatasetMountPath(datasetMountPath string) (pass bool, conflictVolName string, conflictContainerName string)

	TransformDataProcessValues(value *DataProcessValue, datasetVolumes []corev1.Volume, datasetVolumeMounts []corev1.VolumeMount)
}

func GetProcessorImpl(dataProcess *datav1alpha1.DataProcess) Processor {
	if dataProcess.Spec.Processor.Job != nil {
		return &JobProcessorImpl{JobProcessor: dataProcess.Spec.Processor.Job}
	} else if dataProcess.Spec.Processor.Script != nil {
		return &ScriptProcessorImpl{ScriptProcessor: dataProcess.Spec.Processor.Script}
	}

	return nil
}

var _ Processor = &JobProcessorImpl{}

type JobProcessorImpl struct {
	*datav1alpha1.JobProcessor
}

func (p *JobProcessorImpl) ValidateDatasetMountPath(datasetMountPath string) (pass bool, conflictVolName string, conflictContainerName string) {
	for _, ctr := range append(p.JobProcessor.PodSpec.Containers, p.JobProcessor.PodSpec.InitContainers...) {
		for _, volMount := range ctr.VolumeMounts {
			if volMount.MountPath == datasetMountPath {
				return false, volMount.Name, ctr.Name
			}
		}
	}
	return true, "", ""
}

func (p *JobProcessorImpl) TransformDataProcessValues(value *DataProcessValue, datasetVolumes []corev1.Volume, datasetVolumeMounts []corev1.VolumeMount) {
	podSpec := p.JobProcessor.PodSpec
	if len(podSpec.Containers) != 0 || len(podSpec.InitContainers) != 0 {
		podSpec.Volumes = append(podSpec.Volumes, datasetVolumes...)

		if len(podSpec.Containers) != 0 {
			for idx := range podSpec.Containers {
				podSpec.Containers[idx].VolumeMounts = append(podSpec.Containers[idx].VolumeMounts, datasetVolumeMounts...)
			}
		}

		if len(podSpec.InitContainers) != 0 {
			for idx := range podSpec.InitContainers {
				podSpec.InitContainers[idx].VolumeMounts = append(podSpec.InitContainers[idx].VolumeMounts, datasetVolumeMounts...)
			}
		}
	}
	value.DataProcessInfo.JobProcessor = &JobProcessor{
		PodSpec: podSpec,
	}
}

var _ Processor = &ScriptProcessorImpl{}

type ScriptProcessorImpl struct {
	*datav1alpha1.ScriptProcessor
}

func (p *ScriptProcessorImpl) ValidateDatasetMountPath(datasetMountPath string) (pass bool, conflictVolName string, conflictContainerName string) {
	for _, volMount := range p.ScriptProcessor.VolumeMounts {
		if volMount.MountPath == datasetMountPath {
			return false, volMount.Name, DataProcessScriptProcessorContainerName
		}
	}

	return true, "", ""
}

func (p *ScriptProcessorImpl) TransformDataProcessValues(value *DataProcessValue, datasetVolumes []corev1.Volume, datasetVolumeMounts []corev1.VolumeMount) {
	value.DataProcessInfo.ScriptProcessor = &ScriptProcessor{
		Image:           fmt.Sprintf("%s:%s", p.ScriptProcessor.Image, p.ScriptProcessor.ImageTag),
		ImagePullPolicy: corev1.PullPolicy(p.ScriptProcessor.ImagePullPolicy),
		RestartPolicy:   p.ScriptProcessor.RestartPolicy,
		Envs:            p.ScriptProcessor.Env,
		Volumes:         append(p.ScriptProcessor.Volumes, datasetVolumes...),
		VolumeMounts:    append(p.ScriptProcessor.VolumeMounts, datasetVolumeMounts...),
		Command:         p.ScriptProcessor.Command,
		Source:          p.ScriptProcessor.Source,
	}
}
