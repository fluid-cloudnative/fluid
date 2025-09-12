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

package engine

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func (e *CacheEngine) CreateVolume() (err error) {
	if e.value == nil || e.value.Client.Enabled == false {
		return nil
	}
	if err = e.createFusePersistentVolume(); err != nil {
		return err
	}

	if err = e.createFusePersistentVolumeClaim(); err != nil {
		return err
	}
	return nil
}

func (e *CacheEngine) DeleteVolume() (err error) {
	if e.value == nil || e.value.Client.Enabled == false {
		return nil
	}

	if err = e.deleteFusePersistentVolumeClaim(); err != nil {
		return err
	}

	if err = e.deleteFusePersistentVolume(); err != nil {
		return err
	}

	return nil
}

func (e *CacheEngine) createFusePersistentVolume() error {
	accessModes, err := utils.GetAccessModesOfDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	storageCapacity, err := utils.GetPVCStorageCapacityOfDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	pvName := e.GetPersistentVolumeName()

	found, err := kubeclient.IsPersistentVolumeExist(e.Client, pvName, common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}

	runtimeClass, err := utils.GetCacheRuntimeClass(e.Client, runtime.Spec.RuntimeClassName)
	if err != nil {
		return errors.Wrapf(err, "failed to get CacheRuntimeClass %s", runtime.Spec.RuntimeClassName)
	}

	if !found {
		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pvName,
				Namespace: e.namespace,
				Labels: map[string]string{
					utils.GetCommonLabelName(false, e.namespace, e.name, ""): "true",
					common.LabelAnnotationDatasetId:                          utils.GetDatasetId(e.namespace, e.name, ""),
				},
				Annotations: common.GetExpectedFluidAnnotations(),
			},
			Spec: corev1.PersistentVolumeSpec{
				// Specify claim ref for faster volume binding
				// In Fluid, PVC's namespace/name is same as Dataset's namespace/name
				ClaimRef: &corev1.ObjectReference{
					Namespace: e.namespace,
					Name:      e.name,
				},
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
							common.VolumeAttrFluidPath: e.getTargetPath(),
							common.VolumeAttrMountType: runtimeClass.FileSystemType,
							common.VolumeAttrNamespace: runtime.GetNamespace(),
							common.VolumeAttrName:      runtime.GetName(),
						},
					},
				},
			},
		}

		nodeSelector := runtime.Spec.Client.NodeSelector
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

		err = e.Client.Create(context.TODO(), pv)
		if err != nil {
			return err
		}

		// Poll the PV's status until it enters an "Available" phase. The polling process timeouts after 1 second and retries every 200 milliseconds.
		timeoutCtx, cancelFn := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelFn()
		pollErr := wait.PollUntilContextCancel(timeoutCtx, 200*time.Millisecond, true, func(ctx context.Context) (done bool, err error) {
			pvCreated, pvErr := kubeclient.GetPersistentVolume(e.Client, pvName)
			if pvErr != nil {
				if utils.IgnoreNotFound(pvErr) == nil {
					e.Log.Info("The persistent volume not found, waiting for cache to sync up", "pv", pvName)
				} else {
					e.Log.Error(errors.Wrap(pvErr, "failed to get persistent volume"), "pv", pvName)
				}
				// Ignore pvErr to retry
				return false, nil
			}

			if pvCreated.Status.Phase == corev1.VolumeAvailable {
				e.Log.Info("Persistent volume already entered phase Available", "pv", pvName)
				return true, nil
			}

			return false, nil
		})
		if pollErr != nil {
			e.Log.Error(pollErr, "got error when polling PV's status", "pv", pvName)
		}
	} else {
		e.Log.Info("The persistent volume is created", "name", pvName)
	}

	return err
}

func (e *CacheEngine) createFusePersistentVolumeClaim() error {
	accessModes, err := utils.GetAccessModesOfDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	storageCapacity, err := utils.GetPVCStorageCapacityOfDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.name, e.namespace, common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if !found {
		pvc := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      e.name,
				Namespace: e.namespace,
				Labels: map[string]string{
					utils.GetCommonLabelName(false, e.namespace, e.name, ""): "true",
					common.LabelAnnotationDatasetId:                          utils.GetDatasetId(e.namespace, e.name, ""),
				},
				Annotations: common.GetExpectedFluidAnnotations(),
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				VolumeName:       e.GetPersistentVolumeName(),
				StorageClassName: &common.FluidStorageClass,
				AccessModes:      accessModes,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: storageCapacity,
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

func (e *CacheEngine) deleteFusePersistentVolume() error {
	pvName := e.GetPersistentVolumeName()
	found, err := kubeclient.IsPersistentVolumeExist(e.Client, pvName, common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolume(e.Client, pvName)
		if err != nil {
			return err
		}
		retries := 10
		for i := 0; i < retries; i++ {
			found, err = kubeclient.IsPersistentVolumeExist(e.Client, pvName, common.GetExpectedFluidAnnotations())
			if err != nil {
				return err
			}

			if found {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		if found {
			return fmt.Errorf("the PV %s is not cleaned up after 10-second retry",
				pvName)
		} else {
			e.Log.Info("the PV is deleted successfully", "name", pvName)
		}
	}
	return nil
}

func (e *CacheEngine) deleteFusePersistentVolumeClaim() error {
	found, err := kubeclient.IsPersistentVolumeClaimExist(e.Client, e.name, e.namespace, common.GetExpectedFluidAnnotations())
	if err != nil {
		return err
	}

	if found {
		err = kubeclient.DeletePersistentVolumeClaim(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}

		stillFound := false
		retries := 10
		for i := 0; i < retries; i++ {
			stillFound, err = kubeclient.IsPersistentVolumeClaimExist(e.Client, e.name, e.namespace, common.GetExpectedFluidAnnotations())
			if err != nil {
				return err
			}

			if !stillFound {
				break
			}

			should, err := kubeclient.ShouldRemoveProtectionFinalizer(e.Client, e.name, e.namespace)
			if err != nil {
				// ignore NotFound error and re-check existence if the pvc is already deleted
				if utils.IgnoreNotFound(err) == nil {
					continue
				}
			}

			if should {
				e.Log.Info("Should forcibly remove pvc-protection finalizer")
				err = kubeclient.RemoveProtectionFinalizer(e.Client, e.name, e.namespace)
				if err != nil {
					// ignore NotFound error and re-check existence if the pvc is already deleted
					if utils.IgnoreNotFound(err) == nil {
						continue
					}
					e.Log.Info("Failed to remove finalizers", "name", e.name, "namespace", e.namespace)
					return err
				}
			}

			time.Sleep(1 * time.Second)
		}

		if stillFound {
			return fmt.Errorf("the PVC %s in ns %s is not cleaned up after 10-second retry",
				e.name, e.namespace)
		} else {
			e.Log.Info("The PVC is deleted successfully",
				"name", e.name,
				"namespace", e.namespace)
		}
	}
	return nil
}
