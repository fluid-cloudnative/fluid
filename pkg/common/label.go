/*
Copyright 2021 The Fluid Authors.

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

import "regexp"

const (
	// LabelAnnotationPrefix is the prefix of every labels and annotations added by the controller.
	LabelAnnotationPrefix = "fluid.io/"
	// The format is fluid.io/s-{runtime_type}-{data_set_name}, s means storage
	LabelAnnotationStorageCapacityPrefix = LabelAnnotationPrefix + "s-"
	// LabelAnnotationFusePrefix is the prefix for the fuse annotation. The annotation follows
	// fluid.io/f-{runtime type}-{dataset name}, in which f means fuse
	LabelAnnotationFusePrefix = LabelAnnotationPrefix + "f-"
	// The dataset annotation
	LabelAnnotationDataset = LabelAnnotationPrefix + "dataset"
	// LabelAnnotationDatasetNum indicates the number of the dataset in specific node
	LabelAnnotationDatasetNum = LabelAnnotationPrefix + "dataset-num"

	// LabelAnnotationManagedByDeprecated is a deprecated label key for LabelAnnotationManagedBy
	LabelAnnotationManagedByDeprecated = LabelAnnotationPrefix + "wrapped-by"

	// LabelAnnotationManagedBy indicates a pvc that is managed by Fluid
	LabelAnnotationManagedBy = LabelAnnotationPrefix + "managed-by"

	// fluid adminssion webhook inject flag
	EnableFluidInjectionFlag = LabelAnnotationPrefix + "enable-injection"

	// use two lables for name and namespace
	LabelAnnotationDatasetReferringName      = LabelAnnotationDataset + ".referring-name"
	LabelAnnotationDatasetReferringNameSpace = LabelAnnotationDataset + ".referring-namespace"

	RuntimeControllerReplicas = "controller.runtime." + LabelAnnotationPrefix + "replicas"

	// LabelNodePublishMothod is a pv label that indicates the method nodePuhlishVolume use
	LabelNodePublishMothod = LabelAnnotationPrefix + "node-publish-method"
)

var (
	// LabelAnnotationPodSchedRegex is the fluid cache label for scheduling pod, format: 'fluid.io/dataset.{dataset name}.sched]'
	// use string literal to meet security check.
	LabelAnnotationPodSchedRegex = regexp.MustCompile("^fluid\\.io/dataset\\.([A-Za-z0-9.-]*)\\.sched$")
)

type OperationType string

const (
	// AddLabel means adding a new label on the specific node.
	AddLabel OperationType = "Add"

	// DeleteLabel means deleting the label of the specific node.
	DeleteLabel OperationType = "Delete"

	// UpdateLabel means updating the label value of the specific node.
	UpdateLabel OperationType = "UpdateValue"
)

// LabelToModify modifies the labelKey in operationType.
type LabelToModify struct {
	labelKey      string
	labelValue    string
	operationType OperationType
}

func (labelToModify *LabelToModify) GetLabelKey() string {
	return labelToModify.labelKey
}

func (labelToModify *LabelToModify) GetLabelValue() string {
	return labelToModify.labelValue
}

func (labelToModify *LabelToModify) GetOperationType() OperationType {
	return labelToModify.operationType
}

type LabelsToModify struct {
	labels []LabelToModify
}

func (labels *LabelsToModify) GetLabels() []LabelToModify {
	return labels.labels
}

func (labels *LabelsToModify) operator(labelKey string, labelValue string, operationType OperationType) {
	newLabelToModify := LabelToModify{
		labelKey:      labelKey,
		operationType: operationType,
	}
	if operationType != DeleteLabel {
		newLabelToModify.labelValue = labelValue
	}
	labels.labels = append(labels.labels, newLabelToModify)
}

func (labels *LabelsToModify) Add(labelKey string, labelValue string) {
	labels.operator(labelKey, labelValue, AddLabel)
}

func (labels *LabelsToModify) Update(labelKey string, labelValue string) {
	labels.operator(labelKey, labelValue, UpdateLabel)
}

func (labels *LabelsToModify) Delete(labelKey string) {
	labels.operator(labelKey, "", DeleteLabel)
}

func GetDatasetNumLabelName() string {
	return LabelAnnotationDatasetNum
}

// Check if the key has the expected value
func CheckExpectValue(m map[string]string, key string, targetValue string) bool {
	if len(m) == 0 {
		return false
	}
	if v, ok := m[key]; ok {
		return v == targetValue
	}
	return false
}

func GetManagerDatasetFromLabels(labels map[string]string) (datasetName string, exists bool) {
	datasetName, exists = labels[LabelAnnotationManagedBy]
	if exists {
		return
	}

	// fallback to check deprecated "fluid.io/wrapped-by" label
	datasetName, exists = labels[LabelAnnotationManagedByDeprecated]
	return
}
