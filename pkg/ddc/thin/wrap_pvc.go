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
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// bindDatasetToMountedPersistentVolumeClaim appends Dataset's ownerReference with the mounted PVC
// as its owner, so that the lifecycle of the Dataset will be tied to that of the mounted PVC.
func (t *ThinEngine) bindDatasetToMountedPersistentVolumeClaim() (err error) {

	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return errors.Wrapf(err, "can't find dataset \"%s/%s\" when binding PVC and Dataset", t.namespace, t.name)
	}

	var pvc *corev1.PersistentVolumeClaim
	for _, mount := range dataset.Spec.Mounts {
		if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(mount.MountPoint, common.VolumeScheme.String())
			mountedPvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
			if err != nil {
				return errors.Wrapf(err, "failed to get pvc \"%s/%s\" when checking dataset mounts", t.namespace, pvcName)
			}

			if pvc != nil {
				return fmt.Errorf("dataset \"%s/%s\" can only contain one pvc:// mount point", dataset.Namespace, dataset.Name)
			}
			pvc = mountedPvc
		}
	}

	// bind dataset only when there is specified pvc:// scheme mount point
	if pvc != nil {
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			ds, err := utils.GetDataset(t.Client, dataset.Name, dataset.Namespace)
			if err != nil {
				return err
			}

			pvcOwnerReference := metav1.OwnerReference{
				APIVersion: pvc.APIVersion,
				Kind:       pvc.Kind,
				Name:       pvc.Name,
				UID:        pvc.UID,
			}

			var exists bool
			for _, refer := range ds.OwnerReferences {
				if reflect.DeepEqual(refer, pvcOwnerReference) {
					exists = true
				}
			}

			if !exists {
				dsToUpdate := ds.DeepCopy()
				dsToUpdate.OwnerReferences = append(dsToUpdate.OwnerReferences, pvcOwnerReference)

				return t.Client.Update(context.TODO(), dsToUpdate)
			}

			return nil
		})

		if err != nil {
			return errors.Wrapf(err, "failed to update dataset \"%s/%s\"'s ownerReference", dataset.Namespace, dataset.Name)
		}
	}

	return nil
}

// wrapMountedPersistentVolumeClaim wraps mounted PVC specified in Dataset's spec by
// adding a special label on the PVC. The label can be recognized by Fluid's component to
// indicate which Dataset handles the PVC.
func (t *ThinEngine) wrapMountedPersistentVolumeClaim() (err error) {
	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return err
	}

	for _, mount := range dataset.Spec.Mounts {
		if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(mount.MountPoint, common.VolumeScheme.String())

			mountedPvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
			if err != nil {
				return err
			}

			if _, exists := mountedPvc.Labels[common.LabelAnnotationManagedBy]; !exists {
				labelsToModify := common.LabelsToModify{}
				labelsToModify.Add(common.LabelAnnotationManagedBy, t.name)
				_, err = utils.PatchLabels(t.Client, mountedPvc, labelsToModify)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

// unwrapMountedPersistentVolumeClaims unwraps mounted PVC by deleting the label on it.
// The func checks ThinRuntime's status instead of Dataset's spec in case that Dataset is
// already deleted.
func (t *ThinEngine) unwrapMountedPersistentVolumeClaims() (err error) {
	runtime, err := utils.GetThinRuntime(t.Client, t.name, t.namespace)
	if err != nil {
		return errors.Wrapf(err, "failed to get runtime %s/%s when unwrapping mounted pvcs", t.namespace, t.name)
	}

	for _, datasetMount := range runtime.Status.Mounts {
		if strings.HasPrefix(datasetMount.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(datasetMount.MountPoint, common.VolumeScheme.String())
			pvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
			if utils.IgnoreNotFound(err) != nil {
				return errors.Wrapf(err, "failed to get pvc when unwrapping pvc %s", pvcName)
			}

			if wrappedBy, exists := pvc.Labels[common.LabelAnnotationManagedBy]; exists && wrappedBy == t.name {
				labelsToModify := common.LabelsToModify{}
				labelsToModify.Delete(common.LabelAnnotationManagedBy)
				if _, err = utils.PatchLabels(t.Client, pvc, labelsToModify); err != nil {
					return errors.Wrapf(err, "failed to remove label when unwrapping pvc %s", pvc.Name)
				}
			}
		}
	}

	return
}
