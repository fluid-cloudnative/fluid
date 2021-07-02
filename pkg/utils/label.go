package utils

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ChangeNodeLabelWithUpdateModel updates the input labels in UPDATE mode.
func ChangeNodeLabelWithUpdateModel(client client.Client, node *v1.Node, labelsToModify common.LabelsToModify) (modifiedLabels []string, err error) {
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

// ChangeNodeLabelWithPatchModel updates the input labels in PATCH mode.
func ChangeNodeLabelWithPatchModel(cli client.Client, node *v1.Node, labelsToModify common.LabelsToModify) (modifiedLabels []string, err error) {
	patchNode := node.DeepCopy()
	labels := labelsToModify.GetLabels()

	for _, labelToModify := range labels {
		oldLabels := patchNode.Labels
		operationType := labelToModify.GetOperationType()
		labelToModifyKey := labelToModify.GetLabelKey()
		labelToModifyValue := labelToModify.GetLabelValue()

		switch operationType {
		case common.AddLabel:
			if ContainsKey(oldLabels, labelToModifyKey) {
				err = fmt.Errorf("fail to add the label due to the label %s already exist", labelToModifyKey)
				return nil, err
			}
			patchNode.Labels[labelToModifyKey] = labelToModifyValue
		case common.UpdateLabel:
			if !ContainsKey(oldLabels, labelToModifyKey) {
				err = fmt.Errorf("fail to add the label due to the label %s does not exist", labelToModifyKey)
				return nil, err
			}
			patchNode.Labels[labelToModifyKey] = labelToModifyValue
		case common.DeleteLabel:
			if !ContainsKey(oldLabels, labelToModifyKey) {
				err = fmt.Errorf("fail to add the label due to the label %s does not exist", labelToModifyKey)
				return nil, err
			}
			delete(patchNode.Labels, labelToModifyKey)
		default:
			err = fmt.Errorf("fail to update the label due to the wrong operation: %s", operationType)
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModifyKey)
	}
	err = cli.Patch(context.TODO(), patchNode, client.MergeFrom(node))
	if err != nil {
		return nil, errors.Wrapf(err, "patch node labels failed, node name: %s, labels: %v", node.Name, node.Labels)
	}
	return modifiedLabels, nil
}
