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
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestAssignDatasetToNodesWithIllegalLabel(t *testing.T) {
	// runtimeInfoSpark tests destroy Worker in exclusive mode.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	datasetoHadoop := &datav1alpha1.Dataset{Spec: datav1alpha1.DatasetSpec{
		NodeAffinity: &datav1alpha1.CacheableNodeAffinity{
			Required: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "nodeA",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"tr/ue"},
							},
						},
					},
				},
			},
		},
	},
	}

	// runtimeInfoSpark tests destroy Worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-alluxio-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":             "true",
					"fluid.io/s-h-alluxio-d-fluid-spark": "5B",
					"fluid.io/s-h-alluxio-m-fluid-spark": "1B",
					"fluid.io/s-h-alluxio-t-fluid-spark": "6B",
					"fluid_exclusive":                    "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"fluid.io/s-alluxio-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase":  "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase":  "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
	}

	objects := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		objects = append(objects, nodeInput.DeepCopy())
	}

	objects = append(objects, datasetoHadoop)
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)

	client := fake.NewFakeClientWithScheme(testScheme, objects...)

	var testCase = []struct {
		expectedWorkers  int32
		runtimeInfo      base.RuntimeInfoInterface
		wantedNodeNumber int32
		wantedNodeLabels map[string]map[string]string
	}{
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoSpark,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":                "2",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"fluid.io/s-alluxio-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase":  "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase":  "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-alluxio-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hadoop": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hadoop": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hadoop": "6B",
					"node-select":                         "true",
				},
			},
		},
		{
			expectedWorkers:  -1,
			runtimeInfo:      runtimeInfoHadoop,
			wantedNodeNumber: 0,
			wantedNodeLabels: map[string]map[string]string{
				"test-node-spark": {},
				"test-node-share": {
					"fluid.io/dataset-num":               "1",
					"fluid.io/s-alluxio-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":             "true",
					"fluid.io/s-h-alluxio-d-fluid-hbase": "5B",
					"fluid.io/s-h-alluxio-m-fluid-hbase": "1B",
					"fluid.io/s-h-alluxio-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		},
	}
	for _, test := range testCase {
		_, err := AssignDatasetToNodes(test.runtimeInfo, datasetoHadoop, client, 1)
		if err != nil {
			t.Errorf("Failed due to %v", err)
		}
	}
}
