package lifecycle

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestUpdateOrDeleteDatasetNum(t *testing.T) {
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
		if result, _ := UpdateOrDeleteDatasetNum(test.node, &test.runtimeInfo); result != test.expectedResult {
			t.Errorf("expected %v, got %v", test.expectedResult, result)
		}
	}
}
