package lifecycle

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestAlreadyAssigned(t *testing.T) {
	runtimeInfoExclusive, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoExclusive.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	var testCase = []struct {
		runtimeInfo base.RuntimeInfoInterface
		node        v1.Node
		want        bool
	}{
		{
			runtimeInfo: runtimeInfoExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.NodeSpec{},
			},
			want: false,
		},
		{
			runtimeInfo: runtimeInfoExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/s-fluid-hbase": "true"}},
				Spec:       v1.NodeSpec{},
			},
			want: true,
		},
		{
			runtimeInfo: runtimeInfoExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/s-fluid-spark": "true"}},
				Spec:       v1.NodeSpec{},
			},
			want: false,
		},
	}

	for _, test := range testCase {
		if result := AlreadyAssigned(test.runtimeInfo, test.node); result != test.want {
			t.Errorf("expected %v, got %v", test.want, result)
		}
	}
}

func TestCanbeAssigned(t *testing.T) {
	tireStore := datav1alpha1.Tieredstore{
		Levels: []datav1alpha1.Level{
			{
				MediumType: common.Memory,
				Quota:      resource.NewQuantity(2, resource.BinarySI),
			},
		},
	}
	runtimeInfoNotExclusive, err := base.BuildRuntimeInfo("hbase", "default", "alluxio", tireStore)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoNotExclusive.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})

	var testCase = []struct {
		runtimeInfo base.RuntimeInfoInterface
		node        v1.Node
		want        bool
	}{
		{
			runtimeInfo: runtimeInfoNotExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"fluid_exclusive": "default_hbase"},
				},
				Status: v1.NodeStatus{},
			},
			want: false,
		},
		{
			runtimeInfo: runtimeInfoNotExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.NodeSpec{},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceMemory: *resource.NewQuantity(3, resource.BinarySI),
					},
				},
			},
			want: true,
		},
		{
			runtimeInfo: runtimeInfoNotExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.NodeSpec{},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceMemory: *resource.NewQuantity(1, resource.BinarySI),
					},
				},
			},
			want: false,
		},
	}

	for _, test := range testCase {
		if result := CanbeAssigned(test.runtimeInfo, test.node); result != test.want {
			t.Errorf("expected %v, got %v", test.want, result)
		}
	}
}

func TestLabelCacheNode(t *testing.T) {
	runtimeInfoExclusive, err := base.BuildRuntimeInfo("hbase", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoExclusive.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	runtimeInfoShare, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoShare.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ShareMode},
	})

	tireStore := datav1alpha1.Tieredstore{
		Levels: []datav1alpha1.Level{
			{
				MediumType: common.Memory,
				Quota:      resource.NewQuantity(1, resource.BinarySI),
			},
			{
				MediumType: common.SSD,
				Quota:      resource.NewQuantity(2, resource.BinarySI),
			},
			{
				MediumType: common.HDD,
				Quota:      resource.NewQuantity(3, resource.BinarySI),
			},
		},
	}
	runtimeInfoWithTireStore, err := base.BuildRuntimeInfo("spark", "fluid", "alluxio", tireStore)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	runtimeInfoWithTireStore.SetupWithDataset(&datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{PlacementMode: datav1alpha1.ExclusiveMode},
	})

	nodeInputs := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-exclusive",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-tireStore",
			},
		},
	}

	testNodes := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		testNodes = append(testNodes, nodeInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	var testCase = []struct {
		node        v1.Node
		runtimeInfo base.RuntimeInfoInterface
		wantedMap   map[string]string
	}{
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-exclusive",
				},
			},
			runtimeInfo: runtimeInfoExclusive,
			wantedMap: map[string]string{
				"fluid.io/dataset-num":               "1",
				"fluid.io/s-alluxio-fluid-hbase":     "true",
				"fluid.io/s-fluid-hbase":             "true",
				"fluid.io/s-h-alluxio-t-fluid-hbase": "0B",
				"fluid_exclusive":                    "fluid_hbase",
			},
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-exclusive",
				},
			},
			runtimeInfo: runtimeInfoExclusive,
			wantedMap: map[string]string{
				"fluid.io/dataset-num":               "2",
				"fluid.io/s-alluxio-fluid-hbase":     "true",
				"fluid.io/s-fluid-hbase":             "true",
				"fluid.io/s-h-alluxio-t-fluid-hbase": "0B",
				"fluid_exclusive":                    "fluid_hbase",
			},
		},
		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-share",
				},
			},
			runtimeInfo: runtimeInfoShare,
			wantedMap: map[string]string{
				"fluid.io/dataset-num":               "1",
				"fluid.io/s-alluxio-fluid-spark":     "true",
				"fluid.io/s-fluid-spark":             "true",
				"fluid.io/s-h-alluxio-t-fluid-spark": "0B",
			},
		},

		{
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-tireStore",
				},
			},
			runtimeInfo: runtimeInfoWithTireStore,
			wantedMap: map[string]string{
				"fluid.io/dataset-num":               "1",
				"fluid.io/s-alluxio-fluid-spark":     "true",
				"fluid.io/s-fluid-spark":             "true",
				"fluid.io/s-h-alluxio-d-fluid-spark": "5B",
				"fluid.io/s-h-alluxio-m-fluid-spark": "1B",
				"fluid.io/s-h-alluxio-t-fluid-spark": "6B",
				"fluid_exclusive":                    "fluid_spark",
			},
		},
	}

	for _, test := range testCase {
		err := LabelCacheNode(test.node, test.runtimeInfo, client)
		if err != nil {
			t.Errorf("fail to exec the function with the error %v", err)
		}
		node, err := kubeclient.GetNode(client, test.node.Name)
		if err != nil {
			fmt.Println(err)
		}
		if !reflect.DeepEqual(node.Labels, test.wantedMap) {
			t.Errorf("fail to update the labels")
		}
	}
}

func TestDecreaseDatasetNum(t *testing.T) {
	var testCase = []struct {
		node           *v1.Node
		runtimeInfo    base.RuntimeInfo
		expectedResult []common.LabelToModify
	}{
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "2"}},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo: base.RuntimeInfo{},
			expectedResult: []common.LabelToModify{
				{
					LabelKey:      "fluid.io/dataset-num",
					LabelValue:    "1",
					OperationType: common.UpdateLabel,
				},
			},
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "1"}},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo: base.RuntimeInfo{},
			expectedResult: []common.LabelToModify{
				{
					LabelKey:      "fluid.io/dataset-num",
					LabelValue:    "",
					OperationType: common.DeleteLabel,
				},
			},
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "test"}},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo:    base.RuntimeInfo{},
			expectedResult: nil,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo:    base.RuntimeInfo{},
			expectedResult: nil,
		},
	}

	for _, test := range testCase {
		var labels common.LabelsToModify
		_ = DecreaseDatasetNum(test.node, &test.runtimeInfo, &labels)
		if !reflect.DeepEqual(labels.Labels, test.expectedResult) {
			t.Errorf("fail to exec the function with the error ")
		}

	}
}
