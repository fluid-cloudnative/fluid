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

package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
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
