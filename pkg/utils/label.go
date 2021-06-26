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
	for _, labelToModify := range labelsToModify.Labels {
		switch labelToModify.OperationType {
		case common.AddLabel, common.UpdateLabel:
			node.Labels[labelToModify.LabelKey] = labelToModify.LabelValue
		case common.DeleteLabel:
			delete(node.Labels, labelToModify.LabelKey)
		default:
			err = fmt.Errorf("fail to update the label due to the wrong operation: %s", labelToModify.OperationType)
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModify.LabelKey)
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

	for _, labelToModify := range labelsToModify.Labels {
		switch labelToModify.OperationType {
		case common.AddLabel, common.UpdateLabel:
			patchNode.Labels[labelToModify.LabelKey] = labelToModify.LabelValue
		case common.DeleteLabel:
			delete(patchNode.Labels, labelToModify.LabelKey)
		default:
			err = fmt.Errorf("fail to update the label due to the wrong operation: %s", labelToModify.OperationType)
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModify.LabelKey)
	}
	err = cli.Patch(context.TODO(), patchNode, client.MergeFrom(node))
	if err != nil {
		return nil, errors.Wrapf(err, "patch node labels failed, node name: %s, labels: %v", node.Name, node.Labels)
	}
	return modifiedLabels, nil
}
