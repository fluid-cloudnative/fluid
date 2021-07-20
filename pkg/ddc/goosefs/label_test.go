package goosefs

import "testing"

func TestGetCommonLabelname(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			out:       "fluid.io/s-fluid-hbase",
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			out:       "fluid.io/s-fluid-hadoop",
		},
		{
			name:      "common",
			namespace: "default",
			out:       "fluid.io/s-default-common",
		},
	}
	for _, testCase := range testCases {
		engine := &GooseFSEngine{
			name:      testCase.name,
			namespace: testCase.namespace,
		}
		out := engine.getCommonLabelname()
		if out != testCase.out {
			t.Errorf("in: %s-%s, expect: %s, got: %s", testCase.namespace, testCase.name, testCase.out, out)
		}
	}
}
