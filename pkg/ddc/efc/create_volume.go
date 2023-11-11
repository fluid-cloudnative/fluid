/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package efc

import (
	"context"
	"path/filepath"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	volumehelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *EFCEngine) CreateVolume() (err error) {
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
	return
}

// createFusePersistentVolume
func (e *EFCEngine) createFusePersistentVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	mountRoot, err := utils.GetMountRoot()
	if err != nil {
		return err
	}
	// e.g. /runtime-mnt/efc-sock
	sessMgrWorkDir := filepath.Join(mountRoot, "efc-sock")

	return e.createPersistentVolumeForRuntime(runtimeInfo, e.getMountPath(), common.EFCMountType, sessMgrWorkDir)
}

// createFusePersistentVolume
func (e *EFCEngine) createFusePersistentVolumeClaim() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return volumehelper.CreatePersistentVolumeClaimForRuntime(e.Client, runtimeInfo, e.Log)
}

func (e *EFCEngine) createPersistentVolumeForRuntime(runtime base.RuntimeInfoInterface, mountPath string, mountType string, sessMgrWorkDir string) error {
	accessModes, err := utils.GetAccessModesOfDataset(e.Client, runtime.GetName(), runtime.GetNamespace())
	if err != nil {
		return err
	}

	pvName := runtime.GetPersistentVolumeName()

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, pvName, common.ExpectedFluidAnnotations)
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
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("100Pi"),
				},
				StorageClassName: common.FluidStorageClass,
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						Driver:       common.CSIDriver,
						VolumeHandle: pvName,
						VolumeAttributes: map[string]string{
							common.VolumeAttrFluidPath:         mountPath,
							common.VolumeAttrMountType:         mountType,
							common.VolumeAttrNamespace:         runtime.GetNamespace(),
							common.VolumeAttrName:              runtime.GetName(),
							common.VolumeAttrEFCSessMgrWorkDir: sessMgrWorkDir,
						},
					},
				},
				// NodeAffinity: &corev1.VolumeNodeAffinity{
				// 	Required: &corev1.NodeSelector{
				// 		NodeSelectorTerms: []corev1.NodeSelectorTerm{
				// 			{
				// 				MatchExpressions: []corev1.NodeSelectorRequirement{
				// 					{
				// 						Key:      runtime.GetCommonLabelName(),
				// 						Operator: corev1.NodeSelectorOpIn,
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
			e.Log.Info("Enable global mode for fuse in volume")
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
		} else {
			e.Log.Info("Disable global mode for fuse in volume")
			pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      runtime.GetCommonLabelName(),
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							},
						},
					},
				},
			}
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
