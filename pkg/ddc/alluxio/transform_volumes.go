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
			volume := corev1.Volume{}
			found := false
			for _, volume := range runtime.Spec.Volumes {
				if volume.Name == name {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("failed to find the volume for volumeMount %s", name)
			}

			if len(value.Master.VolumeMounts) == 0 {
				value.Master.VolumeMounts = []corev1.VolumeMount{}
			}

			if len(value.Volumes) == 0 {
				value.Volumes = []corev1.Volume{}
			}

		}
	}

	return err
}
