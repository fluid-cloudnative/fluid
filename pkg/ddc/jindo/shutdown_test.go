package jindo

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

func TestDestroyWorker(t *testing.T) {
	runtimeInfoSpark, err := base.BuildRuntimeInfo("spark", "fluid", "jindo", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoSpark.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	runtimeInfoHadoop, err := base.BuildRuntimeInfo("hadoop", "fluid", "jindo", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoHadoop.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})
	nodeSelector := map[string]string{
		"node-select": "true",
	}
	runtimeInfoHadoop.SetupFuseDeployMode(true, nodeSelector)

	var nodeInputs = []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{ // 里面只有fluid的spark
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":             "1",
					"fluid.io/s-jindo-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":           "true",
					"fluid.io/s-h-jindo-d-fluid-spark": "5B",
					"fluid.io/s-h-jindo-m-fluid-spark": "1B",
					"fluid.io/s-h-jindo-t-fluid-spark": "6B",
					"fluid_exclusive":                  "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":              "2",
					"fluid.io/s-jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
					"fluid.io/s-jindo-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":            "true",
					"fluid.io/s-h-jindo-d-fluid-hbase":  "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase":  "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":              "1",
					"fluid.io/s-jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
					"node-select":                       "true",
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
					"fluid.io/dataset-num":              "2",
					"fluid.io/s-jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
					"fluid.io/s-jindo-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":            "true",
					"fluid.io/s-h-jindo-d-fluid-hbase":  "5B",
					"fluid.io/s-h-jindo-m-fluid-hbase":  "1B",
					"fluid.io/s-h-jindo-t-fluid-hbase":  "6B",
				},
				"test-node-hadoop": {
					"fluid.io/dataset-num":              "1",
					"fluid.io/s-jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-jindo-t-fluid-hadoop": "6B",
					"node-select":                       "true",
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
		engine := &JindoEngine{Log: log.NullLogger{}, runtimeInfo: test.runtimeInfo}
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

func TestCleanConfigmap(t *testing.T) {

	namespace := "default"
	runtimeType := "jindo"

	configMapInputs := []*v1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "hbase-alluxio-values", Namespace: namespace},
			Data: map[string]string{
				"data": "image: fluid\nimageTag: 0.6.0",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "hbase-alluxio-config", Namespace: namespace},
			Data:       map[string]string{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "spark-alluxio-values", Namespace: namespace},
			Data: map[string]string{
				"test-data": "image: fluid\n imageTag: 0.6.0",
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{Name: "hadoop-alluxio-config", Namespace: namespace},
		},
	}

	testConfigMaps := []runtime.Object{}
	for _, cm := range configMapInputs {
		testConfigMaps = append(testConfigMaps, cm.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "ConfigMap doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
		},
		{
			name: "ConfigMap value exists",
			args: args{
				name:      "test1",
				namespace: namespace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &JindoEngine{
				Log:         log.NullLogger{},
				name:        tt.args.name,
				namespace:   tt.args.namespace,
				runtimeType: runtimeType,
				Client:      client}
			err := engine.cleanConfigMap()
			if err != nil {
				t.Errorf("fail to clean configmap due to %v", err)
			}
		})
	}
}
