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

package alluxio

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTestAlluxioEngine(client client.Client, name string, namespace string) *AlluxioEngine {
	runTime := &datav1alpha1.AlluxioRuntime{}
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "alluxio")
	engine := &AlluxioEngine{
		runtime:     runTime,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	return engine
}

func TestAlluxioEngine_GetDeprecatedCommonLabelname(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-hbase",
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			out:       "data.fluid.io/storage-fluid-hadoop",
		},
		{
			name:      "fluid",
			namespace: "test",
			out:       "data.fluid.io/storage-test-fluid",
		},
	}
	for _, test := range testCases {
		out := utils.GetCommonLabelName(true, test.namespace, test.name, "")
		if out != test.out {
			t.Errorf("input parameter is %s-%s,expected %s, got %s", test.namespace, test.name, test.out, out)
		}
	}

}

// TestAlluxioEngine_HasDeprecatedCommonLabelname 测试Alluxio引擎检测废弃标签的功能
// 该测试验证HasDeprecatedCommonLabelname方法能否正确识别DaemonSet中是否包含废弃的标签格式
func TestAlluxioEngine_HasDeprecatedCommonLabelname(t *testing.T) {
	// 创建带有特定节点选择器的DaemonSet
	// 该DaemonSet使用了废弃的标签格式: "data.fluid.io/storage-<runtime>-<dataset>"
	daemonSetWithSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase-worker",  // 工作负载名称
			Namespace: "fluid",         // 所属命名空间
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					// 节点选择器使用废弃标签格式
					NodeSelector: map[string]string{
						"data.fluid.io/storage-fluid-hbase": "selector",
					},
				},
			},
		},
	}

	// 创建另一个DaemonSet（虽然名称不同但使用了相同的标签格式）
	daemonSetWithoutSelector := &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hadoop-worker",  // 不同工作负载
			Namespace: "fluid",          // 相同命名空间
		},
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					// 同样使用废弃标签格式
					NodeSelector: map[string]string{
						"data.fluid.io/storage-fluid-hbase": "selector",
					},
				},
			},
		},
	}

	// 准备测试用的Kubernetes API对象
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, daemonSetWithSelector)
	runtimeObjs = append(runtimeObjs, daemonSetWithoutSelector)
	
	// 创建Scheme并注册API类型
	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(v1.SchemeGroupVersion, daemonSetWithSelector)
	
	// 使用伪客户端(fake client)模拟Kubernetes API
	fakeClient := fake.NewFakeClientWithScheme(scheme, runtimeObjs...)

	// 定义测试用例
	testCases := []struct {
		name      string // 数据集名称
		namespace string // 命名空间
		out       bool   // 期望返回值
		isErr     bool   // 是否期望错误
	}{
		{
			name:      "hbase",  // 匹配存在的DaemonSet
			namespace: "fluid",
			out:       true,     // 应检测到废弃标签
			isErr:     false,
		},
		{
			name:      "none",   // 不存在的数据集
			namespace: "fluid",
			out:       false,    // 不应检测到废弃标签
			isErr:     false,
		},
		{
			name:      "hadoop", // 存在但名称不匹配的DaemonSet
			namespace: "fluid",
			out:       false,    // 不应检测到废弃标签
			isErr:     false,
		},
	}

	// 遍历执行所有测试用例
	for _, test := range testCases {
		// 为当前测试用例创建Alluxio引擎实例
		engine := getTestAlluxioEngine(fakeClient, test.name, test.namespace)
		
		// 调用被测试方法
		out, err := engine.HasDeprecatedCommonLabelname()
		
		// 验证返回值是否符合预期
		if out != test.out {
			t.Errorf(
				"测试数据集 %s/%s 失败: 期望 %t, 实际 %t", 
				test.namespace, test.name, test.out, out,
			)
		}
		
		// 验证错误情况是否符合预期
		isErr := err != nil
		if isErr != test.isErr {
			t.Errorf(
				"测试数据集 %s/%s 错误验证失败: 期望错误 %t, 实际 %t", 
				test.namespace, test.name, test.isErr, isErr,
			)
		}
	}
}
