package alluxio

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func getTestAlluxioEngineNode(client client.Client, name string, namespace string, withRunTime bool) *AlluxioEngine {
	engine := &AlluxioEngine{
		runtime:                nil,
		name:                   name,
		namespace:              namespace,
		Client:                 client,
		runtimeInfo:            nil,
		Log:                    log.NullLogger{},
	}
	if withRunTime{
		engine.runtime = &v1alpha1.AlluxioRuntime{}
		engine.runtimeInfo,_ = base.BuildRuntimeInfo(name,namespace,"alluxio",v1alpha1.TieredStore{})
	}
	return engine
}

func TestAssignNodesToCache(t *testing.T) {
	dataSet := &v1alpha1.Dataset{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "hbase",
			Namespace: "fluid",
		},
		Spec:       v1alpha1.DatasetSpec{},
		Status:     v1alpha1.DatasetStatus{},
	}
	nodeInputs := []*v1.Node{
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
	runtimeObjs  := []runtime.Object{}
	runtimeObjs  = append(runtimeObjs, dataSet)
	for _, nodeInput := range nodeInputs {
		runtimeObjs = append(runtimeObjs, nodeInput.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme,runtimeObjs...)


	testCases := []struct{
		withRunTime bool
		name 		string
		namespace 	string
		out 		int32
		isErr 		bool
	}{
		{
			withRunTime: 	true,
			name: 			"hbase",
			namespace: 		"fluid",
			out: 			2,
			isErr:          false,
		},
		{
			withRunTime: 	false,
			name: 			"hbase",
			namespace: 		"fluid",
			out: 			0,
			isErr:          true,
		},
		{
			withRunTime: 	true,
			name: 			"not-found",
			namespace: 		"fluid",
			out: 			0,
			isErr:          true,
		},

	}
	for _,testCase := range testCases{
		engine := getTestAlluxioEngineNode(fakeClient,testCase.name,testCase.namespace,testCase.withRunTime)
		out,err := engine.AssignNodesToCache(3)  // num: 2 err: nil
		if out!=testCase.out{
			t.Errorf("expected %d, got %d.", testCase.out, out)
		}
		isErr := err != nil
		if isErr != testCase.isErr{
			t.Errorf("expected %t, got %t.", testCase.isErr, isErr)
		}
	}
}
