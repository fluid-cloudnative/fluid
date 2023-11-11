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
