package utils

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LabelToModify modifies the labelKey in operationType.
type LabelToModify struct {
	LabelKey      string
	LabelValue    string
	OperationType v1alpha1.OperationType
}

// ChangeNodeLabelWithUpdateModel updates the input labels in UPDATE mode.
func ChangeNodeLabelWithUpdateModel(client client.Client, node *v1.Node, labelsToModify []LabelToModify) (modifiedLabels []string, err error) {
	for _, labelToModify := range labelsToModify {
		if labelToModify.OperationType == v1alpha1.AddLabel || labelToModify.OperationType == v1alpha1.UpdateLabel {
			node.Labels[labelToModify.LabelKey] = labelToModify.LabelValue
		} else if labelToModify.OperationType == v1alpha1.DeleteLabel {
			delete(node.Labels, labelToModify.LabelKey)
		} else {
			err = fmt.Errorf("fail to update the label due to the wrong operation")
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModify.LabelKey)
	}
	err = client.Update(context.TODO(), node)
	if err != nil {
		log.Error(err, "LabelCachedNodes")
		return nil, err
	}
	return modifiedLabels, nil
}

// ChangeNodeLabelWithPatchModel updates the input labels in PATCH mode.
func ChangeNodeLabelWithPatchModel(cli client.Client, node *v1.Node, labelsToModify []LabelToModify) (modifiedLabels []string, err error) {
	patchNode := node.DeepCopy()

	for _, labelToModify := range labelsToModify {
		if labelToModify.OperationType == v1alpha1.AddLabel || labelToModify.OperationType == v1alpha1.UpdateLabel {
			patchNode.Labels[labelToModify.LabelKey] = labelToModify.LabelValue
		} else if labelToModify.OperationType == v1alpha1.DeleteLabel {
			delete(patchNode.Labels, labelToModify.LabelKey)
		} else {
			err = fmt.Errorf("fail to update the label due to the wrong operation")
			return nil, err
		}
		modifiedLabels = append(modifiedLabels, labelToModify.LabelKey)
	}
	err = cli.Patch(context.TODO(), patchNode, client.MergeFrom(node))
	if err != nil {
		log.Error(err, "LabelCachedNodes")
		return nil, err
	}
	return modifiedLabels, nil
}

// AddLabelToModifyToSlice creates new struct LabelToModify with input params and adds it into the slice.
func AddLabelToModifyToSlice(labelKey string, labelValue string, operationType v1alpha1.OperationType, labelsToModify *[]LabelToModify) {
	newLabelToModify := LabelToModify{
		LabelKey:      labelKey,
		OperationType: operationType,
	}
	if operationType != v1alpha1.DeleteLabel {
		newLabelToModify.LabelValue = labelValue
	}
	*labelsToModify = append(*labelsToModify, newLabelToModify)
}
