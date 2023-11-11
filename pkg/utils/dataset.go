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

package utils

import (
	"context"
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// IsTargetPathUnderFluidNativeMounts checks if targetPath is a subpath under some given native mount point.
// We check this for the reason that native mount points need extra metadata sync alluxioOperations.
func IsTargetPathUnderFluidNativeMounts(targetPath string, dataset datav1alpha1.Dataset) bool {
	for _, mount := range dataset.Spec.Mounts {

		mPath := UFSPathBuilder{}.GenAlluxioMountPath(mount)

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
		m := UFSPathBuilder{}.GenAlluxioMountPath(mount)
		specMountPaths = append(specMountPaths, m)
	}
	for _, mount := range u.dataset.Status.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		m := UFSPathBuilder{}.GenAlluxioMountPath(mount)
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
