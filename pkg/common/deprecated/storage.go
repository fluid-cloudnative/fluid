/*
Copyright 2023 The Fluid Authors.

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

package deprecated

import "github.com/fluid-cloudnative/fluid/pkg/common"

const (
	HumanReadType common.ReadType = "human-"

	// rawReadType readType = "raw-"
)

const (
	MemoryStorageType common.StorageType = "mem-"

	DiskStorageType common.StorageType = "disk-"

	TotalStorageType common.StorageType = "total-"
)

const (
	LabelAnnotationPrefix = "data.fluid.io/"
	// The format is data.fluid.io/storage-{runtime_type}-{data_set_name}
	LabelAnnotationStorageCapacityPrefix = LabelAnnotationPrefix + "storage-"
)
