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

package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/common/deprecated"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ChangeNodeLabelWithUpdateMode updates the input labels in UPDATE mode.
func ChangeNodeLabelWithUpdateMode(client client.Client, node *v1.Node, labelsToModify common.LabelsToModify) (modifiedLabels []string, err error) {
	labels := labelsToModify.GetLabels()

	for _, labelToModify := range labels {
		oldLabels := node.Labels
		operationType := labelToModify.GetOperationType()
		labelToModifyKey := labelToModify.GetLabelKey()
		labelToModifyValue := labelToModify.GetLabelValue()

		switch operationType {
		case common.AddLabel:
			if _, exists := oldLabels[labelToModifyKey]; exists {
				err = fmt.Errorf("fail to add the label due to the label %s already exist", labelToModifyKey)
				return nil, err
			}
			node.Labels[labelToModifyKey] = labelToModifyValue
		case common.UpdateLabel:
			if _, exists := oldLabels[labelToModifyKey]; !exists {
				err = fmt.Errorf("fail to update the label due to the label %s does not exist", labelToModifyKey)
				return nil, err
			}
			node.Labels[labelToModifyKey] = labelToModifyValue
		case common.DeleteLabel:
			if _, exists := oldLabels[labelToModifyKey]; !exists {
				err = fmt.Errorf("fail to delete the label due to the label %s does not exist", labelToModifyKey)
				return nil, err
			}
			delete(node.Labels, labelToModifyKey)
		default:
			err = fmt.Errorf("fail to update the label due to the wrong operation: %s", operationType)
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModifyKey)
	}
	err = client.Update(context.TODO(), node)
	if err != nil {
		return nil, errors.Wrapf(err, "update node labels failed, node name: %s, labels: %v", node.Name, node.Labels)
	}
	return modifiedLabels, nil
}

func PatchLabels(cli client.Client, obj client.Object, labelsToModify common.LabelsToModify) (modifiedLabels []string, err error) {
	labels := labelsToModify.GetLabels()
	labelValuePair := map[string]interface{}{}

	for _, labelToModify := range labels {
		operationType := labelToModify.GetOperationType()
		labelToModifyKey := labelToModify.GetLabelKey()
		labelToModifyValue := labelToModify.GetLabelValue()

		switch operationType {
		case common.AddLabel, common.UpdateLabel:
			labelValuePair[labelToModifyKey] = labelToModifyValue
		case common.DeleteLabel:
			labelValuePair[labelToModifyKey] = nil
		default:
			err = fmt.Errorf("fail to update the label due to the wrong operation: %s", operationType)
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModifyKey)
	}

	metadata := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labelValuePair,
		},
	}

	patchByteData, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	err = cli.Patch(context.TODO(), obj, client.RawPatch(types.StrategicMergePatchType, patchByteData))
	if err != nil {
		return nil, errors.Wrapf(err, "patch node labels failed, node name: %s, labels: %v", obj.GetName(), obj.GetLabels())
	}
	return modifiedLabels, nil
}

// ChangeNodeLabelWithPatchMode updates the input labels in PATCH mode.
func ChangeNodeLabelWithPatchMode(cli client.Client, node *v1.Node, labelsToModify common.LabelsToModify) (modifiedLabels []string, err error) {
	return PatchLabels(cli, node, labelsToModify)
}

func GetStoragetLabelName(read common.ReadType, storage common.StorageType, isDeprecated bool, runtimeType string, namespace string, name string) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if isDeprecated {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}
	return prefix +
		string(read) +
		runtimeType +
		"-" +
		string(storage) +
		namespace +
		"-" +
		name
}

func GetLabelNameForMemory(isDeprecated bool, runtimeType string, namespace string, name string) string {
	read := common.HumanReadType
	storage := common.MemoryStorageType
	if isDeprecated {
		read = deprecated.HumanReadType
		storage = deprecated.MemoryStorageType
	}
	return GetStoragetLabelName(read, storage, isDeprecated, runtimeType, namespace, name)
}

func GetLabelNameForDisk(isDeprecated bool, runtimeType string, namespace string, name string) string {
	read := common.HumanReadType
	storage := common.DiskStorageType
	if isDeprecated {
		read = deprecated.HumanReadType
		storage = deprecated.DiskStorageType
	}
	return GetStoragetLabelName(read, storage, isDeprecated, runtimeType, namespace, name)
}

func GetLabelNameForTotal(isDeprecated bool, runtimeType string, namespace string, name string) string {
	read := common.HumanReadType
	storage := common.TotalStorageType
	if isDeprecated {
		read = deprecated.HumanReadType
		storage = deprecated.TotalStorageType
	}
	return GetStoragetLabelName(read, storage, isDeprecated, runtimeType, namespace, name)
}

func GetCommonLabelName(isDeprecated bool, namespace string, name string) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if isDeprecated {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	return prefix + namespace + "-" + name
}

func GetRuntimeLabelName(isDeprecated bool, namespace string, name string) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if isDeprecated {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	return prefix + namespace + "-" + name
}
