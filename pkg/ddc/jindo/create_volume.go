package jindo

import (
	"context"
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

	mountPath := "/mnt/jfs"

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, e.runtime.Name, expectedAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:        e.runtime.Name,
				Namespace:   e.runtime.Namespace,
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
						Driver:       "fuse.csi.fluid.io",
						VolumeHandle: e.runtime.Name,
						VolumeAttributes: map[string]string{
							"fluid_path": mountPath,
							"mount_type": "fuse.jindofs-fuse",
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
		e.Log.Info("The persistent volume is created", "name", e.runtime.Name)
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
