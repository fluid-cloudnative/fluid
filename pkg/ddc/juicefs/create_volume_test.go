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
	"context"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

// TestJuiceFSEngine_CreateVolume 是一个单元测试函数，用于测试 JuiceFSEngine 的 CreateVolume 方法。
// 该测试函数的主要目的是验证 JuiceFSEngine 是否能够正确创建 PersistentVolume (PV) 和 PersistentVolumeClaim (PVC)。
// 测试过程包括：
// 1. 构建 RuntimeInfo 对象。
// 2. 创建测试数据集 (Dataset) 对象。
// 3. 使用 fake client 初始化 JuiceFSEngine。
// 4. 调用 CreateVolume 方法创建 PV 和 PVC。
// 5. 验证 PV 和 PVC 是否成功创建。
// 如果任何步骤失败，测试将返回错误。
//
// 参数:
//   - t *testing.T: testing.T 对象，用于管理测试状态和支持测试日志输出。
//
// 返回值:
//   - 无返回值。如果测试失败，会通过 t.Errorf 报告错误。
func TestJuiceFSEngine_CreateVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", common.JuiceFSRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "hbase",
		runtimeInfo: runtimeInfo,
		runtime: &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
		},
	}

	err = engine.CreateVolume()
	if err != nil {
		t.Errorf("fail to exec CreateVolume with error %v", err)
	}

	var pvs v1.PersistentVolumeList
	err = client.List(context.TODO(), &pvs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvs.Items) != 1 {
		t.Errorf("fail to create the pv")
	}

	var pvcs v1.PersistentVolumeClaimList
	err = client.List(context.TODO(), &pvcs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvcs.Items) != 1 {
		t.Errorf("fail to create the pvc")
	}
}

func TestJuiceFSEngine_createFusePersistentVolume(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", common.JuiceFSRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "test",
		runtimeInfo: runtimeInfo,
	}

	err = engine.createFusePersistentVolume()
	if err != nil {
		t.Errorf("fail to exec createFusePersistentVolume with error %v", err)
	}

	var pvs v1.PersistentVolumeList
	err = client.List(context.TODO(), &pvs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvs.Items) != 1 {
		t.Errorf("fail to create the pv")
	}
}

func TestJuiceFSEngine_createFusePersistentVolumeClaim(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "fluid", common.JuiceFSRuntime)
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	testDatasetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "juicefs",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	testObjs := []runtime.Object{}
	for _, datasetInput := range testDatasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &JuiceFSEngine{
		Client:      client,
		Log:         fake.NullLogger(),
		namespace:   "fluid",
		name:        "test",
		runtimeInfo: runtimeInfo,
	}

	err = engine.createFusePersistentVolumeClaim()
	if err != nil {
		t.Errorf("fail to exec createFusePersistentVolumeClaim with error %v", err)
	}

	var pvcs v1.PersistentVolumeClaimList
	err = client.List(context.TODO(), &pvcs)
	if err != nil {
		t.Errorf("fail to exec the function with error %v", err)
		return
	}
	if len(pvcs.Items) != 1 {
		t.Errorf("fail to create the pvc")
	}
}
