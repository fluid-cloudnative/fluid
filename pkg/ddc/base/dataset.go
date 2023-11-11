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

package base

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
)

func GetDatasetRefName(name, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func GetPhysicalDatasetFromMounts(mounts []datav1alpha1.Mount) []types.NamespacedName {
	// virtual dataset can only mount dataset
	var physicalNameSpacedName []types.NamespacedName
	for _, mount := range mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
			namespaceAndName := strings.Split(datasetPath, "/")
			if len(namespaceAndName) >= 2 {
				physicalNameSpacedName = append(physicalNameSpacedName, types.NamespacedName{
					Namespace: namespaceAndName[0],
					Name:      namespaceAndName[1],
				})
			}
		}
	}
	return physicalNameSpacedName
}

func GetPhysicalDatasetSubPath(virtualDataset *datav1alpha1.Dataset) []string {
	var paths []string
	for _, mount := range virtualDataset.Spec.Mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
			splitsStrings := strings.SplitAfterN(datasetPath, "/", 3)
			if len(splitsStrings) == 3 {
				paths = append(paths, splitsStrings[2])
			}
		}
	}
	return paths
}

func CheckReferenceDataset(dataset *datav1alpha1.Dataset) (check bool, err error) {
	mounts := len(GetPhysicalDatasetFromMounts(dataset.Spec.Mounts))
	totalMounts := len(dataset.Spec.Mounts)
	switch {
	case mounts == 1:
		if totalMounts == mounts {
			check = true
		} else {
			err = fmt.Errorf("the dataset is not validated, since it has 1 dataset mounts but also contains other types of mounts %v", dataset.Spec.Mounts)
		}
	case mounts > 1:
		err = fmt.Errorf("the dataset is not validated, since it has %v dataset mounts which only expects 1", mounts)
	}

	return
}
