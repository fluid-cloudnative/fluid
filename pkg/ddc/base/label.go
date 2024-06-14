/*
Copyright 2020 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the Licensinfo.
You may obtain a copy of the License at

    http://www.apachinfo.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the Licensinfo.
*/

package base

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/common/deprecated"
)

func (info *RuntimeInfo) getStoragetLabelName(read common.ReadType, storage common.StorageType) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if info.IsDeprecatedNodeLabel() {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}
	return prefix +
		string(read) +
		info.runtimeType +
		"-" +
		string(storage) +
		info.namespace +
		"-" +
		info.name
}

func (info *RuntimeInfo) GetLabelNameForMemory() string {
	read := common.HumanReadType
	storage := common.MemoryStorageType
	if info.IsDeprecatedNodeLabel() {
		read = deprecated.HumanReadType
		storage = deprecated.MemoryStorageType
	}
	return info.getStoragetLabelName(read, storage)
}

func (info *RuntimeInfo) GetLabelNameForDisk() string {
	read := common.HumanReadType
	storage := common.DiskStorageType
	if info.IsDeprecatedNodeLabel() {
		read = deprecated.HumanReadType
		storage = deprecated.DiskStorageType
	}
	return info.getStoragetLabelName(read, storage)
}

func (info *RuntimeInfo) GetLabelNameForTotal() string {
	read := common.HumanReadType
	storage := common.TotalStorageType
	if info.IsDeprecatedNodeLabel() {
		read = deprecated.HumanReadType
		storage = deprecated.TotalStorageType
	}
	return info.getStoragetLabelName(read, storage)
}

func (info *RuntimeInfo) GetCommonLabelName() string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if info.IsDeprecatedNodeLabel() {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	return prefix + info.namespace + "-" + info.name
}

func (info *RuntimeInfo) GetRuntimeLabelName() string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if info.IsDeprecatedNodeLabel() {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	return prefix + info.runtimeType + "-" + info.namespace + "-" + info.name
}

// GetDatasetNumLabelname get the label to record how much datasets on a node
func (info *RuntimeInfo) GetDatasetNumLabelName() string {
	return common.GetDatasetNumLabelName()
}

// GetFuseLabelName gets the label indicating a fuse running on some node.
func (info *RuntimeInfo) GetFuseLabelName() string {
	return common.LabelAnnotationFusePrefix + info.namespace + "-" + info.name
}
