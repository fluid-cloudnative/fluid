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
			if ContainsKey(oldLabels, labelToModifyKey) {
				err = fmt.Errorf("fail to add the label due to the label %s already exist", labelToModifyKey)
				return nil, err
			}
			node.Labels[labelToModifyKey] = labelToModifyValue
		case common.UpdateLabel:
			if !ContainsKey(oldLabels, labelToModifyKey) {
				err = fmt.Errorf("fail to add the label due to the label %s does not exist", labelToModifyKey)
				return nil, err
			}
			node.Labels[labelToModifyKey] = labelToModifyValue
		case common.DeleteLabel:
			if !ContainsKey(oldLabels, labelToModifyKey) {
				err = fmt.Errorf("fail to add the label due to the label %s does not exist", labelToModifyKey)
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

// ChangeNodeLabelWithPatchMode updates the input labels in PATCH mode.
func ChangeNodeLabelWithPatchMode(cli client.Client, node *v1.Node, labelsToModify common.LabelsToModify) (modifiedLabels []string, err error) {
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
	err = cli.Patch(context.TODO(), node, client.RawPatch(types.StrategicMergePatchType, patchByteData))
	if err != nil {
		return nil, errors.Wrapf(err, "patch node labels failed, node name: %s, labels: %v", node.Name, node.Labels)
	}
	return modifiedLabels, nil
}
