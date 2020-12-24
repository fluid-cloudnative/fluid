/*

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

import "github.com/fluid-cloudnative/fluid/pkg/common"

func (info *RuntimeInfo) GetStoragetLabelname(read common.ReadType, storage common.StorageType) string {
	return common.LabelAnnotationStorageCapacityPrefix +
		string(read) +
		info.runtimeType +
		"-" +
		string(storage) +
		info.namespace +
		"-" +
		info.name
}

func (info *RuntimeInfo) GetCommonLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + info.namespace + "-" + info.name
}

func (info *RuntimeInfo) GetRuntimeLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + info.runtimeType + "-" + info.namespace + "-" + info.name
}

func (info *RuntimeInfo) GetRuntimeExclusivenessLabelname() string {
	return common.LabelAnnotationStorageCapacityPrefix + info.runtimeType + "-" + common.Exclusiveness
}
