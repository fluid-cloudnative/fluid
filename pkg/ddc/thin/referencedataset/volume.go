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

package referencedataset

import (
	"context"
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *ReferenceDatasetEngine) CreateVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	physicalRuntimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil {
		return err
	}

	accessModes, err := createFusePersistentVolume(e.Client, runtimeInfo, physicalRuntimeInfo, e.Log)
	if err != nil {
		return err
	}
	return createFusePersistentVolumeClaim(e.Client, runtimeInfo, physicalRuntimeInfo, accessModes)
}

func (e *ReferenceDatasetEngine) DeleteVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	err = volumeHelper.DeleteFusePersistentVolumeClaim(e.Client, runtimeInfo, e.Log)
	if err != nil {
		return err
	}

	err = volumeHelper.DeleteFusePersistentVolume(e.Client, runtimeInfo, e.Log)

	return err
}

func createFusePersistentVolume(client client.Client, virtualRuntime base.RuntimeInfoInterface, physicalRuntime base.RuntimeInfoInterface, log logr.Logger) (accessModes []corev1.PersistentVolumeAccessMode, err error) {
	virtualPvName := virtualRuntime.GetPersistentVolumeName()
	found, err := kubeclient.IsPersistentVolumeExist(client, virtualPvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return accessModes, err
	}

	if !found {
		physicalPV, err := kubeclient.GetPersistentVolume(client, physicalRuntime.GetPersistentVolumeName())
		if err != nil {
			return accessModes, err
		}

		copiedPvSpec := physicalPV.Spec.DeepCopy()

		virtualDataset, err := utils.GetDataset(client, virtualRuntime.GetName(), virtualRuntime.GetNamespace())
		if err != nil {
			return accessModes, err
		}
		// set the sub path attribute
		subPaths := base.GetPhysicalDatasetSubPath(virtualDataset)
		if len(subPaths) > 1 {
			return accessModes, fmt.Errorf("the dataset is not validated, only support dataset mounts which expects 1")
		}
		if len(subPaths) == 1 && subPaths[0] != "" {
			copiedPvSpec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath] = subPaths[0]
		}

		// set the accessModes
		// only allow readOnly when physical
		accessModes = accessModesForVirtualDataset(virtualDataset, copiedPvSpec.AccessModes)
		copiedPvSpec.AccessModes = accessModes

		if len(virtualDataset.Spec.AccessModes) > 0 &&
			!reflect.DeepEqual(virtualDataset.Spec.AccessModes, accessModes) {
			log.Info("AccessMode to set, expect and got",
				"dataset.AccessModes", virtualDataset.Spec.AccessModes,
				"pv.AccessModes", accessModes)
		}

		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: virtualPvName,
				Labels: map[string]string{
					virtualRuntime.GetCommonLabelName(): "true",
				},
				Annotations: physicalPV.ObjectMeta.Annotations,
			},
			Spec: *copiedPvSpec,
		}
		err = client.Create(context.TODO(), pv)
		if err != nil {
			return accessModes, err
		}
	} else {
		log.Info("The ref persistent volume is created", "name", virtualPvName)
	}

	return accessModes, err
}

// accessModesForVirtualDataset generates accessMode based on virtualDataset and copiedPvSpec
func accessModesForVirtualDataset(virtualDataset *datav1alpha1.Dataset, modes []corev1.PersistentVolumeAccessMode) []corev1.PersistentVolumeAccessMode {
	accessModes := virtualDataset.Spec.AccessModes
	readOnly := false
	//  If the physcial dataset is readOnly, the virtual dataset's accessMode shouldn't be greater than read
	for _, accessMode := range modes {
		if accessMode == corev1.ReadOnlyMany {
			readOnly = true
			break
		}
	}

	if len(accessModes) == 0 || readOnly {
		accessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadOnlyMany,
		}
	}
	return accessModes
}

func createFusePersistentVolumeClaim(client client.Client, virtualRuntime base.RuntimeInfoInterface, physicalRuntime base.RuntimeInfoInterface, accessModes []corev1.PersistentVolumeAccessMode) (err error) {
	virtualName := virtualRuntime.GetName()
	virtualNamespace := virtualRuntime.GetNamespace()

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, virtualName, virtualNamespace, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		runtimePVC, err := kubeclient.GetPersistentVolumeClaim(client, physicalRuntime.GetName(), physicalRuntime.GetNamespace())
		if err != nil {
			return err
		}
		// set the accessModes
		// only allow readOnly when physical

		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      virtualName,
				Namespace: virtualNamespace,
				Labels: map[string]string{
					// see 'pkg/util/webhook/scheduler/mutating/schedule_pod_handler.go' 'CheckIfPVCIsDataset' function usage
					common.LabelAnnotationStorageCapacityPrefix + virtualNamespace + "-" + virtualName: "true",
					common.LabelAnnotationDatasetReferringName:                                         runtimePVC.Name,
					common.LabelAnnotationDatasetReferringNameSpace:                                    runtimePVC.Namespace,
				},
				Annotations: common.ExpectedFluidAnnotations,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						virtualRuntime.GetCommonLabelName(): "true",
					},
				},
				StorageClassName: &common.FluidStorageClass,
				AccessModes:      accessModes,
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
