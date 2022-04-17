/*
Copyright 2021 The Fluid Authors.

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

package utils

import (
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
