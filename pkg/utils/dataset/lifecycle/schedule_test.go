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

package lifecycle

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

func TestAssignDatasetToNodes(t *testing.T) {

	var nodes []corev1.Node
	pvcMountNodesMap := map[string]int64{}

	fuseSelectLabel := map[string]string{"fuse": "true"}
	fuseNotSelectLabel := map[string]string{"fuse": "false"}

	for i := 1; i <= 100; i++ {
		node := corev1.Node{}
		nodeName := "node" + strconv.Itoa(i)
		node.Name = nodeName
		pvcMountPodsNum := rand.Int63n(5)
		if pvcMountPodsNum != 0 {
			pvcMountNodesMap[nodeName] = pvcMountPodsNum
			node.Labels = fuseSelectLabel
		} else {
			fuseSelect := rand.Intn(2)
			if fuseSelect == 1 {
				node.Labels = fuseSelectLabel
			} else {
				node.Labels = fuseNotSelectLabel
			}
		}
		nodes = append(nodes, node)
	}
	nodes = sortNodesToBeScheduled(nodes, pvcMountNodesMap, fuseSelectLabel)

	for i := 0; i < len(nodes)-1; i++ {
		if nodes[i].Labels["fuse"] == "false" && nodes[i+1].Labels["fuse"] == "true" {
			t.Errorf("the result of sort is not right")
		}

		numFront, found := pvcMountNodesMap[nodes[i].Name]
		if !found {
			numFront = 0
		}
		numBehind, found := pvcMountNodesMap[nodes[i+1].Name]
		if !found {
			numBehind = 0
		}
		if numFront < numBehind {
			t.Errorf("the result of sort is not right")
		}

	}
}

func TestSortNodesToBeScheduled(t *testing.T) {
	var nodes = []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
			Status: corev1.NodeStatus{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node2",
			},
			Status: corev1.NodeStatus{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node3",
			},
			Status: corev1.NodeStatus{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node4",
				Labels: map[string]string{"test_label_key": "test_label_value"},
			},
			Status: corev1.NodeStatus{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node5",
			},
			Status: corev1.NodeStatus{},
		},
	}

	var pvcMountNodesMap = map[string]int64{
		"node3": 3,
		"node2": 2,
		"node1": 1,
	}

	var nodeSelector = map[string]string{
		"test_label_key": "test_label_value",
	}

	var tests = []struct {
		testNodes            []corev1.Node
		testPvcMountNodesMap map[string]int64
		testNodeSelector     map[string]string
		testExpectedResult   []corev1.Node
	}{
		{
			testNodes:            nodes,
			testPvcMountNodesMap: pvcMountNodesMap,
			testNodeSelector:     nodeSelector,
			testExpectedResult: []corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
					},
					Status: corev1.NodeStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: corev1.NodeStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
					Status: corev1.NodeStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node4",
						Labels: map[string]string{"test_label_key": "test_label_value"},
					},
					Status: corev1.NodeStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node5",
					},
					Status: corev1.NodeStatus{},
				},
			},
		},
	}

	for _, test := range tests {
		if result := sortNodesToBeScheduled(test.testNodes, test.testPvcMountNodesMap, test.testNodeSelector); !reflect.DeepEqual(result, test.testExpectedResult) {
			t.Errorf("expeced %v, get %v", test.testExpectedResult, result)
		}
	}
}
