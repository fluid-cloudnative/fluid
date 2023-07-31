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
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func GenDataProcessValueFile(dataset *datav1alpha1.Dataset, dataProcess *datav1alpha1.DataProcess) (valueFileName string, err error) {
	dataProcessValue := GenDataProcessValue(dataset, dataProcess)

	data, err := yaml.Marshal(dataProcessValue)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal dataProcessValue of DataProcess %s/%s", dataProcess.GetNamespace(), dataProcess.GetName())
	}

	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-process-values.yaml", dataProcess.Namespace, dataProcess.Name))
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temp file to store values for DataProcess %s/%s", dataProcess.Namespace, dataProcess.Name)
	}

	err = os.WriteFile(valueFile.Name(), data, 0o400)
	if err != nil {
		return "", errors.Wrapf(err, "failed to write temp file %s", valueFile.Name())
	}

	return valueFile.Name(), nil
}

func GenDataProcessValue(dataset *datav1alpha1.Dataset, dataProcess *datav1alpha1.DataProcess) *DataProcessValue {
	value := &DataProcessValue{
		DataProcessInfo: DataProcessInfo{
			TargetDataset: dataProcess.Spec.Dataset.Name,
		},
	}

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
		podSpec := dataProcess.Spec.Processor.Job.PodSpec
		if len(podSpec.Containers) != 0 || len(podSpec.InitContainers) != 0 {
			podSpec.Volumes = append(podSpec.Volumes, volumes...)

			if len(podSpec.Containers) != 0 {
				for idx := range podSpec.Containers {
					podSpec.Containers[idx].VolumeMounts = append(podSpec.Containers[idx].VolumeMounts, volumeMounts...)
				}
			}

			if len(podSpec.InitContainers) != 0 {
				for idx := range podSpec.InitContainers {
					podSpec.InitContainers[idx].VolumeMounts = append(podSpec.InitContainers[idx].VolumeMounts, volumeMounts...)
				}
			}
		}
		value.DataProcessInfo.JobProcessor = &JobProcessor{
			PodSpec: podSpec,
		}
		return value
	}

	if dataProcess.Spec.Processor.Script != nil {
		value.Name = dataProcess.Name
		value.DataProcessInfo.ScriptProcessor = &ScriptProcessor{
			Image:           fmt.Sprintf("%s:%s", dataProcess.Spec.Processor.Script.Image, dataProcess.Spec.Processor.Script.ImageTag),
			ImagePullPolicy: corev1.PullPolicy(dataProcess.Spec.Processor.Script.ImagePullPolicy),
			Envs:            dataProcess.Spec.Processor.Script.Env,
			Volumes:         append(dataProcess.Spec.Processor.Script.Volumes, volumes...),
			VolumeMounts:    append(dataProcess.Spec.Processor.Script.VolumeMounts, volumeMounts...),
			Command:         dataProcess.Spec.Processor.Script.Command,
			Args:            dataProcess.Spec.Processor.Script.Args,
		}
		return value
	}

	// unreachable code
	return nil
}
