/*

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

package alluxio

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cloudnativefluid/fluid/pkg/utils/kubeclient"
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
func (e *AlluxioEngine) CreateVolume() (err error) {
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

	err = e.createHCFSPersistentVolume()
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
func (e *AlluxioEngine) createFusePersistentVolume() (err error) {
	mountPath := fmt.Sprintf("%s/%s/%s/alluxio-fuse", ALLUXIO_MOUNT, e.namespace, e.name)

	if len(mountPath) == 0 {
		return fmt.Errorf("Failed to get the mountPath for %s in namespace %s", e.name,
			e.namespace)
	}

	if !filepath.IsAbs(mountPath) {
		return fmt.Errorf("The mount path is illegal %v", mountPath)
	}

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, e.runtime.Name, expectedAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      e.runtime.Name,
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
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("100Gi"),
				},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver:       CSI_DRIVER,
						VolumeHandle: e.runtime.Name,
						VolumeAttributes: map[string]string{
							fluid_PATH: mountPath,
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
		e.Log.Info("The persistent volume is created", "name", e.runtime.Name)
	}

	return err
}

// createFusePersistentVolume
func (e *AlluxioEngine) createFusePersistentVolumeClaim() (err error) {

	found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.runtime.Name, e.runtime.Namespace, expectedAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      e.runtime.Name,
				Namespace: e.runtime.Namespace,
				Labels: map[string]string{
					e.getCommonLabelname(): "true",
				},
				Annotations: expectedAnnotations,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					// MatchExpressions: []metav1.LabelSelectorRequirement{
					// 	{

					// 	},
					// },
					MatchLabels: map[string]string{
						e.getCommonLabelname(): "true",
					},
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteMany,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("100Gi"),
					},
				},
			},
		}

		err = e.Client.Create(context.TODO(), pvc)
		if err != nil {
			return err
		}
	} else {

	}

	return err
}

// createHCFSVolume (TODO: cheyang)
func (e *AlluxioEngine) createHCFSPersistentVolume() (err error) {
	return nil
}

// createHCFSVolume (TODO: cheyang)
func (e *AlluxioEngine) createHCFSPersistentVolumeClaim() (err error) {
	return nil
}
