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
	value.DataProcessInfo = DataProcessInfo{
		TargetDataset: dataset.Name,
		ProcessJobInfo: ProcessJobInfo{
			Image:        dataProcess.Spec.Processor.Job.Image,
			Script:       dataProcess.Spec.Processor.Job.Script,
			Volumes:      volumes,
			VolumeMounts: volumeMounts,
		},
	}

	return value
}
