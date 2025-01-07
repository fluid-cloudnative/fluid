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
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
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

func TestGetFullNamespacedNameWithPrefixValue(t *testing.T) {
	tests := []struct {
		prefix, namespace, name, ownerDatasetUID string
		expected                                 string
	}{
		{"normal-", "default", "test-dataset", "", "normal-default-test-dataset"},
		{"overlimit-", "namespace-ajsdjikebnfacdsvwcaxqcackjascnbaksjcnakjscnackjasn", "dataset-demo", "58df5bd9cc", "overlimit-58df5bd9cc"},
		{"overlimit-", "namespace-demo", "dataset-ajsdjikebnfacdsvwcaxqcackjascnbaksjcnakjscnackjasn", "6dfd85695", "overlimit-6dfd85695"},
	}

	for _, test := range tests {
		result := GetNamespacedNameValueWithPrefix(test.prefix, test.namespace, test.name, test.ownerDatasetUID)
		if result != test.expected {
			t.Errorf("GetNamespacedNameValueWithPrefix(%v) = %v, want %v", test, result, test.expected)
		}
	}
}
