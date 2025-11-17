/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

func TestSetupMasterInternal(t *testing.T) {
	pr := net.ParsePortRangeOrDie("14000-15999")
	dummyPorts := func(client client.Client) (ports []int, err error) {
		return []int{14000, 14001, 14002, 14003}, nil
	}
	err := portallocator.SetupRuntimePortAllocator(nil, pr, "bitmap", dummyPorts)
	if err != nil {
		t.Fatal(err.Error())
	}
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

	quota := resource.MustParse("1Gi")
	vineyardruntime := &datav1alpha1.VineyardRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.VineyardRuntimeSpec{
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{
					{
						MediumType: "MEM",
						Quota:      &quota,
					},
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*vineyardruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "vineyard", base.WithTieredStore(vineyardruntime.Spec.TieredStore))
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := VineyardEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.VineyardRuntime{
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Master: datav1alpha1.MasterSpec{
					VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
						Replicas: 2,
					},
				},
			},
		},
		runtimeInfo: runtimeInfo,
	}
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}

	// check release found
	patches := gomonkey.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseCommonFound)
	defer patches.Reset()
	err = engine.setupMasterInternal()
	if err != nil {
		t.Errorf("fail to exec check helm release, %v", err)
	}

	// check release error
	patches.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseErr)
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}

	// check release not found
	patches.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseCommonNotFound)

	// install release with error
	patches.ApplyFunc(helm.InstallRelease, mockExecInstallReleaseErr)
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}

	// install release successfully
	patches.ApplyFunc(helm.InstallRelease, mockExecInstallReleaseCommon)
	err = engine.setupMasterInternal()
	if err != nil {
		t.Errorf("fail to install release, %v", err)
	}
}

func TestGenerateVineyardValueFile(t *testing.T) {

	quota := resource.MustParse("1Gi")
	vineyardruntime := &datav1alpha1.VineyardRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.VineyardRuntimeSpec{
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{
					{
						MediumType: "MEM",
						Quota:      &quota,
					},
				},
			},
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*vineyardruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "vineyard", base.WithTieredStore(vineyardruntime.Spec.TieredStore))
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	engine := VineyardEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.VineyardRuntime{
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Master: datav1alpha1.MasterSpec{
					VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
						Replicas: 2,
					},
				},
			},
		},
		runtimeInfo: runtimeInfo,
	}

	_, err = engine.generateVineyardValueFile(vineyardruntime)
	if err != nil {
		t.Errorf("fail to generateVineyardValueFile %v", err)
	}
}

func TestGetConfigmapName(t *testing.T) {
	engine := VineyardEngine{
		name:       "hbase",
		engineImpl: "vineyard",
	}
	expectedResult := "hbase-vineyard-values"
	if engine.getConfigmapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
