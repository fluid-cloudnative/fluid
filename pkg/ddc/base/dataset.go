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

package base

import (
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	transformerutils "github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func GetDatasetRefName(name, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func GetPhysicalDatasetFromMounts(mounts []datav1alpha1.Mount) []types.NamespacedName {
	// virtual dataset can only mount dataset
	var physicalNamespacedName []types.NamespacedName
	for _, mount := range mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
			namespaceAndName := strings.Split(datasetPath, "/")
			if len(namespaceAndName) >= 2 {
				physicalNamespacedName = append(physicalNamespacedName, types.NamespacedName{
					Namespace: namespaceAndName[0],
					Name:      namespaceAndName[1],
				})
			}
		}
	}
	return physicalNamespacedName
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

func GetOwnerDatasetUIDFromRuntimeMeta(metaObj metav1.ObjectMeta) (types.UID, error) {
	datasetOwners := transformerutils.FilterOwnerByKind(metaObj.GetOwnerReferences(), datav1alpha1.Datasetkind)
	if len(datasetOwners) > 0 {
		ownerPath := field.NewPath("metadata").Child("ownerReferences")
		if len(datasetOwners) > 1 {
			// num of dataset owners should be at most 1
			return "", fmt.Errorf("found multiple Dataset owners in %s, something went wrong", ownerPath.String())
		}

		if datasetOwners[0].Name != metaObj.GetName() {
			return "", fmt.Errorf("owner Dataset of the runtime in %s has different name with runtime, expected to be same", ownerPath.String())
		}

		return datasetOwners[0].UID, nil
	}

	return "", nil
}
