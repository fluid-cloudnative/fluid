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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGenerateDataLoadValueFile(t *testing.T) {
	// 定义一个数据集（dataset），这是测试所需的输入之一
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",  // 数据集名称
				Namespace: "fluid",         // 数据集所在的命名空间
			},
			Spec: datav1alpha1.DatasetSpec{},  // 数据集的规格，当前没有配置内容
		},
	}

	// 定义一个 JindoRuntime 对象，这是测试所需的第二个输入
	jindo := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",  // JindoRuntime 的名称
			Namespace: "fluid",         // JindoRuntime 的命名空间
		},
	}

	// 配置 JindoRuntime 的规格，设置了 `Secret` 和 `TieredStore` 等字段
	jindo.Spec = datav1alpha1.JindoRuntimeSpec{
		Secret: "secret",  // JindoRuntime 的 Secret
		TieredStore: datav1alpha1.TieredStore{
			Levels: []datav1alpha1.Level{{
				MediumType: common.Memory,  // 存储介质类型：内存
				Quota:      resource.NewQuantity(1, resource.BinarySI),  // 存储配额：1 字节
				High:       "0.8",  // 高水位线
				Low:        "0.1",  // 低水位线
			}},
		},
	}

	// 将 JindoRuntime 类型添加到测试方案中
	testScheme.AddKnownTypes(datav1alpha1.GroupVersion, jindo)

	// 将数据集对象和 JindoRuntime 对象添加到测试对象列表中
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	testObjs = append(testObjs, jindo.DeepCopy())

	// 创建一个假的 Kubernetes 客户端并初始化 context
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	context := cruntime.ReconcileRequestContext{
		Client: client,
	}

	// 定义两个 DataLoad 对象，一个没有目标路径，一个有目标路径
	dataLoadNoTarget := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataload",  // DataLoad 对象名称
			Namespace: "fluid",         // DataLoad 对象命名空间
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "test-dataset",  // 关联的数据集名称
				Namespace: "fluid",         // 数据集命名空间
			},
		},
	}
	dataLoadWithTarget := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataload",  // DataLoad 对象名称
			Namespace: "fluid",         // DataLoad 对象命名空间
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "test-dataset",  // 关联的数据集名称
				Namespace: "fluid",         // 数据集命名空间
			},
			Target: []datav1alpha1.TargetPath{  // 目标路径，数据加载目标
				{
					Path:     "/test",  // 目标路径
					Replicas: 1,        // 副本数量
				},
			},
		},
	}

	// 定义测试用例，包含 DataLoad 对象和期望的文件路径
	testCases := []struct {
		dataLoad       datav1alpha1.DataLoad  // 测试用例中的 DataLoad 对象
		expectFileName string                  // 期望生成的文件路径
	}{
		{
			dataLoad:       dataLoadNoTarget,  // 不带目标路径的 DataLoad
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-dataload-loader-values.yaml"),
		},
		{
			dataLoad:       dataLoadWithTarget,  // 带有目标路径的 DataLoad
			expectFileName: filepath.Join(os.TempDir(), "fluid-test-dataload-loader-values.yaml"),
		},
	}

	// 循环遍历每个测试用例，执行生成文件路径的逻辑
	for _, test := range testCases {
		engine := JindoEngine{}  // 初始化 JindoEngine 实例
		// 调用 generateDataLoadValueFile 函数生成文件路径
		if fileName, _ := engine.generateDataLoadValueFile(context, &test.dataLoad); !strings.Contains(fileName, test.expectFileName) {
			// 如果生成的文件路径不包含期望的文件路径，测试失败
			t.Errorf("fail to generate the dataload value file")
		}
	}
}

