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
	"k8s.io/apimachinery/pkg/util/validation"
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

func GetStorageLabelName(read common.ReadType, storage common.StorageType, isDeprecated bool, runtimeType string, namespace, name, ownerDatasetUID string) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if isDeprecated {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	prefix = prefix + string(read) + runtimeType + "-" + string(storage)

	return GetNamespacedNameValueWithPrefix(prefix, namespace, name, ownerDatasetUID)
}

func GetLabelNameForMemory(isDeprecated bool, runtimeType string, namespace, name, ownerDatasetUID string) string {
	read := common.HumanReadType
	storage := common.MemoryStorageType
	if isDeprecated {
		read = deprecated.HumanReadType
		storage = deprecated.MemoryStorageType
	}
	return GetStorageLabelName(read, storage, isDeprecated, runtimeType, namespace, name, ownerDatasetUID)
}

func GetLabelNameForDisk(isDeprecated bool, runtimeType string, namespace, name, ownerDatasetUID string) string {
	read := common.HumanReadType
	storage := common.DiskStorageType
	if isDeprecated {
		read = deprecated.HumanReadType
		storage = deprecated.DiskStorageType
	}
	return GetStorageLabelName(read, storage, isDeprecated, runtimeType, namespace, name, ownerDatasetUID)
}

func GetLabelNameForTotal(isDeprecated bool, runtimeType string, namespace, name, ownerDatasetUID string) string {
	read := common.HumanReadType
	storage := common.TotalStorageType
	if isDeprecated {
		read = deprecated.HumanReadType
		storage = deprecated.TotalStorageType
	}
	return GetStorageLabelName(read, storage, isDeprecated, runtimeType, namespace, name, ownerDatasetUID)
}

func GetCommonLabelName(isDeprecated bool, namespace, name, ownerDatasetUID string) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if isDeprecated {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	return GetNamespacedNameValueWithPrefix(prefix, namespace, name, ownerDatasetUID)
}

func GetRuntimeLabelName(isDeprecated bool, runtimeType string, namespace, name, ownerDatasetUID string) string {
	prefix := common.LabelAnnotationStorageCapacityPrefix
	if isDeprecated {
		prefix = deprecated.LabelAnnotationStorageCapacityPrefix
	}

	prefix = prefix + runtimeType + "-"

	return GetNamespacedNameValueWithPrefix(prefix, namespace, name, ownerDatasetUID)
}

func GetFuseLabelName(namespace, name, ownerDatasetUID string) string {
	return GetNamespacedNameValueWithPrefix(common.LabelAnnotationFusePrefix, namespace, name, ownerDatasetUID)
}

func GetExclusiveKey() string {
	return common.FluidExclusiveKey
}

// GetNamespacedNameValueWithPrefix Transfer a fully namespaced name with a prefix to a legal value which under max length limit.
// If the full namespaced name exceeds 63 characters, it calculates the hash value of the name and truncates the name and namespace,
// then appends the hash value to ensure the name's uniqueness and length constraint.
func GetNamespacedNameValueWithPrefix(prefix, namespace, name, ownerDatasetUID string) (fullNamespacedNameWithPrefix string) {
	namespacedName := fmt.Sprintf("%s-%s", namespace, name)
	fullNamespacedNameWithPrefix = fmt.Sprintf("%s%s", prefix, namespacedName)
	// ensure forward compatibility
	if len(fullNamespacedNameWithPrefix) < validation.DNS1035LabelMaxLength {
		return
	}

	if ownerDatasetUID == "" {
		log.Info("The ownerDatasetUID is absent, fall back to original value which causes the resource creation failed by scheme validation", "key", fmt.Sprintf("%s-%s", namespace, name))
		return fullNamespacedNameWithPrefix
	}

	fullNamespacedNameWithPrefix = fmt.Sprintf("%s%s", prefix, ownerDatasetUID)

	return
}

func GetDatasetId(namespace, name, ownerDatasetUID string) (fullNamespacedNameWithPrefix string) {
	return GetNamespacedNameValueWithPrefix("", namespace, name, ownerDatasetUID)
}

func PatchLabelToObjects(cli client.Client, labelKey, labelValue string, objs ...client.Object) (errs []error) {
	for _, obj := range objs {
		if obj.GetLabels() == nil ||
			!ContainsLabel(obj.GetLabels(), common.LabelAnnotationDatasetId, labelValue) {
			labelValuePair := map[string]string{
				labelKey: labelValue,
			}

			metadata := map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": labelValuePair,
				},
			}

			patchByteData, err := json.Marshal(metadata)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			err = cli.Patch(context.TODO(), obj, client.RawPatch(types.MergePatchType, patchByteData))
			if err != nil {
				log.Error(err, "failed to patch labels", "objKind", obj.GetObjectKind(), "name", obj.GetName(), "labels", obj.GetLabels())
				errs = append(errs, err)
				continue
			}
		}
	}
	return
}
