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

package utils

import (
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// TrimVolumes trims the volumes
func TrimVolumes(inputs []corev1.Volume, excludeNames []string) (outputs []corev1.Volume) {
	outputs = []corev1.Volume{}
outer:
	for _, in := range inputs {
		for _, excludeName := range excludeNames {
			if strings.HasPrefix(in.Name, excludeName) {
				log.V(1).Info("Skip the volume", "volume", in, "exludeName", excludeName)
				continue outer
			}
		}

		outputs = append(outputs, in)
	}

	return
}

func TrimVolumeMounts(inputs []corev1.VolumeMount, excludeNames []string) (outputs []corev1.VolumeMount) {
	outputs = []corev1.VolumeMount{}
outer:
	for _, in := range inputs {
		for _, excludeName := range excludeNames {
			if strings.HasPrefix(in.Name, excludeName) {
				log.V(1).Info("Skip the volumeMount", "volumeMount", in, "exludeName", excludeName)
				continue outer
			}
		}

		outputs = append(outputs, in)

	}

	return
}

func FindVolumeByVolumeMount(volumeMount corev1.VolumeMount, volumes []corev1.Volume) *corev1.Volume {
	for _, vol := range volumes {
		if vol.Name == volumeMount.Name {
			return &vol
		}
	}

	return nil
}

func AppendOrOverrideVolume(volumes []corev1.Volume, vol corev1.Volume) []corev1.Volume {
	var existed bool
	for idx, v := range volumes {
		if v.Name == vol.Name {
			if !reflect.DeepEqual(v, vol) {
				// override existing volume
				volumes[idx] = vol
			}
			existed = true
			break
		}
	}

	if !existed {
		volumes = append(volumes, vol)
	}

	return volumes
}

func AppendOrOverrideVolumeMounts(volumeMounts []corev1.VolumeMount, vm corev1.VolumeMount) []corev1.VolumeMount {
	var existed bool
	for idx, m := range volumeMounts {
		if m.Name == vm.Name {
			if !reflect.DeepEqual(m, vm) {
				// override existing volume mount
				volumeMounts[idx] = vm
			}
			existed = true
			break
		}
	}

	if !existed {
		volumeMounts = append(volumeMounts, vm)
	}

	return volumeMounts
}

// FilterVolumesByVolumeMounts returns volumes that exists in the volumeMounts
func FilterVolumesByVolumeMounts(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount) []corev1.Volume {
	retVolumes := []corev1.Volume{}
	for _, vol := range volumes {
		var exists bool
		for _, volMount := range volumeMounts {
			if volMount.Name == vol.Name {
				exists = true
				break
			}
		}

		if exists {
			retVolumes = append(retVolumes, vol)
		}
	}

	return retVolumes
}
