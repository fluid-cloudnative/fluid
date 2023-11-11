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

package alluxio

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"path/filepath"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// transform master volumes
func (e *AlluxioEngine) transformMasterVolumes(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	if len(runtime.Spec.Master.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Master.VolumeMounts {
			var volume *corev1.Volume
			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Master.VolumeMounts) == 0 {
				value.Master.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Master.VolumeMounts = append(value.Master.VolumeMounts, volumeMount)

			if len(value.Master.Volumes) == 0 {
				value.Master.Volumes = []corev1.Volume{}
			}
			value.Master.Volumes = append(value.Master.Volumes, *volume)
		}
	}

	return err
}

func (e *AlluxioEngine) transformEncryptOptionsToMasterVolumes(dataset *datav1alpha1.Dataset, value *Alluxio) (options map[string]string) {
	for _, m := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(m.MountPoint) {
			continue
		}
		options = make(map[string]string)
		for _, encryptOpt := range append(dataset.Spec.SharedEncryptOptions, m.EncryptOptions...) {
			secretName := encryptOpt.ValueFrom.SecretKeyRef.Name
			secretMountPath := fmt.Sprintf("/etc/fluid/secrets/%s", secretName)

			volName := fmt.Sprintf("alluxio-mount-secret-%s", secretName)
			volumeToAdd := corev1.Volume{
				Name: volName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			}
			value.Master.Volumes = utils.AppendOrOverrideVolume(value.Master.Volumes, volumeToAdd)
			volumeMountToAdd := corev1.VolumeMount{
				Name:      volName,
				ReadOnly:  true,
				MountPath: secretMountPath,
			}
			value.Master.VolumeMounts = utils.AppendOrOverrideVolumeMounts(value.Master.VolumeMounts, volumeMountToAdd)
			options[encryptOpt.Name] = filepath.Join(secretMountPath, encryptOpt.ValueFrom.SecretKeyRef.Key)
		}
	}
	return options
}

// transform worker volumes
func (e *AlluxioEngine) transformWorkerVolumes(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	if len(runtime.Spec.Worker.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Worker.VolumeMounts {
			var volume *corev1.Volume

			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Worker.VolumeMounts) == 0 {
				value.Worker.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Worker.VolumeMounts = append(value.Worker.VolumeMounts, volumeMount)

			if len(value.Worker.Volumes) == 0 {
				value.Worker.Volumes = []corev1.Volume{}
			}
			value.Worker.Volumes = append(value.Worker.Volumes, *volume)
		}
	}

	return err
}

// transform fuse volumes
func (e *AlluxioEngine) transformFuseVolumes(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	if len(runtime.Spec.Fuse.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Fuse.VolumeMounts {
			var volume *corev1.Volume
			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Fuse.VolumeMounts) == 0 {
				value.Fuse.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Fuse.VolumeMounts = append(value.Fuse.VolumeMounts, volumeMount)

			if len(value.Fuse.Volumes) == 0 {
				value.Fuse.Volumes = []corev1.Volume{}
			}
			value.Fuse.Volumes = append(value.Fuse.Volumes, *volume)
		}
	}

	return err
}
