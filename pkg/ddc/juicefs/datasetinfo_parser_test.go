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

// Test_parseCacheInfoFromConfigMap 是 parseCacheInfoFromConfigMap 函数的单元测试函数。
// 该测试函数用于验证 parseCacheInfoFromConfigMap 函数是否能够正确解析 ConfigMap 中的数据，
// 并返回预期的缓存信息，同时检查函数在不同输入条件下的错误处理是否符合预期。
//
// 测试用例设计：
// 1. 正常用例：输入包含有效数据的 ConfigMap，验证函数是否能正确解析并返回缓存信息。
// 2. 错误用例：输入包含无效数据的 ConfigMap，验证函数是否能正确处理错误并返回预期结果。
//
// 输入：
// - 无显式输入参数，测试用例通过结构体定义。
//
// 输出：
// - 无显式返回值，测试结果通过 testing.T 的方法输出。
func Test_parseCacheInfoFromConfigMap(t *testing.T) {
    // 定义 args 结构体，用于表示测试用例的输入参数
    type args struct {
        configMap *v1.ConfigMap // configMap 是测试函数的输入参数，类型为 *v1.ConfigMap
    }

    // 定义测试用例集合
    tests := []struct {
        name          string            // 测试用例的名称
        args          args              // 测试用例的输入参数
        wantCacheInfo map[string]string // 期望的缓存信息结果
        wantErr       bool              // 是否期望返回错误
    }{
        // 第一个测试用例：正常解析 ConfigMap 数据
        {
            name: "parseCacheInfoFromConfigMap", // 测试用例名称
            args: args{configMap: &v1.ConfigMap{ // 输入参数
                Data: map[string]string{ // ConfigMap 的数据字段
                    "data": valuesConfigMapData, // 假设 valuesConfigMapData 是一个预定义的字符串，包含有效的配置数据
                },
            }},
            wantCacheInfo: map[string]string{ // 期望的缓存信息结果
                "mountpath": "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse", // 预期的挂载路径
                "edition":   "community", // 预期的版本信息
            },
            wantErr: false, // 不期望返回错误
        },
        // 第二个测试用例：解析错误的 ConfigMap 数据
        {
            name: "parseCacheInfoFromConfigMap-err", // 测试用例名称
            args: args{configMap: &v1.ConfigMap{ // 输入参数
                Data: map[string]string{ // ConfigMap 的数据字段
                    "data": `test`, // 无效的配置数据
                },
            }},
            wantCacheInfo: nil, // 期望的缓存信息结果为 nil
            wantErr:       true, // 期望返回错误
        },
    }

    // 遍历测试用例集合，逐个运行测试
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { // 使用 t.Run 运行子测试
            // 调用待测试的函数 parseCacheInfoFromConfigMap，获取返回值和错误
            gotPorts, err := parseCacheInfoFromConfigMap(tt.args.configMap)

            // 检查错误是否符合预期
            if (err != nil) != tt.wantErr {
                t.Errorf("parseCacheInfoFromConfigMap() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // 检查返回的缓存信息是否符合预期
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
