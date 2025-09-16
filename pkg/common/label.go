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
	// LabelAnnotationPrefix is the prefix of every label and annotations added by the controller.
	LabelAnnotationPrefix = "fluid.io/"

	// LabelAnnotationStorageCapacityPrefix is the prefix for the storage annotaion.
	// i.e. fluid.io/s-{runtime_type}-{data_set_name}, in which s means storage
	LabelAnnotationStorageCapacityPrefix = LabelAnnotationPrefix + "s-"

	// LabelAnnotationFusePrefix is the prefix for the fuse annotation. The annotation follows
	// i.e. fluid.io/f-{runtime type}-{dataset name}, in which f means fuse
	LabelAnnotationFusePrefix = LabelAnnotationPrefix + "f-"

	// The dataset annotation
	// i.e. fluid.io/dataset
	LabelAnnotationDataset = LabelAnnotationPrefix + "dataset"

	// LabelAnnotationDatasetId indicates the uuid of the dataset
	// i.e. fluid.io/dataset-id
	LabelAnnotationDatasetId = LabelAnnotationDataset + "-id"

	// LabelAnnotationDatasetNum indicates the number of the dataset in specific node
	// i.e. fluid.io/dataset-num
	LabelAnnotationDatasetNum = LabelAnnotationPrefix + "dataset-num"

	// LabelAnnotationManagedByDeprecated is a deprecated label key for LabelAnnotationManagedBy
	// i.e. fluid.io/wrapped-by
	LabelAnnotationManagedByDeprecated = LabelAnnotationPrefix + "wrapped-by"

	// LabelAnnotationManagedBy indicates a resource(like pvc) that is managed by Fluid
	// i.e. fluid.io/managed-by
	LabelAnnotationManagedBy = LabelAnnotationPrefix + "managed-by"

	// LabelAnnotationCopyFrom indicates a resource that is copied from another resource
	// i.e. fluid.io/copied-from
	LabelAnnotationCopyFrom = LabelAnnotationPrefix + "copied-from"

	// fluid adminssion webhook inject flag
	// i.e. fluid.io/enable-injection
	EnableFluidInjectionFlag = LabelAnnotationPrefix + "enable-injection"

	// use two lables for name and namespace
	// i.e. fluid.io/dataset.referring-name
	LabelAnnotationDatasetReferringName = LabelAnnotationDataset + ".referring-name"
	// i.e. fluid.io/dataset.referring-namespace
	LabelAnnotationDatasetReferringNameSpace = LabelAnnotationDataset + ".referring-namespace"

	// LabelNodePublishMethod is a pv label that indicates the method nodePuhlishVolume use
	// i.e. fluid.io/node-publish-method
	LabelNodePublishMethod = LabelAnnotationPrefix + "node-publish-method"

	// AnnotationSkipCheckMountReadyTarget is a runtime annotation that indicates if the fuse mount related with this runtime is ready should be checked in nodePuhlishVolume
	// i.e. key: fluid.io/skip-check-mount-ready-target
	//      value:
	//   	"": Skip none,
	//      "All": Skill all mount mode to check mount ready,
	//   	"MountPod": for only mountPod to skip check mount ready,
	//   	"Sidecar": for only sidecar to skip check mount ready,
	AnnotationSkipCheckMountReadyTarget = LabelAnnotationPrefix + "skip-check-mount-ready-target"

	// AnnotationDisableRuntimeHelmValueConfig is a runtime label indicates the configmap contains helm value will not be created in setup.
	AnnotationDisableRuntimeHelmValueConfig = "runtime." + LabelAnnotationPrefix + "disable-helm-value-config"

	// LabelAnnotationMountingDatasets is a label/annotation key indicating which datasets are currently being used by a pod.
	// i.e. fluid.io/datasets-in-use
	LabelAnnotationDatasetsInUse = LabelAnnotationPrefix + "datasets-in-use"

	// i.e. fuse.runtime.fluid.io/generation
	LabelRuntimeFuseGeneration = "fuse.runtime." + LabelAnnotationPrefix + "generation"
)

const (
	// i.e. controller.runtime.fluid.io/replicas
	RuntimeControllerReplicas = "controller.runtime." + LabelAnnotationPrefix + "replicas"

	// i.e. prometheus.fuse.fluid.io/scrape
	AnnotationPrometheusFuseMetricsScrapeKey = "prometheus.fuse." + LabelAnnotationPrefix + "scrape"

	// i.e. container-dataset-mapping.sidecar.fluid.io/
	LabelContainerDatasetMappingKeyPrefix = "container-dataset-mapping.sidecar." + LabelAnnotationPrefix

	// AnnotationDataFlowAffinityScopePrefix is an annotation prefix representing dataflow affinity related functions.
	// i.e. affinity.dataflow.fluid.io/
	AnnotationDataFlowAffinityScopePrefix = "affinity.dataflow." + LabelAnnotationPrefix
	// AnnotationDataFlowAffinityInject is an annotation representing enabled the dataflow affinity injection, for internal use.
	// i.e. affinity.dataflow.fluid.io/inject
	AnnotationDataFlowAffinityInject = AnnotationDataFlowAffinityScopePrefix + "inject"
	// AnnotationDataFlowAffinityLabelsName is an annotation key name for exposed affinity labels for an operation in a dataflow.
	// i.e. affinity.dataflow.fluid.io/labels
	AnnotationDataFlowAffinityLabelsName = AnnotationDataFlowAffinityScopePrefix + "labels"

	// AnnotationDataFlowCustomizedAffinityPrefix is a prefix used to
	// i.e. affinity.dataflow.fluid.io.
	AnnotationDataFlowCustomizedAffinityPrefix = "affinity.dataflow.fluid.io."
)

const (
	// AnnotationServerlessPlatform is an annotation key name for the platform type of serverless.
	// i.e. serverless.fluid.io/platform
	AnnotationServerlessPlatform = "serverless." + LabelAnnotationPrefix + "platform"
)

var (
	// LabelAnnotationPodSchedRegex is the fluid cache label for scheduling pod, format: 'fluid.io/dataset.{dataset name}.sched]'
	// use string literal to meet security check.
	LabelAnnotationPodSchedRegex = regexp.MustCompile(`^fluid\.io/dataset\.([A-Za-z0-9.-]*)\.sched$`)
)

const (
	// RuntimePodType is the label key for runtime pod type
	RuntimePodType   = "fluid.io/runtime-pod-type"
	RuntimeWorkerPod = "worker"

	// AnnotationRuntimeName is the annotation key for the runtime name
	AnnotationRuntimeName = LabelAnnotationPrefix + "runtime-name"
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
