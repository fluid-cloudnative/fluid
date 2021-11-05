/*
Copyright 2021 The Fluid Authors.

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

package jindo

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestCleanupFuse(t *testing.T) {
	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{ // 里面只有fluid的spark
				Name:   "no-fuse",
				Labels: map[string]string{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fuse",
				Labels: map[string]string{
					"fluid.io/f-jindo-fluid-hadoop":    "true",
					"node-select":                      "true",
					"fluid.io/f-jindo-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "multiple-fuse",
				Labels: map[string]string{
					"fluid.io/dataset-num":            "1",
					"fluid.io/f-jindo-fluid-hadoop":   "true",
					"fluid.io/f-jindo-fluid-hadoop-1": "true",
					"node-select":                     "true",
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
		name             string
		namespace        string
		wantedNodeLabels map[string]map[string]string
		wantedCount      int
	}{
		{
			wantedCount: 2,
			name:        "fluid-hadoop",
			namespace:   "jindo",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"fuse": {
					"node-select": "true",
				},
				"multiple-fuse": {
					"fluid.io/dataset-num":            "1",
					"fluid.io/f-jindo-fluid-hadoop-1": "true",
					"node-select":                     "true",
				},
			},
		},
		{
			wantedCount: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":             "1",
					"fluid.io/s-jindo-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		},
	}
	for _, test := range testCase {
		engine := &JindoEngine{Log: log.NullLogger{}}
		engine.Client = client
		engine.name = test.name
		engine.namespace = test.namespace
		count, err := engine.cleanupFuse()
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		if count != test.wantedCount {
			t.Errorf("with the wrong number of the fuse")
		}
		for _, node := range nodeInputs {
			newNode, err := kubeclient.GetNode(client, node.Name)
			if err != nil {
				t.Errorf("fail to get the node with the error %v", err)
			}

			if len(newNode.Labels) != len(test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to clean up the labels")
			}
			if len(newNode.Labels) != 0 && !reflect.DeepEqual(newNode.Labels, test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to clean up the labels")
			}
		}

	}
}
