/*
Copyright 2023 The Fluid Author.

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

package volume

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

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

	storageCapacity, err := utils.GetPVCStorageCapacityOfDataset(client, runtime.GetName(), runtime.GetNamespace())
	if err != nil {
		return err
	}

	pvName := runtime.GetPersistentVolumeName()

	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvName,
				Namespace: runtime.GetNamespace(),
				Labels: map[string]string{
					runtime.GetCommonLabelName(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: corev1.PersistentVolumeSpec{
				AccessModes: accessModes,
				Capacity: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): storageCapacity,
				},
				StorageClassName: common.FluidStorageClass,
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver:       common.CSIDriver,
						VolumeHandle: pvName,
						VolumeAttributes: map[string]string{
							common.VolumeAttrFluidPath: mountPath,
							common.VolumeAttrMountType: mountType,
							common.VolumeAttrNamespace: runtime.GetNamespace(),
							common.VolumeAttrName:      runtime.GetName(),
						},
					},
				},
				// NodeAffinity: &v1.VolumeNodeAffinity{
				// 	Required: &v1.NodeSelector{
				// 		NodeSelectorTerms: []v1.NodeSelectorTerm{
				// 			{
				// 				MatchExpressions: []v1.NodeSelectorRequirement{
				// 					{
				// 						Key:      runtime.GetCommonLabelName(),
				// 						Operator: v1.NodeSelectorOpIn,
				// 						Values:   []string{"true"},
				// 					},
				// 				},
				// 			},
				// 		},
				// 	},
				// },
			},
		}

		nodeSelector := runtime.GetFuseNodeSelector()
		log.Info("Enable global mode for fuse in volume")
		if len(nodeSelector) > 0 {
			nodeSelectorRequirements := []corev1.NodeSelectorRequirement{}
			for key, value := range nodeSelector {
				nodeSelectorRequirements = append(nodeSelectorRequirements, corev1.NodeSelectorRequirement{
					Key:      key,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{value},
				})
			}
			pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: nodeSelectorRequirements,
						},
					},
				},
			}
		}

		// set from runtime
		for key, value := range runtime.GetAnnotations() {
			if key == common.AnnotationSkipCheckMountReadyTarget {
				pv.Spec.PersistentVolumeSource.CSI.VolumeAttributes[common.AnnotationSkipCheckMountReadyTarget] = value
			}
		}

		// set from annotations[data.fluid.io/metadataList]
		metadataList := runtime.GetMetadataList()
		for i := range metadataList {
			if selector := metadataList[i].Selector; selector.Group != corev1.GroupName || selector.Kind != "PersistentVolume" {
				continue
			}
			pv.Labels = utils.UnionMapsWithOverride(pv.Labels, metadataList[i].Labels)
			pv.Annotations = utils.UnionMapsWithOverride(pv.Annotations, metadataList[i].Annotations)
			// if pv labels has common.LabelNodePublishMethod and it's value is symlink, add to volumeAttributes
			if v, ok := metadataList[i].Labels[common.LabelNodePublishMethod]; ok && v == common.NodePublishMethodSymlink {
				pv.Spec.PersistentVolumeSource.CSI.VolumeAttributes[common.NodePublishMethod] = v
			}
		}

		err = client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}

		// Poll the PV's status until it enters an "Available" phase. The polling process timeouts after 1 second and retries every 200 milliseconds.
		timeoutCtx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelFn()
		pollErr := wait.PollUntilContextCancel(timeoutCtx, 200*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
			pvCreated, pvErr := kubeclient.GetPersistentVolume(client, pvName)
			if pvErr != nil {
				if utils.IgnoreNotFound(pvErr) == nil {
					log.Info("The persistent volume not found, waiting for cache to sync up", "pv", pvName)
				} else {
					log.Error(errors.Wrap(pvErr, "failed to get persistent volume"), "pv", pvName)
				}
				// Ignore pvErr to retry
				return false, nil
			}

			if pvCreated.Status.Phase == corev1.VolumeAvailable {
				log.Info("Persistent volume already entered phase Available", "pv", pvName)
				return true, nil
			}

			return false, nil
		})
		if pollErr != nil {
			log.Error(pollErr, "got error when polling PV's status", "pv", pvName)
		}
	} else {
		log.Info("The persistent volume is created", "name", pvName)
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

	storageCapacity, err := utils.GetPVCStorageCapacityOfDataset(client, runtime.GetName(), runtime.GetNamespace())
	if err != nil {
		return err
	}

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, runtime.GetName(), runtime.GetNamespace(), common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtime.GetName(),
				Namespace: runtime.GetNamespace(),
				Labels: map[string]string{
					runtime.GetCommonLabelName(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						runtime.GetCommonLabelName(): "true",
					},
				},
				StorageClassName: &common.FluidStorageClass,
				AccessModes:      accessModes,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: storageCapacity,
					},
				},
			},
		}
		metadataList := runtime.GetMetadataList()
		for i := range metadataList {
			if selector := metadataList[i].Selector; selector.Group != corev1.GroupName || selector.Kind != "PersistentVolumeClaim" {
				continue
			}
			pvc.Labels = utils.UnionMapsWithOverride(pvc.Labels, metadataList[i].Labels)
			pvc.Annotations = utils.UnionMapsWithOverride(pvc.Annotations, metadataList[i].Annotations)
		}

		err = client.Create(context.TODO(), pvc)
		if err != nil {
			return err
		}
	}

	return err
}
