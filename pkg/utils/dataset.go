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

package utils

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//GetDataset gets the dataset.
//It returns a pointer to the dataset if successful.
func GetDataset(client client.Client, name, namespace string) (*datav1alpha1.Dataset, error) {

	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	var dataset datav1alpha1.Dataset
	if err := client.Get(context.TODO(), key, &dataset); err != nil {
		return nil, err
	}
	return &dataset, nil
}

// checks the setup is done
func IsSetupDone(dataset *datav1alpha1.Dataset) (done bool) {
	index, _ := GetDatasetCondition(dataset.Status.Conditions, datav1alpha1.DatasetReady)
	if index != -1 {
		// e.Log.V(1).Info("The runtime is already setup.")
		done = true
	}

	return
}

func GetAccessModesOfDataset(client client.Client, name, namespace string) (accessModes []corev1.PersistentVolumeAccessMode, err error) {
	dataset, err := GetDataset(client, name, namespace)
	if err != nil {
		return accessModes, err
	}

	accessModes = dataset.Spec.AccessModes
	if len(accessModes) == 0 {
		accessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadOnlyMany,
		}
	}

	return accessModes, err

}

// IsTargetPathUnderFluidNativeMounts checks if targetPath is a subpath under some given native mount point.
// We check this for the reason that native mount points need extra metadata sync alluxioOperations.
func IsTargetPathUnderFluidNativeMounts(targetPath string, dataset datav1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {

		mPath := UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)

		//TODO(xuzhihao): HasPrefix is not enough.

		// only for native scheme
		if !common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}

		if IsSubPath(mPath, targetPath) {
			return true
		}
	}

	return false
}

// GetUpdateUFSMap get a map of mount points to be added and removed
func GetUpdateUFSMap(dataset *datav1alpha1.Dataset) (map[string][]string, error) {
	updateUFSMap := make(map[string][]string)
	// 1. get spec MountPaths and mounted MountPaths from dataset
	specMountPaths, mountedMountPaths, err := getMounts(dataset)
	if err != nil {
		return updateUFSMap, err
	}

	// 2. get mount points which are needed to be added/removed
	added, removed := calculateMountPointsChanges(specMountPaths, mountedMountPaths)
	if len(added) > 0 {
		updateUFSMap["added"] = added
	}

	if len(removed) > 0 {
		updateUFSMap["removed"] = removed
	}

	return updateUFSMap, err
}

// getMounts get the spec and current mounted mountPaths of dataset
// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
func getMounts(dataset *datav1alpha1.Dataset) (specMountPaths, mountedMountPaths []string, err error) {
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		alluxioPathInCtx := UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Spec.Mounts)
		specMountPaths = append(specMountPaths, alluxioPathInCtx)
	}
	for _, mount := range dataset.Status.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		alluxioPathHaveMounted := UFSPathBuilder{}.GenAlluxioMountPath(mount, dataset.Status.Mounts)
		mountedMountPaths = append(mountedMountPaths, alluxioPathHaveMounted)
	}

	return specMountPaths, mountedMountPaths, err
}

// calculateMountPointsChanges compare the spec and status of a dataset to calculate MountPoints Changes
func calculateMountPointsChanges(specMountPaths, mountedMountPaths []string) (add, remove []string) {
	add = funk.SubtractString(specMountPaths, mountedMountPaths)
	remove = funk.SubtractString(mountedMountPaths, specMountPaths)

	return
}

// UpdateMountStatus updates the mount status of the dataset according to the given phase
func UpdateMountStatus(client client.Client, name string, namespace string, phase datav1alpha1.DatasetPhase) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := GetDataset(client, name, namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()
		var cond datav1alpha1.DatasetCondition

		switch phase {
		case datav1alpha1.UpdatingDatasetPhase:
			cond = NewDatasetCondition(datav1alpha1.DatasetUpdating, datav1alpha1.DatasetUpdatingReason,
				"The ddc runtime is updating.",
				corev1.ConditionTrue)
		case datav1alpha1.BoundDatasetPhase:
			datasetToUpdate.Status.Mounts = datasetToUpdate.Spec.Mounts
			cond = NewDatasetCondition(datav1alpha1.DatasetReady, datav1alpha1.DatasetReadyReason,
				"The ddc runtime has updated completely.",
				corev1.ConditionFalse)
		default:
			return fmt.Errorf("cannot change dataset phase to %s", phase)
		}

		datasetToUpdate.Status.Phase = phase
		datasetToUpdate.Status.Conditions = UpdateDatasetCondition(datasetToUpdate.Status.Conditions, cond)

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			err = client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
