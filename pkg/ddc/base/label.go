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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)
func (info *RuntimeInfo) GetLabelNameForMemory() string {
	return utils.GetLabelNameForMemory(info.IsDeprecatedNodeLabel(), info.runtimeType, info.namespace, info.name)
}

func (info *RuntimeInfo) GetLabelNameForDisk() string {
	return utils.GetLabelNameForDisk(info.IsDeprecatedNodeLabel(), info.runtimeType, info.namespace, info.name)
}

func (info *RuntimeInfo) GetLabelNameForTotal() string {
	return utils.GetLabelNameForTotal(info.IsDeprecatedNodeLabel(), info.runtimeType, info.namespace, info.name)
}

func (info *RuntimeInfo) GetCommonLabelName() string {
	return utils.GetCommonLabelName(info.IsDeprecatedNodeLabel(), info.namespace, info.name)
}

func (info *RuntimeInfo) GetRuntimeLabelName() string {
	return utils.GetRuntimeLabelName(info.IsDeprecatedNodeLabel(), info.namespace, info.name)
}

// GetDatasetNumLabelname get the label to record how much datasets on a node
func (info *RuntimeInfo) GetDatasetNumLabelName() string {
	return common.GetDatasetNumLabelName()
}

// GetFuseLabelName gets the label indicating a fuse running on some node.
func (info *RuntimeInfo) GetFuseLabelName() string {
	return common.LabelAnnotationFusePrefix + info.namespace + "-" + info.name
}
