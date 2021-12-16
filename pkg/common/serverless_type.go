package common

import corev1 "k8s.io/api/core/v1"

type ServerlessInjectionTemplate struct {
	FuseContainer        corev1.Container
	VolumeMountsToUpdate []corev1.VolumeMount
	VolumeMountsToAdd    []corev1.VolumeMount
	VolumesToUpdate      []corev1.Volume
	VolumesToAdd         []corev1.Volume
}
