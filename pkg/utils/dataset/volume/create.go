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

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
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

	pvName := runtime.GetPersistentVolumeName()

	found, err := kubeclient.IsPersistentVolumeExist(client, pvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		pv := &v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvName,
				Namespace: runtime.GetNamespace(),
				Labels: map[string]string{
					runtime.GetCommonLabelName(): "true",
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: v1.PersistentVolumeSpec{
				AccessModes: accessModes,
				Capacity: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("100Pi"),
				},
				StorageClassName: common.FluidStorageClass,
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
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

		global, nodeSelector := runtime.GetFuseDeployMode()
		if global {
			log.Info("Enable global mode for fuse in volume")
			if len(nodeSelector) > 0 {
				nodeSelectorRequirements := []v1.NodeSelectorRequirement{}
				for key, value := range nodeSelector {
					nodeSelectorRequirements = append(nodeSelectorRequirements, v1.NodeSelectorRequirement{
						Key:      key,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{value},
					})
				}
				pv.Spec.NodeAffinity = &v1.VolumeNodeAffinity{
					Required: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: nodeSelectorRequirements,
							},
						},
					},
				}
			}
		} else {
			log.Info("Disable global mode for fuse in volume")
			pv.Spec.NodeAffinity = &v1.VolumeNodeAffinity{
				Required: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      runtime.GetCommonLabelName(),
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							},
						},
					},
				},
			}
		}
		metadataList := runtime.GetMetadataList()
		for i := range metadataList {
			if selector := metadataList[i].Selector; selector.Group != v1.GroupName || selector.Kind != "PersistentVolume" {
				continue
			}
			pv.Labels = utils.UnionMapsWithOverride(pv.Labels, metadataList[i].Labels)
			pv.Annotations = utils.UnionMapsWithOverride(pv.Annotations, metadataList[i].Annotations)
		}

		err = client.Create(context.TODO(), pv)
		if err != nil {
			return err
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
				AccessModes:      accessModes,
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse("100Pi"),
					},
				},
			},
		}
		metadataList := runtime.GetMetadataList()
		for i := range metadataList {
			if selector := metadataList[i].Selector; selector.Group != v1.GroupName || selector.Kind != "PersistentVolumeClaim" {
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
