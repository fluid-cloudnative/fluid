package utils

import (
	"context"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
		labelToModify common.LabelsToModify
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
			labelToModify: common.LabelsToModify{},
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
		test.labelToModify.Add("commonLabel", "true")
		test.labelToModify.Update("datasetNum", "1")
		test.labelToModify.Delete("deleteLabel")
		_, err := ChangeNodeLabelWithUpdateMode(client, &test.node, test.labelToModify)
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
		labelToModify common.LabelsToModify
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
			labelToModify: common.LabelsToModify{},

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
		test.labelToModify.Add("commonLabel", "true")
		test.labelToModify.Update("datasetNum", "1")
		test.labelToModify.Delete("deleteLabel")
		_, err := ChangeNodeLabelWithPatchMode(client, &test.node, test.labelToModify)
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
