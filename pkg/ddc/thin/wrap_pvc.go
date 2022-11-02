package thin

import (
	"context"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

			labelsToModify := common.LabelsToModify{}
			labelsToModify.Add(common.LabelAnnotationWrappedBy, t.name)
			_, err = utils.PatchLabels(t.Client, mountedPvc, labelsToModify)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *ThinEngine) unwrapMountedPersistentVolumeClaims() (err error) {
	selector := labels.SelectorFromSet(labels.Set{common.LabelAnnotationWrappedBy: t.name})

	var list corev1.PersistentVolumeClaimList
	err = t.Client.List(context.TODO(), &list, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return
	}

	labelsToModify := common.LabelsToModify{}
	labelsToModify.Delete(common.LabelAnnotationWrappedBy)

	for _, pvc := range list.Items {
		if _, err = utils.PatchLabels(t.Client, &pvc, labelsToModify); err != nil {
			return errors.Wrapf(err, "failed to remove label when unwrapping pvc %s", pvc.Name)
		}
	}

	return
}
