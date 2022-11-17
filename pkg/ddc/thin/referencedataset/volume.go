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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	volumeHelper "github.com/fluid-cloudnative/fluid/pkg/utils/dataset/volume"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *ReferenceDatasetEngine) CreateVolume() (err error) {
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	mountedRuntimeInfo, err := e.getMountedRuntimeInfo()
	if err != nil {
		return err
	}

	err = createFusePersistentVolume(e.Client, runtimeInfo, mountedRuntimeInfo, e.Log)
	if err != nil {
		return err
	}
	err = createFusePersistentVolumeClaim(e.Client, runtimeInfo, mountedRuntimeInfo)

	return err
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

func createFusePersistentVolume(client client.Client, virtualRuntime base.RuntimeInfoInterface, physicalRuntime base.RuntimeInfoInterface, log logr.Logger) (err error) {
	virtualPvName := virtualRuntime.GetPersistentVolumeName()
	found, err := kubeclient.IsPersistentVolumeExist(client, virtualPvName, common.ExpectedFluidAnnotations)
	if err != nil {
		return err
	}

	if !found {
		physicalPV, err := kubeclient.GetPersistentVolume(client, physicalRuntime.GetPersistentVolumeName())
		if err != nil {
			return err
		}

		pv := &v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: virtualPvName,
				Labels: map[string]string{
					virtualRuntime.GetCommonLabelName(): "true",
				},
				Annotations: physicalPV.ObjectMeta.Annotations,
			},
			Spec: *physicalPV.Spec.DeepCopy(),
		}
		err = client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}
	} else {
		log.Info("The ref persistent volume is created", "name", virtualPvName)
	}

	return err
}

func createFusePersistentVolumeClaim(client client.Client, virtualRuntime base.RuntimeInfoInterface, physicalRuntime base.RuntimeInfoInterface) (err error) {
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
		pvc := &v1.PersistentVolumeClaim{
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
			Spec: v1.PersistentVolumeClaimSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						virtualRuntime.GetCommonLabelName(): "true",
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
