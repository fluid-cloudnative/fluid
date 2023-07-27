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

package dataprocess

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func GenDataProcessValue(dataset *datav1alpha1.Dataset, dataProcess *datav1alpha1.DataProcess) *DataProcessValue {
	value := &DataProcessValue{}

	volumes := []corev1.Volume{
		{
			Name: "fluid-dataset-vol",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: dataset.Name,
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "fluid-dataset-vol",
			MountPath: dataProcess.Spec.Dataset.MountPath,
			SubPath:   dataProcess.Spec.Dataset.SubPath,
		},
	}

	value.Name = dataProcess.Name
	if dataProcess.Spec.Processor.Job != nil {
		podTemplate := dataProcess.Spec.Processor.Job.PodTemplate
		if len(podTemplate.Template.Spec.Containers) != 0 || len(podTemplate.Template.Spec.InitContainers) != 0 {
			podTemplate.Template.Spec.Volumes = append(podTemplate.Template.Spec.Volumes, volumes...)

			if len(podTemplate.Template.Spec.Containers) != 0 {
				for idx, _ := range podTemplate.Template.Spec.Containers {
					podTemplate.Template.Spec.Containers[idx].VolumeMounts = append(podTemplate.Template.Spec.Containers[idx].VolumeMounts, volumeMounts...)
				}
			}

			if len(podTemplate.Template.Spec.InitContainers) != 0 {
				for idx, _ := range podTemplate.Template.Spec.InitContainers {
					podTemplate.Template.Spec.InitContainers[idx].VolumeMounts = append(podTemplate.Template.Spec.InitContainers[idx].VolumeMounts, volumeMounts...)
				}
			}
		}
		value.DataProcessInfo.JobProcessor = &JobProcessor{
			PodTemplate: podTemplate,
		}
		return value
	}

	if dataProcess.Spec.Processor.Script != nil {
		value.Name = dataProcess.Name
		value.DataProcessInfo.ScriptProcessor = &ScriptProcessor{
			Image:           fmt.Sprintf("%s:%s", dataProcess.Spec.Processor.Script.Image, dataProcess.Spec.Processor.Script.ImageTag),
			ImagePullPolicy: corev1.PullPolicy(dataProcess.Spec.Processor.Script.ImagePullPolicy),
			Envs:            dataProcess.Spec.Processor.Script.Env,
			Volumes:         dataProcess.Spec.Processor.Script.Volumes,
			VolumeMounts:    dataProcess.Spec.Processor.Script.VolumeMounts,
			Command:         dataProcess.Spec.Processor.Script.Command,
			Args:            dataProcess.Spec.Processor.Script.Args,
		}
		return value
	}

	// unreachable code
	return nil
}
