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

package juicefs

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestGetCacheInfoFromConfigmap(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSet := &v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "fluid",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "test-dataset",
					Namespace: "fluid",
					Type:      "juicefs",
				},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantCacheInfo := map[string]string{"mountpath": "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse", "edition": "community"}
	cacheinfo, err := GetCacheInfoFromConfigmap(fakeClient, dataSet.Name, dataSet.Namespace)
	if err != nil {
		t.Errorf("GetCacheInfoFromConfigmap failed.")
	}
	if !reflect.DeepEqual(cacheinfo, wantCacheInfo) {
		t.Errorf("gotcacheinfo = %v, want %v", cacheinfo, wantCacheInfo)
	}

}

// Test_parseCacheInfoFromConfigMap is a unit test function for the parseCacheInfoFromConfigMap function.
// This test function verifies whether parseCacheInfoFromConfigMap correctly parses data from a ConfigMap
// and returns the expected cache information. It also checks if the function handles errors as expected
// under different input conditions.
//
// Test Cases:
// 1. Normal Case: Input a ConfigMap with valid data, and verify if the function correctly parses and returns the cache information.
// 2. Error Case: Input a ConfigMap with invalid data, and verify if the function handles the error and returns the expected result.
//
// Input:
// - No explicit input parameters; test cases are defined via a struct.
//
// Output:
// - No explicit return value; test results are reported using methods from testing.T.
func Test_parseCacheInfoFromConfigMap(t *testing.T) {
    // args struct defines the input parameters for the test cases.
    // configMap is the input parameter for parseCacheInfoFromConfigMap, of type *v1.ConfigMap.
    type args struct {
        configMap *v1.ConfigMap
    }

    // Define a set of test cases. Each test case contains the following fields:
    // - name: The name of the test case, used to identify the test scenario.
    // - args: The input parameters for the test case, of type args.
    // - wantCacheInfo: The expected cache information result, of type map[string]string.
    // - wantErr: Whether an error is expected.
    tests := []struct {
        name          string            // Name of the test case
        args          args              // Input parameters for the test case
        wantCacheInfo map[string]string // Expected cache information result
        wantErr       bool              // Whether an error is expected
    }{
        // First test case: Normal parsing of ConfigMap data
        {
            name: "parseCacheInfoFromConfigMap", // Name of the test case
            args: args{configMap: &v1.ConfigMap{ // Input parameters
                Data: map[string]string{ // Data field of the ConfigMap
                    "data": valuesConfigMapData, // Assume valuesConfigMapData is a predefined string containing valid configuration data
                },
            }},
            wantCacheInfo: map[string]string{ // Expected cache information result
                "mountpath": "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse", // Expected mount path
                "edition":   "community", // Expected edition information
            },
            wantErr: false, // No error is expected
        },
        // Second test case: Parsing invalid ConfigMap data
        {
            name: "parseCacheInfoFromConfigMap-err", // Name of the test case
            args: args{configMap: &v1.ConfigMap{ // Input parameters
                Data: map[string]string{ // Data field of the ConfigMap
                    "data": `test`, // Invalid configuration data
                },
            }},
            wantCacheInfo: nil, // Expected cache information result is nil
            wantErr:       true, // An error is expected
        },
    }

    // Iterate through the test cases and run each test
    for _, tt := range tests {
        // Use t.Run to execute the subtest, with the subtest name as tt.name
        t.Run(tt.name, func(t *testing.T) {
            // Call the function under test, parseCacheInfoFromConfigMap, and get the return value and error
            gotPorts, err := parseCacheInfoFromConfigMap(tt.args.configMap)

            // Check if the error matches the expectation
            if (err != nil) != tt.wantErr {
                t.Errorf("parseCacheInfoFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // Check if the returned cache information matches the expectation
            if !reflect.DeepEqual(gotPorts, tt.wantCacheInfo) {
                t.Errorf("parseCacheInfoFromConfigMap() gotPorts = %v, want %v", gotPorts, tt.wantCacheInfo)
            }
        })
    }
}

func TestGetFSInfoFromConfigMap(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset-juicefs-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}
	dataSet := &v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "fluid",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "test-dataset",
					Namespace: "fluid",
					Type:      "juicefs",
				},
			},
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, configMap)
	runtimeObjs = append(runtimeObjs, dataSet.DeepCopy())
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
	wantMetaurlInfo := map[string]string{
		MetaurlSecret:      "jfs-secret",
		MetaurlSecretKey:   "metaurl",
		SecretKeySecret:    "jfs-secret",
		SecretKeySecretKey: "secretkey",
		TokenSecret:        "",
		TokenSecretKey:     "",
		AccessKeySecret:    "jfs-secret",
		AccessKeySecretKey: "accesskey",
		FormatCmd:          "/usr/local/bin/juicefs format --trash-days=0 --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY} --storage=minio --bucket=http://minio.default.svc.cluster.local:9000/minio/test2 ${METAURL} minio",
		Name:               "minio",
		Edition:            "community",
	}
	metaurlInfo, err := GetFSInfoFromConfigMap(fakeClient, dataSet.Name, dataSet.Namespace)
	if err != nil {
		t.Errorf("GetMetaUrlInfoFromConfigMap failed.")
	}
	if len(metaurlInfo) != len(wantMetaurlInfo) {
		t.Errorf("parseCacheInfoFromConfigMap() gotMetaurlInfo = %v,\n want %v", metaurlInfo, wantMetaurlInfo)
	}
	for k, v := range metaurlInfo {
		if v != wantMetaurlInfo[k] {
			t.Errorf("parseCacheInfoFromConfigMap() got %s = %v,\n want %v", k, v, wantMetaurlInfo[k])
		}
	}
}

