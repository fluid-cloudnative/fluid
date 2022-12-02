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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

func GetDatasetRefName(name, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func GetMountedDatasetNamespacedName(virtualDataset *datav1alpha1.Dataset) []types.NamespacedName {
	// virtual dataset can only mount dataset
	var physicalNameSpacedName []types.NamespacedName
	for _, mount := range virtualDataset.Spec.Mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
			namespaceAndName := strings.Split(datasetPath, "/")
			if len(namespaceAndName) == 2 {
				physicalNameSpacedName = append(physicalNameSpacedName, types.NamespacedName{
					Namespace: namespaceAndName[0],
					Name:      namespaceAndName[1],
				})
			}
		}
	}
	return physicalNameSpacedName
}
