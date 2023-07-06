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

package thin

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *ThinEngine) transfromSecretsForPersistentVolumeClaimMounts(dataset *datav1alpha1.Dataset, policy datav1alpha1.NodePublishSecretPolicy, value *ThinValue) error {
	owner := transfromer.GenerateOwnerReferenceFromObject(t.runtime)
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
				secretName := pv.Spec.CSI.NodePublishSecretRef.Name
				if len(secretName) == 0 {
					continue
				}

				secretNamespace := pv.Spec.CSI.NodePublishSecretRef.Namespace
				if len(secretNamespace) == 0 {
					secretNamespace = corev1.NamespaceDefault
				}

				switch policy {
				case datav1alpha1.NotMountNodePublishSecret:
					break
				case datav1alpha1.MountNodePublishSecretIfExists:
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

				// CopyNodePublishSecretAndMountIfNotExists is supported but not allowed by default. Users must explicitly define role and rolebinding
				// for the service account "thinruntime-controller" in namespace "fluid-system".
				case datav1alpha1.CopyNodePublishSecretAndMountIfNotExists:
					fromNamespacedName := types.NamespacedName{Namespace: secretNamespace, Name: secretName}
					toNamespacedName := types.NamespacedName{Namespace: t.namespace, Name: fmt.Sprintf("%s-%s-publish-secret", t.name, t.runtimeType)}

					err = kubeclient.CopySecretToNamespace(t.Client, fromNamespacedName, toNamespacedName, owner)
					if err != nil {
						return errors.Wrapf(err, "failed to copy secret \"%s\" from namespace \"%s\" to namespace \"%s\"", secretName, secretNamespace, t.namespace)
					}

					volumeToAdd := corev1.Volume{
						Name: toNamespacedName.Name,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: toNamespacedName.Name,
							},
						},
					}
					value.Fuse.Volumes = utils.AppendOrOverrideVolume(value.Fuse.Volumes, volumeToAdd)

					volumeMountToAdd := corev1.VolumeMount{
						Name:      toNamespacedName.Name,
						ReadOnly:  true,
						MountPath: fmt.Sprintf("/etc/fluid/secrets/%s", secretName),
					}
					value.Fuse.VolumeMounts = utils.AppendOrOverrideVolumeMounts(value.Fuse.VolumeMounts, volumeMountToAdd)
				}
			}
		}
	}

	return nil
}
