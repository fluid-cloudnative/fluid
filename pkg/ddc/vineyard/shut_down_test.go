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

package vineyard

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testScheme        *runtime.Scheme
	mockConfigMapData = `master:
	ports:
	  rpc: 30399
	  web: 31203
  jobMaster:
	ports:
	  rpc: 28362
	  web: 31380
  worker:
	ports:
	  rpc: 31285
	  web: 31674
  jobWorker:
	ports:
	  rpc: 29476
	  web: 27403
	  data: 30918`
)

func init() {
	testScheme = runtime.NewScheme()
	_ = corev1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

func TestDestroyWorker(t *testing.T) {
	// runtimeInfoSpark tests destroy Worker in exclusive mode.
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "vineyard", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	// runtimeInfoSpark tests destroy Worker in shareMode mode.
	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "vineyard", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	var nodeInputs = []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-vineyard-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":              "true",
					"fluid.io/s-h-vineyard-d-fluid-spark": "5B",
					"fluid.io/s-h-vineyard-m-fluid-spark": "1B",
					"fluid.io/s-h-vineyard-t-fluid-spark": "6B",
					"fluid_exclusive":                     "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":                 "2",
					"fluid.io/s-vineyard-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
					"fluid.io/s-vineyard-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":               "true",
					"fluid.io/s-h-vineyard-d-fluid-hbase":  "5B",
					"fluid.io/s-h-vineyard-m-fluid-hbase":  "1B",
					"fluid.io/s-h-vineyard-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":                 "1",
					"fluid.io/s-vineyard-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
					"node-select":                          "true",
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
					"fluid.io/dataset-num":                 "2",
					"fluid.io/s-vineyard-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
					"fluid.io/s-vineyard-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":               "true",
					"fluid.io/s-h-vineyard-d-fluid-hbase":  "5B",
					"fluid.io/s-h-vineyard-m-fluid-hbase":  "1B",
					"fluid.io/s-h-vineyard-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":                 "1",
					"fluid.io/s-vineyard-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hadoop": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hadoop": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hadoop": "6B",
					"node-select":                          "true",
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
					"fluid.io/dataset-num":                "1",
					"fluid.io/s-vineyard-fluid-hbase":     "true",
					"fluid.io/s-fluid-hbase":              "true",
					"fluid.io/s-h-vineyard-d-fluid-hbase": "5B",
					"fluid.io/s-h-vineyard-m-fluid-hbase": "1B",
					"fluid.io/s-h-vineyard-t-fluid-hbase": "6B",
				},
				"test-node-hadoop": {
					"node-select": "true",
				},
			},
		},
	}
	for _, test := range testCase {
		engine := &VineyardEngine{Log: fake.NullLogger(), runtimeInfo: test.runtimeInfo}
		engine.Client = client
		engine.name = test.runtimeInfo.GetName()
		engine.namespace = test.runtimeInfo.GetNamespace()
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		currentWorkers, err := engine.destroyWorkers(test.expectedWorkers)
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		if currentWorkers != test.wantedNodeNumber {
			t.Errorf("shutdown the worker with the wrong number of the workers")
		}
		for _, node := range nodeInputs {
			newNode, err := kubeclient.GetNode(client, node.Name)
			if err != nil {
				t.Errorf("fail to get the node with the error %v", err)
			}

			if len(newNode.Labels) != len(test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to decrease the labels")
			}
			if len(newNode.Labels) != 0 && !reflect.DeepEqual(newNode.Labels, test.wantedNodeLabels[node.Name]) {
				t.Errorf("fail to decrease the labels")
			}
		}

	}
}

func TestVineyardEngineCleanAll(t *testing.T) {
	type fields struct {
		name        string
		namespace   string
		cm          *corev1.ConfigMap
		runtimeType string
		log         logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:        "spark",
				namespace:   "fluid",
				runtimeType: "vineyard",
				cm: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-vineyard-values",
						Namespace: "fluid",
					},
					Data: map[string]string{"data": mockConfigMapData},
				},
				log: fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testObjs := []runtime.Object{}
			testObjs = append(testObjs, tt.fields.cm.DeepCopy())
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

			helper := &ctrl.Helper{}
			patch1 := ApplyMethod(reflect.TypeOf(helper), "CleanUpFuse", func(_ *ctrl.Helper) (int, error) {
				return 0, nil
			})
			defer patch1.Reset()
			e := &VineyardEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       tt.fields.log,
			}
			if err := e.cleanAll(); (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.cleanAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVineyardEngineCleanupCache(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:      "spark",
				namespace: "field",
				Log:       fake.NullLogger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &VineyardEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}

			patch := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
				func(_ *VineyardEngine) ([]string, error) {
					return []string{"instances_memory_usage_bytes 1234", "instances_memory_usage_bytes 4321"}, nil
				})
			defer patch.Reset()

			if err := engine.cleanupCache(); (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.cleanupCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVineyardEngineDestroyMaster(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "spark",
			fields: fields{
				name:      "spark",
				namespace: "fluid",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &VineyardEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}

			patch1 := ApplyFunc(helm.CheckRelease,
				func(_ string, _ string) (bool, error) {
					d := true
					return d, nil
				})
			defer patch1.Reset()

			patch2 := ApplyFunc(helm.DeleteRelease,
				func(_ string, _ string) error {
					return nil
				})
			defer patch2.Reset()

			if err := e.destroyMaster(); (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.destroyMaster() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
