package alluxio

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// transform master volumes
func (e *AlluxioEngine) transformMasterVolumes(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	if len(runtime.Spec.Master.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Master.VolumeMounts {
			name := volumeMount.Name
			volume := &corev1.Volume{}

			for _, v := range runtime.Spec.Volumes {
				if v.Name == name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", name)
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

// transform worker volumes
func (e *AlluxioEngine) transformWorkerVolumes(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
	if len(runtime.Spec.Worker.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Worker.VolumeMounts {
			name := volumeMount.Name
			volume := &corev1.Volume{}

			for _, v := range runtime.Spec.Volumes {
				if v.Name == name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", name)
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
			name := volumeMount.Name
			volume := &corev1.Volume{}

			for _, v := range runtime.Spec.Volumes {
				if v.Name == name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", name)
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
