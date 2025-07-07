/*
  Copyright 2022 The Fluid Authors.

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
	found, err := kubeclient.IsPersistentVolumeExist(client, virtualPvName, common.GetExpectedFluidAnnotations())
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

		// set persistent volume's claim ref
		copiedPvSpec.ClaimRef = &corev1.ObjectReference{
			Namespace: virtualRuntime.GetNamespace(),
			Name:      virtualRuntime.GetName(),
		}

		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: virtualPvName,
				Labels: map[string]string{
					virtualRuntime.GetCommonLabelName(): "true",
					common.LabelAnnotationDatasetId:     utils.GetDatasetId(virtualDataset.GetNamespace(), virtualDataset.GetName(), string(virtualDataset.GetUID())),
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

	found, err := kubeclient.IsPersistentVolumeClaimExist(client, virtualName, virtualNamespace, common.GetExpectedFluidAnnotations())
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
					utils.GetNamespacedNameValueWithPrefix(common.LabelAnnotationStorageCapacityPrefix, virtualNamespace, virtualName, virtualRuntime.GetOwnerDatasetUID()): "true",
					common.LabelAnnotationDatasetReferringName:      runtimePVC.Name,
					common.LabelAnnotationDatasetReferringNameSpace: runtimePVC.Namespace,
					common.LabelAnnotationDatasetId:                 utils.GetDatasetId(virtualNamespace, virtualName, virtualRuntime.GetOwnerDatasetUID()),
				},
				Annotations: common.GetExpectedFluidAnnotations(),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				VolumeName: virtualRuntime.GetPersistentVolumeName(),
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
