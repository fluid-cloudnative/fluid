/*

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

package common

const (
	// LabelAnnotationPrefix is the prefix of every labels and annotations added by the controller.
	LabelAnnotationPrefix = "fluid.io/"
	// The format is fluid.io/s-{runtime_type}-{data_set_name}, s means storage
	LabelAnnotationStorageCapacityPrefix = LabelAnnotationPrefix + "s-"
	// The dataset annotation
	LabelAnnotationDataset = LabelAnnotationPrefix + "dataset"
	// LabelAnnotationDatasetNum indicates the number of the dataset in specific node
	LabelAnnotationDatasetNum = LabelAnnotationPrefix + "dataset-num"

	// fluid adminssion webhook inject pod affinity strategy flag
	LabelFluidSchedulingStrategyFlag = LabelAnnotationPrefix + "enable-scheduling-strategy"
)

func GetDatasetNumLabelName() string {
	return LabelAnnotationDatasetNum
}

func HitTarget(m map[string]string, target string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[target]; ok {
		return v == TrueStr
	}
	return false
}
