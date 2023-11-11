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