func Test_genDataLoadValue(t *testing.T) {
	testCases := map[string]struct {
		image         string
		targetDataset *datav1alpha1.Dataset
		dataload      *datav1alpha1.DataLoad
		want          *cdataload.DataLoadValue
	}{
		"test case with scheduler name": {
			image: "fluid:v0.0.1",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataload",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: "fluid",
					},
					Target: []datav1alpha1.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
				},
			},
			want: &cdataload.DataLoadValue{
				Name:           "test-dataload",
				OwnerDatasetId: "fluid-test-dataset",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
				},
			},
		},
		"test case with affinity": {
			image: "fluid:v0.0.1",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataload",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: "fluid",
					},
					Target: []datav1alpha1.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "topology.kubernetes.io/zone",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"antarctica-east1",
													"antarctica-west1",
												},
											},
										},
									},
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Weight: 1,
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "another-node-label-key",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"another-node-label-value",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: &cdataload.DataLoadValue{
				Name:           "test-dataload",
				OwnerDatasetId: "fluid-test-dataset",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "topology.kubernetes.io/zone",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"antarctica-east1",
													"antarctica-west1",
												},
											},
										},
									},
								},
							},
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Weight: 1,
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "another-node-label-key",
												Operator: corev1.NodeSelectorOpIn,
												Values: []string{
													"another-node-label-value",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"test case with node selector": {
			image: "fluid:v0.0.1",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataload",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: "fluid",
					},
					Target: []datav1alpha1.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
					NodeSelector: map[string]string{
						"diskType": "ssd",
					},
				},
			},
			want: &cdataload.DataLoadValue{
				Name:           "test-dataload",
				OwnerDatasetId: "fluid-test-dataset",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
					NodeSelector: map[string]string{
						"diskType": "ssd",
					},
				},
			},
		},
		"test case with tolerations": {
			image: "fluid:v0.0.1",
			targetDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "spark",
							MountPoint: "local://mnt/data0",
							Path:       "/mnt",
						},
					},
				},
			},
			dataload: &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataload",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DataLoadSpec{
					Dataset: datav1alpha1.TargetDataset{
						Name:      "test-dataset",
						Namespace: "fluid",
					},
					Target: []datav1alpha1.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					SchedulerName: "scheduler-test",
					Tolerations: []corev1.Toleration{
						{
							Key:      "example-key",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			want: &cdataload.DataLoadValue{
				Name:           "test-dataload",
				OwnerDatasetId: "fluid-test-dataset",
				Owner: &common.OwnerReference{
					APIVersion:         "/",
					Enabled:            true,
					Name:               "test-dataload",
					BlockOwnerDeletion: false,
					Controller:         true,
				},
				DataLoadInfo: cdataload.DataLoadInfo{
					BackoffLimit:  3,
					Image:         "fluid:v0.0.1",
					TargetDataset: "test-dataset",
					SchedulerName: "scheduler-test",
					TargetPaths: []cdataload.TargetPath{
						{
							Path:     "/test",
							Replicas: 1,
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{},
					Tolerations: []corev1.Toleration{
						{
							Key:      "example-key",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
		},
	}
	engine := AlluxioEngine{
		namespace: "fluid",
		name:      "test",
		Log:       fake.NullLogger(),
	}
	for k, item := range testCases {
		got, _ := engine.genDataLoadValue(item.image, item.targetDataset, item.dataload)
		if !reflect.DeepEqual(got, item.want) {
			t.Errorf("case %s, got %v,want:%v", k, got, item.want)
		}
	}
}

func TestCheckRuntimeReady(t *testing.T) {
	mockExecCommon := func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "", "", nil
	}
	mockExecErr := func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "err", "", errors.New("error")
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainerWithFullOutput)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	engine := AlluxioEngine{
		namespace: "fluid",
		name:      "hbase",
		Log:       fake.NullLogger(),
	}

	err := gohook.Hook(kubeclient.ExecCommandInContainerWithFullOutput, mockExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if ready := engine.CheckRuntimeReady(); ready != true {
		fmt.Println(ready)
		t.Errorf("fail to exec the function CheckRuntimeReady")
	}
	wrappedUnhook()

	err = gohook.Hook(kubeclient.ExecCommandInContainerWithFullOutput, mockExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if ready := engine.CheckRuntimeReady(); ready != false {
		fmt.Println(ready)
		t.Errorf("fail to exec the function CheckRuntimeReady")
	}
	wrappedUnhook()
}
