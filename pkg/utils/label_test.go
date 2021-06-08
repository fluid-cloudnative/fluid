package utils

import (
	"context"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestChangeNodeLabelWithUpdateModel(t *testing.T) {
	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				Labels: map[string]string{
					"datasetNum":  "2",
					"deleteLabel": "true",
				},
			},
		},
	}

	testNodes := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		testNodes = append(testNodes, nodeInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	var testCase = []struct {
		node          v1.Node
		labelToModify []LabelToModify
		wantedNode    v1.Node
	}{
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"datasetNum":  "2",
						"deleteLabel": "true",
					},
				},
			},
			labelToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					LabelValue:    "true",
					OperationType: v1alpha1.AddLabel,
				},
				{
					LabelKey:      "datasetNum",
					LabelValue:    "1",
					OperationType: v1alpha1.UpdateLabel,
				},
				{
					LabelKey:      "deleteLabel",
					OperationType: v1alpha1.DeleteLabel,
				},
			},
			wantedNode: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"commonLabel": "true",
						"datasetNum":  "1",
					},
				},
			},
		},
	}

	for _, test := range testCase {
		_, err := ChangeNodeLabelWithUpdateModel(client, &test.node, test.labelToModify)
		if err != nil {
			t.Errorf("fail to add label to modify to slice")
		}
		updatedNode := &v1.Node{}
		key := types.NamespacedName{
			Name: "test-node",
		}
		err = client.Get(context.TODO(), key, updatedNode)
		if err != nil {
			t.Errorf("fail to update label to modify to slice")
		}
		if !reflect.DeepEqual((*updatedNode).Labels, test.wantedNode.Labels) {
			t.Errorf("fail to add label to modify to slice")
		}
	}
}

func TestChangeNodeLabelWithPatchModel(t *testing.T) {
	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				Labels: map[string]string{
					"datasetNum":  "2",
					"deleteLabel": "true",
				},
			},
		},
	}

	testNodes := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		testNodes = append(testNodes, nodeInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	var testCase = []struct {
		node          v1.Node
		labelToModify []LabelToModify
		wantedNode    v1.Node
	}{
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"datasetNum":  "2",
						"deleteLabel": "true",
					},
				},
			},
			labelToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					LabelValue:    "true",
					OperationType: v1alpha1.AddLabel,
				},
				{
					LabelKey:      "datasetNum",
					LabelValue:    "1",
					OperationType: v1alpha1.UpdateLabel,
				},
				{
					LabelKey:      "deleteLabel",
					OperationType: v1alpha1.DeleteLabel,
				},
			},
			wantedNode: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"commonLabel": "true",
						"datasetNum":  "1",
					},
				},
			},
		},
	}

	for _, test := range testCase {
		_, err := ChangeNodeLabelWithPatchModel(client, &test.node, test.labelToModify)
		if err != nil {
			t.Errorf("fail to add label to modify to slice")
		}
		updatedNode := &v1.Node{}
		key := types.NamespacedName{
			Name: "test-node",
		}
		err = client.Get(context.TODO(), key, updatedNode)
		if err != nil {
			t.Errorf("fail to update label to modify to slice")
		}
		if !reflect.DeepEqual((*updatedNode).Labels, test.wantedNode.Labels) {
			t.Errorf("fail to add label to modify to slice")
		}
	}
}

func TestAddLabelToModifyToSlice(t *testing.T) {
	var testCase = []struct {
		labelKey      string
		labelValue    string
		operationType v1alpha1.OperationType
		//labelsToModify *[]LabelToModify
		wantedLabelsToModify []LabelToModify
	}{
		{
			labelKey:      "commonLabel",
			labelValue:    "true",
			operationType: v1alpha1.AddLabel,
			wantedLabelsToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					LabelValue:    "true",
					OperationType: v1alpha1.AddLabel,
				},
			},
		},

		{
			labelKey:      "datasetNum",
			labelValue:    "12",
			operationType: v1alpha1.DeleteLabel,
			wantedLabelsToModify: []LabelToModify{
				{
					LabelKey:      "commonLabel",
					LabelValue:    "true",
					OperationType: v1alpha1.AddLabel,
				},
				{
					LabelKey:      "datasetNum",
					OperationType: v1alpha1.DeleteLabel,
				},
			},
		},
	}

	var labelsToModify []LabelToModify
	for _, test := range testCase {
		AddLabelToModifyToSlice(test.labelKey, test.labelValue, test.operationType, &labelsToModify)
		if !reflect.DeepEqual(labelsToModify, test.wantedLabelsToModify) {
			t.Errorf("fail to add labe to modify to slice")
		}
	}
}
