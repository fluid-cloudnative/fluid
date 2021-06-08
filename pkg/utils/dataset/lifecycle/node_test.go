package lifecycle

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

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
				Quota:  resource.NewQuantity(2,resource.BinarySI),
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
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceMemory: *resource.NewQuantity(3,resource.BinarySI),
					},
				},
			},
			want: true,
		},
		{
			runtimeInfo: runtimeInfoNotExclusive,
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: v1.NodeSpec{},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceMemory: *resource.NewQuantity(1,resource.BinarySI),
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

func TestDecreaseDatasetNum(t *testing.T) {
	var testCase = []struct {
		node           *v1.Node
		runtimeInfo    base.RuntimeInfo
		expectedResult bool
	}{
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "2"}},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo:    base.RuntimeInfo{},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "1"}},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo:    base.RuntimeInfo{},
			expectedResult: true,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"fluid.io/dataset-num": "test"}},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo:    base.RuntimeInfo{},
			expectedResult: false,
		},
		{
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.NodeSpec{},
			},
			runtimeInfo:    base.RuntimeInfo{},
			expectedResult: false,
		},
	}

	for _, test := range testCase {
		if result, _ := DecreaseDatasetNum(test.node, &test.runtimeInfo); result != test.expectedResult {
			t.Errorf("expected %v, got %v", test.expectedResult, result)
		}
	}
}
