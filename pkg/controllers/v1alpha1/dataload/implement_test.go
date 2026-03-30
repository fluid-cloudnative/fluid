/*
Copyright 2020 The Fluid Authors.

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

package dataload

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

var _ = Describe("IsTargetPathUnderFluidNativeMounts", func() {
	const (
		datasetName    = "imagenet"
		imageNetPath   = "/imagenet"
		pvcMountPoint  = "pvc://nfs-imagenet"
		emptyMountPath = ""
	)

	mockDataset := func(name, mountPoint, path string) v1alpha1.Dataset {
		return v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{
					{
						Name:       name,
						MountPoint: mountPoint,
						Path:       path,
					},
				},
			},
		}
	}

	DescribeTable("path matching",
		func(targetPath string, dataset v1alpha1.Dataset, want bool) {
			got := utils.IsTargetPathUnderFluidNativeMounts(targetPath, dataset)
			Expect(got).To(Equal(want))
		},
		Entry("non-fluid-native OSS mount returns false",
			imageNetPath,
			mockDataset(datasetName, "oss://imagenet-data/", emptyMountPath),
			false,
		),
		Entry("PVC mount returns true",
			imageNetPath,
			mockDataset(datasetName, pvcMountPoint, emptyMountPath),
			true,
		),
		Entry("hostpath mount returns true",
			imageNetPath,
			mockDataset(datasetName, "local:///hostpath_imagenet", emptyMountPath),
			true,
		),
		Entry("target subpath under PVC mount returns true",
			"/imagenet/data/train",
			mockDataset(datasetName, pvcMountPoint, emptyMountPath),
			true,
		),
		Entry("target path under mount path returns true",
			"/dataset/data/train",
			mockDataset(datasetName, pvcMountPoint, "/dataset"),
			true,
		),
		Entry("non-matching path returns false",
			"/dataset",
			mockDataset(datasetName, pvcMountPoint, emptyMountPath),
			false,
		),
	)
})
