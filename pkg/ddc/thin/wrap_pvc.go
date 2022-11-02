package thin

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
)

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

			if _, exists := mountedPvc.Labels[common.LabelAnnotationWrappedBy]; !exists {
				labelsToModify := common.LabelsToModify{}
				labelsToModify.Add(common.LabelAnnotationWrappedBy, t.name)
				_, err = utils.PatchLabels(t.Client, mountedPvc, labelsToModify)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func (t *ThinEngine) unwrapMountedPersistentVolumeClaims() (err error) {
	for _, datasetMount := range t.runtime.Status.DatasetMounts {
		if strings.HasPrefix(datasetMount.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(datasetMount.MountPoint, common.VolumeScheme.String())
			pvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
			if err != nil {
				return errors.Wrapf(err, "failed to get pvc when unwrapping pvc %s", pvcName)
			}

			if wrappedBy, exists := pvc.Labels[common.LabelAnnotationWrappedBy]; exists && wrappedBy == t.name {
				labelsToModify := common.LabelsToModify{}
				labelsToModify.Delete(common.LabelAnnotationWrappedBy)
				if _, err = utils.PatchLabels(t.Client, pvc, labelsToModify); err != nil {
					return errors.Wrapf(err, "failed to remove label when unwrapping pvc %s", pvc.Name)
				}
			}
		}
	}

	return
}
