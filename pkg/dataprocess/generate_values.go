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
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
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

	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	if len(dataProcess.Spec.Dataset.MountPath) != 0 {
		volumes = []corev1.Volume{
			{
				Name: "fluid-dataset-vol",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: dataset.Name,
					},
				},
			},
		}

		volumeMounts = []corev1.VolumeMount{
			{
				Name:      "fluid-dataset-vol",
				MountPath: dataProcess.Spec.Dataset.MountPath,
				SubPath:   dataProcess.Spec.Dataset.SubPath,
			},
		}
	}

	transformCommonPart(value, dataProcess)

	processorImpl := GetProcessorImpl(dataProcess)
	if processorImpl != nil { // processorImpl should always be non-nil
		processorImpl.TransformDataProcessValues(value, volumes, volumeMounts)
		return value
	}

	// unreachable code
	return nil
}

func transformCommonPart(value *DataProcessValue, dataProcess *datav1alpha1.DataProcess) {
	value.Name = dataProcess.Name
	value.DataProcessInfo.Labels = dataProcess.Spec.Processor.PodMetadata.Labels
	value.DataProcessInfo.Annotations = dataProcess.Spec.Processor.PodMetadata.Annotations
	value.Owner = transfromer.GenerateOwnerReferenceFromObject(dataProcess)
	if len(dataProcess.Spec.Processor.ServiceAccountName) != 0 {
		value.DataProcessInfo.ServiceAccountName = dataProcess.Spec.Processor.ServiceAccountName
	}
}
