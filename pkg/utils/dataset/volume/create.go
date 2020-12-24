package volume

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreatePersistentVolumeForRuntime creates PersistentVolume with the runtime Info
func CreatePersistentVolumeForRuntime(client client.Client,
	runtime base.RuntimeInfoInterface,
	mountPath string,
	mountType string,
	log logr.Logger) (err error) {
	accessModes, err := utils.GetAccessModesOfDataset(client, runtime.GetName(), runtime.GetNamespace())
	if err != nil {
		return err
	}

	found, err := kubeclient.IsPersistentVolumeExist(client, runtime.GetName(), common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pv := &v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtime.GetName(),
				Namespace: runtime.GetNamespace(),
				Labels: map[string]string{
					runtime.GetCommonLabelname(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeSpec{
				AccessModes: accessModes,
				Capacity: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("100Gi"),
				},
				StorageClassName: common.FLUID_STORAGECLASS,
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:       common.CSI_DRIVER,
						VolumeHandle: runtime.GetName(),
						VolumeAttributes: map[string]string{
							common.FLUID_PATH: mountPath,
							common.Mount_TYPE: mountType,
						},
					},
				},
				NodeAffinity: &v1.VolumeNodeAffinity{
					Required: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      runtime.GetCommonLabelname(),
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"true"},
									},
								},
							},
						},
					},
				},
			},
		}

		err = client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}
	} else {
		log.Info("The persistent volume is created", "name", runtime.GetName())
	}

	return err
}

// CreatePersistentVolumeClaimForRuntime creates PersistentVolumeClaim with the runtime Info
func CreatePersistentVolumeClaimForRuntime(client client.Client,
	runtime base.RuntimeInfoInterface,
	log logr.Logger) (err error) {
	accessModes, err := utils.GetAccessModesOfDataset(client, runtime.GetName(), runtime.GetNamespace())
	if err != nil {
		return err
	}

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtime.GetName(),
				Namespace: runtime.GetNamespace(),
				Labels: map[string]string{
					runtime.GetCommonLabelname(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						runtime.GetCommonLabelname(): "true",
					},
				},
				StorageClassName: &common.FLUID_STORAGECLASS,
				AccessModes:      accessModes,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse("100Gi"),
					},
				},
			},
		}

		err = client.Create(context.TODO(), pvc)
		if err != nil {
			return err
		}
	}

	return err
}
