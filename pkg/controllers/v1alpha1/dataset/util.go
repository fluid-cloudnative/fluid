package dataset

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getDatasetRef(name, namespace string) string {
	// split by '#', can not use '-' because namespace or name can contain '-'
	return fmt.Sprintf("%s#%s", namespace, name)
}

func getPvName(name, namespace string) string {
	return fmt.Sprintf("%s-%s", name, namespace)
}

func createPersistentVolumeForRefDataset(client client.Client, virtualName string, virtualNameSpace string,
	runtime base.RuntimeInfoInterface, log logr.Logger) (err error) {

	pvName := getPvName(virtualName, virtualNameSpace)
	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		runtimePV, err := kubeclient.GetPersistentVolume(client, runtime.GetPersistentVolumeName())
		if err != nil {
			return err
		}

		runtimePVAttributes := runtimePV.Spec.PersistentVolumeSource.CSI.VolumeAttributes
		pv := &v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvName,
				Namespace: virtualNameSpace,
				Labels: map[string]string{
					runtime.GetCommonLabelName(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeSpec{
				AccessModes:      runtimePV.Spec.AccessModes,
				Capacity:         runtimePV.Spec.Capacity.DeepCopy(),
				StorageClassName: common.FluidStorageClass,
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       common.CSIDriver,
						VolumeHandle: runtimePV.Spec.PersistentVolumeSource.CSI.VolumeHandle,
						VolumeAttributes: map[string]string{
							// 挂载点，使用runtime pv fuse path
							common.VolumeAttrFluidPath: runtimePVAttributes[common.VolumeAttrFluidPath],
							common.VolumeAttrMountType: runtimePVAttributes[common.VolumeAttrMountType],
							common.VolumeAttrNamespace: runtimePVAttributes[common.VolumeAttrNamespace],
							common.VolumeAttrName:      runtimePVAttributes[common.VolumeAttrName],
						},
					},
				},
				NodeAffinity: runtimePV.Spec.NodeAffinity.DeepCopy(),
			},
		}
		err = client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}
	} else {
		log.Info("The ref persistent volume is created", "name", pvName)
	}

	return err
}

func createPersistentVolumeClaimForRefDataset(client client.Client, virtualName string, virtualNameSpace string,
	runtime base.RuntimeInfoInterface) (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, virtualName, virtualNameSpace, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		// for accessMode
		runtimePVC, err := kubeclient.GetPersistentVolumeClaim(client, runtime.GetName(), runtime.GetNamespace())
		if err != nil {
			return err
		}
		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      virtualName,
				Namespace: virtualNameSpace,
				Labels: map[string]string{
					//
					runtime.GetCommonLabelName(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						runtime.GetCommonLabelName(): "true",
					},
				},
				StorageClassName: &common.FluidStorageClass,
				AccessModes:      runtimePVC.Spec.AccessModes,
				Resources:        *runtimePVC.Spec.Resources.DeepCopy(),
			},
		}

		err = client.Create(context.TODO(), pvc)
		if err != nil {
			return err
		}
	}

	return err
}
