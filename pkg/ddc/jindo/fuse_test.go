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
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
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
			wantedCount: 1,
			name:        "fluid-hadoop-1",
			namespace:   "jindo",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"fuse": {
					"fluid.io/f-jindo-fluid-hadoop":    "true",
					"node-select":                      "true",
					"fluid.io/f-jindo-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"multiple-fuse": {
					"fluid.io/dataset-num":          "1",
					"fluid.io/f-jindo-fluid-hadoop": "true",
					"node-select":                   "true",
				},
			},
		}, {
			wantedCount: 2,
			name:        "fluid-hadoop",
			namespace:   "jindo",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"fuse": {
					"node-select":                      "true",
					"fluid.io/f-jindo-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"multiple-fuse": {
					"fluid.io/dataset-num": "1",
					"node-select":          "true",
				},
			},
		},
		{
			wantedCount: 0,
			name:        "test",
			namespace:   "default",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"fuse": {
					"node-select":                      "true",
					"fluid.io/f-jindo-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"multiple-fuse": {
					"fluid.io/dataset-num": "1",
					"node-select":          "true",
				},
			},
		},
	}
	for _, test := range testCase {
		nodeList := &v1.NodeList{}
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

		err = client.List(context.TODO(), nodeList, &cli.ListOptions{})
		if err != nil {
			t.Errorf("testcase %s: fail to get the node with the error %v  ", test.name, err)
		}

		for _, node := range nodeList.Items {
			if len(node.Labels) != len(test.wantedNodeLabels[node.Name]) {
				t.Errorf("testcase %s: fail to clean up the labels for node %s  expected %v, got %v", test.name, node.Name, test.wantedNodeLabels[node.Name], node.Labels)
			}
			if len(node.Labels) != 0 && !reflect.DeepEqual(node.Labels, test.wantedNodeLabels[node.Name]) {
				t.Errorf("testcase %s: fail to clean up the labels for node  %s  expected %v, got %v", test.name, node.Name, test.wantedNodeLabels[node.Name], node.Labels)
			}
		}

	}
}

func TestCleanAll(t *testing.T) {
	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
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

	engine := &JindoEngine{Log: log.NullLogger{}}
	engine.Client = client
	engine.name = "fluid-hadoop"
	engine.namespace = "default"
	err := engine.cleanAll()
	if err != nil {
		t.Errorf("failed to cleanAll due to %v", err)
	}

}
