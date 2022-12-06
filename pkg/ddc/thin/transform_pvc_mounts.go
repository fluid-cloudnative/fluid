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
					secret, err := kubeclient.GetSecret(t.Client, secretName, t.namespace)
					if err != nil {
						if utils.IgnoreNotFound(err) == nil {
							return fmt.Errorf("failed to transform pvc secret mount because secret \"%s/%s\" not found", t.namespace, secretName)
						}
						return err
					}

					volumeToAdd := corev1.Volume{
						Name: secret.Name,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secret.Name,
							},
						},
					}
					value.Fuse.Volumes = utils.AppendOrOverrideVolume(value.Fuse.Volumes, volumeToAdd)

					volumeMountToAdd := corev1.VolumeMount{
						Name:      secret.Name,
						ReadOnly:  true,
						MountPath: fmt.Sprintf("/etc/fluid/secrets/%s", secret.Name),
					}
					value.Fuse.VolumeMounts = utils.AppendOrOverrideVolumeMounts(value.Fuse.VolumeMounts, volumeMountToAdd)

				case datav1alpha1.CopyNodePublishSecretIfNotExists:
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
