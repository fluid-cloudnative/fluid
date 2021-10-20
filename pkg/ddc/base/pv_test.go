package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"testing"
)

func TestGetPersistentVolumeName(t *testing.T) {
	var testCases = []struct {
		runtimeName        string
		runtimeNamespace   string
		isDeprecatedPVName bool
		expectedPVName     string
	}{
		{
			runtimeName:        "spark",
			runtimeNamespace:   "fluid",
			isDeprecatedPVName: false,
			expectedPVName:     "fluid-spark",
		},
		{
			runtimeName:        "hadoop",
			runtimeNamespace:   "test",
			isDeprecatedPVName: false,
			expectedPVName:     "test-hadoop",
		},
		{
			runtimeName:        "hbase",
			runtimeNamespace:   "fluid",
			isDeprecatedPVName: true,
			expectedPVName:     "hbase",
		},
	}
	for _, testCase := range testCases {
		runtimeInfo, err := BuildRuntimeInfo(testCase.runtimeName, testCase.runtimeNamespace, "alluxio", datav1alpha1.TieredStore{})
		if err != nil {
			t.Errorf("fail to create the runtimeInfo with error %v", err)
		}
		runtimeInfo.SetDeprecatedPVName(testCase.isDeprecatedPVName)
		result := runtimeInfo.GetPersistentVolumeName()
		if result != testCase.expectedPVName {
			t.Errorf("get failure, expected %s, get %s", testCase.expectedPVName, result)
		}
	}
}
