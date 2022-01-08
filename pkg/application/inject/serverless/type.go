package serverless

import corev1 "k8s.io/api/core/v1"

type serverlessInjectionTemplate struct {
	FuseContainer        corev1.Container
	VolumeMountsToUpdate []corev1.VolumeMount
	VolumeMountsToAdd    []corev1.VolumeMount
	VolumesToUpdate      map[string]corev1.Volume
	VolumesToAdd         []corev1.Volume
}
