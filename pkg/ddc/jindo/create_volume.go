package jindo

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	expectedAnnotations = map[string]string{
		"CreatedBy": "fluid",
	}
)

// CreateVolume creates volume
func (e *JindoEngine) CreateVolume() (err error) {
	if e.runtime == nil {
		e.runtime, err = e.getRuntime()
		if err != nil {
			return
		}
	}

	err = e.createFusePersistentVolume()
	if err != nil {
		return err
	}

	err = e.createFusePersistentVolumeClaim()
	if err != nil {
		return err
	}

	return nil

}

// createFusePersistentVolume
func (e *JindoEngine) createFusePersistentVolume() (err error) {

	runtime, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	pvName := runtime.GetPersistentVolumeName()

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, pvName, expectedAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvName,
				Namespace: e.runtime.Namespace,
				Labels: map[string]string{
					e.getCommonLabelname(): "true",
				},
				Annotations: expectedAnnotations,
			},
			Spec: corev1.PersistentVolumeSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteMany,
				},
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("100Gi"),
				},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver:       CSI_DRIVER,
						VolumeHandle: pvName,
						VolumeAttributes: map[string]string{
							fluid_PATH: e.getMountPoint(),
							Mount_TYPE: common.JINDO_MOUNT_TYPE,
						},
					},
				},
				NodeAffinity: &corev1.VolumeNodeAffinity{
					Required: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      e.getCommonLabelname(),
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"true"},
									},
								},
							},
						},
					},
				},
			},
		}

		err = e.Client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}
	} else {
		e.Log.Info("The persistent volume is created", "name", pvName)
	}

	return err
}

// createFusePersistentVolume
func (e *JindoEngine) createFusePersistentVolumeClaim() (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.runtime.Name, e.runtime.Namespace, expectedAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        e.runtime.Name,
				Namespace:   e.runtime.Namespace,
				Annotations: expectedAnnotations,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteMany,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("100Gi"),
					},
				},
			},
		}

		err = e.Client.Create(context.TODO(), pvc)
		if err != nil {
			return err
		}
	}

	return err
}
