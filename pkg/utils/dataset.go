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
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PVCStorageAnnotation   = "pvc.fluid.io/resources.requests.storage"
	DefaultStorageCapacity = "100Pi"
)

// GetDataset gets the dataset.
// It returns a pointer to the dataset if successful.
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

func GetPVCStorageCapacityOfDataset(client client.Client, name, namespace string) (storageCapacity resource.Quantity, err error) {
	dataset, err := GetDataset(client, name, namespace)
	if err != nil {
		return storageCapacity, fmt.Errorf("failed to get dataset %s/%s: %w", namespace, name, err)
	}
	annotations := dataset.GetObjectMeta().GetAnnotations()
	if annotations == nil {
		storageCapacity = resource.MustParse(DefaultStorageCapacity)
		return
	}
	size := annotations[PVCStorageAnnotation]
	if size == "" {
		storageCapacity = resource.MustParse(DefaultStorageCapacity)
		return
	}

	storageCapacity, err = resource.ParseQuantity(size)
	if err != nil {
		log.Info(fmt.Sprintf("failed to parse storage capacity '%s', using default '%s': %v\n", size, DefaultStorageCapacity, err))
		return resource.MustParse(DefaultStorageCapacity), nil
	}
	return
}

// IsTargetPathUnderFluidNativeMounts checks if targetPath is a subpath under some given native mount point.
// We check this for the reason that native mount points need extra metadata sync alluxioOperations.
func IsTargetPathUnderFluidNativeMounts(targetPath string, dataset datav1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {

		mPath := UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)

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

// UFSToUpdate records the mountPath to change in virtual file system of dataset
type UFSToUpdate struct {
	toAdd    []string
	toRemove []string
	dataset  *datav1alpha1.Dataset
}

// NewUFSToUpdate get UFSToUpdate according the given dataset
func NewUFSToUpdate(ds *datav1alpha1.Dataset) *UFSToUpdate {
	return &UFSToUpdate{
		dataset: ds,
	}
}

// AnalyzePathsDelta analyze the ToAdd and ToRemove from the spec and mounted mountPaths of dataset
// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
func (u *UFSToUpdate) AnalyzePathsDelta() (specMountPaths, mountedMountPaths []string) {
	for _, mount := range u.dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		m := UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		specMountPaths = append(specMountPaths, m)
	}
	for _, mount := range u.dataset.Status.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		m := UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		mountedMountPaths = append(mountedMountPaths, m)
	}

	u.toAdd = SubtractString(specMountPaths, mountedMountPaths)
	u.toRemove = SubtractString(mountedMountPaths, specMountPaths)
	return
}

// ShouldUpdate check if needs to update the mount points according to ToAdd and ToRemove
func (u *UFSToUpdate) ShouldUpdate() bool {
	return len(u.toAdd) > 0 || len(u.toRemove) > 0
}

// ToAdd get the mountPaths to add into virtual file system of dataset
func (u *UFSToUpdate) ToAdd() []string {
	return u.toAdd
}

// ToRemove get the mountPaths to remove from virtual file system of dataset
func (u *UFSToUpdate) ToRemove() []string {
	return u.toRemove
}

// AddMountPaths add mounted path to ToAdd
func (u *UFSToUpdate) AddMountPaths(mountPaths []string) {
	if len(u.toAdd) == 0 {
		u.toAdd = mountPaths
		return
	}

	set := make(map[string]struct{}, len(u.toAdd))
	for _, i := range u.toAdd {
		set[i] = struct{}{}
	}

	for _, mountPath := range mountPaths {
		_, ok := set[mountPath]
		if !ok {
			u.toAdd = append(u.toAdd, mountPath)
		}
	}
}
