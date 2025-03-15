/*
Copyright 2023 The Fluid Authors.

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
	"errors"
	"fmt"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/net"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = rbacv1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

func TestSetupMasterInternal(t *testing.T) {
	mockExecCheckReleaseCommonFound := func(name string, namespace string) (exist bool, err error) {
		return true, nil
	}
	mockExecCheckReleaseCommonNotFound := func(name string, namespace string) (exist bool, err error) {
		return false, nil
	}
	mockExecCheckReleaseErr := func(name string, namespace string) (exist bool, err error) {
		return false, errors.New("fail to check release")
	}
	mockExecInstallReleaseCommon := func(name string, namespace string, valueFile string, chartName string) error {
		return nil
	}
	mockExecInstallReleaseErr := func(name string, namespace string, valueFile string, chartName string) error {
		return errors.New("fail to install dataload chart")
	}

	wrappedUnhookCheckRelease := func() {
		err := gohook.UnHook(helm.CheckRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookInstallRelease := func() {
		err := gohook.UnHook(helm.InstallRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	juicefsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"metaurl": []byte("test"),
		},
	}
	juicefsruntime := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsruntime).DeepCopy(), (*juicefsSecret).DeepCopy())

	var datasetInputs = []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{
				MountPoint: "juicefs://mnt",
				Name:       "test",
				EncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "metaurl",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: "test",
							Key:  "metaurl",
						},
					},
				}},
			}}},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "juicefs")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
		runtimeInfo: runtimeInfo,
	}
	err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}

	// check release found
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInternal()
	if err != nil {
		t.Errorf("fail to exec check helm release: %v", err)
	}
	wrappedUnhookCheckRelease()

	// check release error
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error: %v", err)
	}
	wrappedUnhookCheckRelease()

	// check release not found
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonNotFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// install release with error
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookInstallRelease()

	// install release successfully
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInternal()
	fmt.Println(err)
	if err != nil {
		t.Errorf("fail to install release")
	}
	wrappedUnhookInstallRelease()
	wrappedUnhookCheckRelease()
}
// TestGenerateJuiceFSValueFile 单元测试，验证 JuiceFS 引擎生成配置文件的能力
// 测试场景覆盖以下关键点：
// 1. 模拟 Kubernetes 环境中的核心资源（Secret/JuiceFSRuntime/Dataset）
// 2. 验证配置文件的生成逻辑与错误处理机制
// 3. 测试数据包括：
//    - Secret 对象：存储 JuiceFS 的元数据服务连接信息（metaurl）
//    - JuiceFSRuntime 对象：定义分布式缓存系统的存储配置（路径、介质类型、配额等）
//    - Dataset 对象：描述数据集的挂载点信息及加密参数
// 4. 使用 fake Kubernetes client 模拟 API 交互，实现测试环境隔离
// 5. 验证引擎运行时信息的正确构建（runtimeInfo）
// 6. 测试核心方法 generateJuicefsValueFile 的健壮性，确保生成符合预期的 JuiceFS 配置文件
func TestGenerateJuiceFSValueFile(t *testing.T) {
	juicefsSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Data: map[string][]byte{
			"metaurl": []byte("test"),
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*juicefsSecret).DeepCopy())
	juicefsruntime := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.JuiceFSRuntimeSpec{
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{{
					MediumType: "SSD",
					Path:       "/data",
					Quota:      resource.NewQuantity(1024, resource.BinarySI),
				}},
			},
		},
	}
	testObjs = append(testObjs, (*juicefsruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{{
				MountPoint: "juicefs:///mnt/test",
				Name:       "test",
				EncryptOptions: []datav1alpha1.EncryptOption{{
					Name: "metaurl",
					ValueFrom: datav1alpha1.EncryptOptionSource{
						SecretKeyRef: datav1alpha1.SecretKeySelector{
							Name: "test",
							Key:  "metaurl",
						},
					},
				}},
			}},
		},
	}}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "juicefs")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := JuiceFSEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JuiceFSRuntime{
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Fuse: datav1alpha1.JuiceFSFuseSpec{},
			},
		},
		runtimeInfo: runtimeInfo,
	}

	_, err = engine.generateJuicefsValueFile(juicefsruntime)
	if err != nil {
		t.Errorf("fail to exec the function: %v", err)
	}
}
