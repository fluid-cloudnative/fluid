/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
)

// transform worker volumes
func (t *ThinEngine) transformWorkerVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, value *ThinValue) (err error) {
	if len(volumeMounts) > 0 {
		for _, volumeMount := range volumeMounts {
			var volume *corev1.Volume

			for _, v := range volumes {
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
func (t *ThinEngine) transformFuseVolumes(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, value *ThinValue) (err error) {
	if len(volumeMounts) > 0 {
		for _, volumeMount := range volumeMounts {
			var volume *corev1.Volume
			for _, v := range volumes {
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

// transform secrets for any specified mount points with volume schemes
func (t *ThinEngine) transfromSecretsForPersistentVolumeClaimMounts(dataset *datav1alpha1.Dataset, value *ThinValue) (err error) {
	for _, mount := range dataset.Spec.Mounts {
		if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(mount.MountPoint, common.VolumeScheme.String())

			pvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
			if err != nil {
				return err
			}

			if len(pvc.Spec.VolumeName) == 0 || pvc.Status.Phase != corev1.ClaimBound {
				return fmt.Errorf("persistent volume claim %s is not bounded yet", pvcName)
			}

			pv, err := kubeclient.GetPersistentVolume(t.Client, pvc.Spec.VolumeName)
			if err != nil {
				return err
			}

			// Currently only handle NodePublishSecret and ignore other secret refs.
			if pv.Spec.CSI == nil {
				return fmt.Errorf("persistent volume %s has unsupported volume source. only CSI is supported", pv.Name)
			}

			if pv.Spec.CSI.NodePublishSecretRef != nil {
				if len(pv.Spec.CSI.NodePublishSecretRef.Namespace) != 0 &&
					pv.Spec.CSI.NodePublishSecretRef.Namespace != t.namespace {
					return fmt.Errorf("namespace of node publish secret in the persistent volume %s must be equal to dataset %s's namespace", pv.Name, dataset.Name)
				}

				secretName := pv.Spec.CSI.NodePublishSecretRef.Name
				if len(secretName) == 0 {
					// skip mounting secret volume
					continue
				}

				volumeToAdd := corev1.Volume{
					Name: secretName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				}
				value.Fuse.Volumes = utils.AppendOrOverrideVolume(value.Fuse.Volumes, volumeToAdd)

				volumeMountToAdd := corev1.VolumeMount{
					Name:      secretName,
					ReadOnly:  true,
					MountPath: fmt.Sprintf("/etc/fluid/secrets/%s", secretName),
				}
				value.Fuse.VolumeMounts = utils.AppendOrOverrideVolumeMounts(value.Fuse.VolumeMounts, volumeMountToAdd)
			}
		}
	}

	return nil
}