// Test_parseFSInfoFromConfigMap is a unit test function for the parseFSInfoFromConfigMap method.
// It validates whether the function correctly extracts and parses dataset information 
// from a given Kubernetes ConfigMap.
//
// Steps:
// 1. Define test cases with different ConfigMap data inputs.
// 2. Execute parseFSInfoFromConfigMap using the provided test cases.
// 3. Verify the returned metadata information against expected values.
// 4. Check for expected errors in erroneous cases.
// 5. Use assertions to ensure the function behaves as intended.
func Test_parseFSInfoFromConfigMap(t *testing.T) {
	type args struct {
		configMap *v1.ConfigMap
	}
	tests := []struct {
		name            string
		args            args
		wantMetaurlInfo map[string]string
		wantErr         bool
	}{
		{
			name: "test",
			args: args{
				configMap: &v1.ConfigMap{
					Data: map[string]string{
						"data": valuesConfigMapData,
					},
				},
			},
			wantMetaurlInfo: map[string]string{
				MetaurlSecret:      "jfs-secret",
				MetaurlSecretKey:   "metaurl",
				SecretKeySecret:    "jfs-secret",
				SecretKeySecretKey: "secretkey",
				TokenSecret:        "",
				TokenSecretKey:     "",
				AccessKeySecret:    "jfs-secret",
				AccessKeySecretKey: "accesskey",
				FormatCmd:          "/usr/local/bin/juicefs format --trash-days=0 --access-key=${ACCESS_KEY} --secret-key=${SECRET_KEY} --storage=minio --bucket=http://minio.default.svc.cluster.local:9000/minio/test2 ${METAURL} minio",
				Name:               "minio",
				Edition:            "community",
			},
			wantErr: false,
		},
		{
			name: "test-err",
			args: args{
				configMap: &v1.ConfigMap{
					Data: map[string]string{
						"data": "test",
					},
				},
			},
			wantMetaurlInfo: map[string]string{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetaurlInfo, err := parseFSInfoFromConfigMap(tt.args.configMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFSInfoFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMetaurlInfo, tt.wantMetaurlInfo) {
				t.Errorf("parseFSInfoFromConfigMap() gotMetaurlInfo = %v, want %v", gotMetaurlInfo, tt.wantMetaurlInfo)
			}
			if len(gotMetaurlInfo) != len(tt.wantMetaurlInfo) {
				t.Errorf("parseCacheInfoFromConfigMap() gotMetaurlInfo = %v,\n want %v", gotMetaurlInfo, tt.wantMetaurlInfo)
			}
			for k, v := range gotMetaurlInfo {
				if v != tt.wantMetaurlInfo[k] {
					t.Errorf("parseCacheInfoFromConfigMap() got %s = %v,\n want %v", k, v, tt.wantMetaurlInfo[k])
				}
			}
		})
	}
}
