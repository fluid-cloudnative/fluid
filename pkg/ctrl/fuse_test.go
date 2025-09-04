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

package ctrl

import (
	"context"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testScheme *runtime.Scheme
)

func TestCleanUpFuse(t *testing.T) {
	var testCase = []struct {
		name             string
		namespace        string
		wantedNodeLabels map[string]map[string]string
		wantedCount      int
		context          cruntime.ReconcileRequestContext
		log              logr.Logger
		runtimeType      string
		nodeInputs       []*corev1.Node
	}{
		{
			wantedCount: 1,
			name:        "hbase",
			namespace:   "fluid",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"multiple-fuse": {
					"fluid.io/f-fluid-hadoop":          "true",
					"node-select":                      "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
				},
				"fuse": {
					"fluid.io/dataset-num":    "1",
					"fluid.io/f-fluid-hadoop": "true",
					"node-select":             "true",
				},
			},
			log:         fake.NullLogger(),
			runtimeType: "jindo",
			nodeInputs: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "no-fuse",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "multiple-fuse",
						Labels: map[string]string{
							"fluid.io/f-fluid-hadoop":          "true",
							"node-select":                      "true",
							"fluid.io/f-fluid-hbase":           "true",
							"fluid.io/s-fluid-hbase":           "true",
							"fluid.io/s-h-jindo-d-fluid-hbase": "5B",
							"fluid.io/s-h-jindo-m-fluid-hbase": "1B",
							"fluid.io/s-h-jindo-t-fluid-hbase": "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fuse",
						Labels: map[string]string{
							"fluid.io/dataset-num":    "1",
							"fluid.io/f-fluid-hadoop": "true",
							"node-select":             "true",
						},
					},
				},
			},
		},
		{
			wantedCount: 2,
			name:        "spark",
			namespace:   "fluid",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"multiple-fuse": {
					"node-select":                        "true",
					"fluid.io/s-fluid-hbase":             "true",
					"fluid.io/f-fluid-hbase":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
				},
				"fuse": {
					"fluid.io/dataset-num": "1",
					"node-select":          "true",
				},
			},
			log:         fake.NullLogger(),
			runtimeType: "alluxio",
			nodeInputs: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "no-fuse",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "multiple-fuse",
						Labels: map[string]string{
							"fluid.io/f-fluid-spark":             "true",
							"node-select":                        "true",
							"fluid.io/f-fluid-hbase":             "true",
							"fluid.io/s-fluid-hbase":             "true",
							"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
							"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
							"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fuse",
						Labels: map[string]string{
							"fluid.io/dataset-num":   "1",
							"fluid.io/f-fluid-spark": "true",
							"node-select":            "true",
						},
					},
				},
			},
		},
		{
			wantedCount: 0,
			name:        "hbase",
			namespace:   "fluid",
			wantedNodeLabels: map[string]map[string]string{
				"no-fuse": {},
				"multiple-fuse": {
					"fluid.io/f-fluid-spark":              "true",
					"node-select":                         "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/f-fluid-hadoop":             "true",
					"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
					"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
					"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
				},
				"fuse": {
					"fluid.io/dataset-num":   "1",
					"fluid.io/f-fluid-spark": "true",
					"node-select":            "true",
				},
			},
			log:         fake.NullLogger(),
			runtimeType: "goosefs",
			nodeInputs: []*corev1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "no-fuse",
						Labels: map[string]string{},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "multiple-fuse",
						Labels: map[string]string{
							"fluid.io/f-fluid-spark":              "true",
							"node-select":                         "true",
							"fluid.io/f-fluid-hadoop":             "true",
							"fluid.io/s-fluid-hadoop":             "true",
							"fluid.io/s-h-goosefs-d-fluid-hadoop": "5B",
							"fluid.io/s-h-goosefs-m-fluid-hadoop": "1B",
							"fluid.io/s-h-goosefs-t-fluid-hadoop": "6B",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fuse",
						Labels: map[string]string{
							"fluid.io/dataset-num":   "1",
							"fluid.io/f-fluid-spark": "true",
							"node-select":            "true",
						},
					},
				},
			},
		},
	}
	for _, test := range testCase {

		testNodes := []runtime.Object{}
		for _, nodeInput := range test.nodeInputs {
			testNodes = append(testNodes, nodeInput.DeepCopy())
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, testNodes...)

		nodeList := &corev1.NodeList{}
		runtimeInfo, err := base.BuildRuntimeInfo(
			test.name,
			test.namespace,
			test.runtimeType,
		)
		if err != nil {
			t.Errorf("build runtime info error %v", err)
		}
		h := &Helper{
			runtimeInfo: runtimeInfo,
			client:      fakeClient,
			log:         test.log,
		}

		count, err := h.CleanUpFuse()
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		if count != test.wantedCount {
			t.Errorf("with the wrong number of the fuse ,count %v", count)
		}

		err = fakeClient.List(context.TODO(), nodeList, &client.ListOptions{})
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
