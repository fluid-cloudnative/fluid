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

package vineyard

import (
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
)

func TestSetupMasterInternal(t *testing.T) {
	mockExecCreateConfigMapFromFileCommon := func(name string, key, fileName string, namespace string) (err error) {
		return nil
	}
	mockExecCreateConfigMapFromFileErr := func(name string, key, fileName string, namespace string) (err error) {
		return errors.New("fail to exec command")
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

	wrappedUnhookCreateConfigMapFromFile := func() {
		err := gohook.UnHook(kubectl.CreateConfigMapFromFile)
		if err != nil {
			t.Fatal(err.Error())
		}
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
	}
	err := gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCreateConfigMapFromFile()

	// create configmap successfully
	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileCommon, nil)
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
		t.Errorf("fail to exec check helm release")
	}
	wrappedUnhookCheckRelease()

	// check release error
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInternal()
	if err == nil {
		t.Errorf("fail to catch the error")
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
	if err != nil {
		t.Errorf("fail to install release")
	}
	wrappedUnhookInstallRelease()
	wrappedUnhookCheckRelease()
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGenerateVineyardValueFile(t *testing.T) {
	mockExecCreateConfigMapFromFileCommon := func(name string, key, fileName string, namespace string) (err error) {
		return nil
	}
	mockExecCreateConfigMapFromFileErr := func(name string, key, fileName string, namespace string) (err error) {
		return errors.New("fail to exec command")
	}

	wrappedUnhookCreateConfigMapFromFile := func() {
		err := gohook.UnHook(kubectl.CreateConfigMapFromFile)
		if err != nil {
			t.Fatal(err.Error())
		}
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
	}

	err := gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = engine.generateVineyardValueFile(vineyardruntime)
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCreateConfigMapFromFile()

	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = engine.generateVineyardValueFile(vineyardruntime)
	if err != nil {
		t.Errorf("fail to generateVineyardValueFile %v", err)
	}
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGetConfigmapName(t *testing.T) {
	engine := VineyardEngine{
		name:        "hbase",
		runtimeType: "vineyard",
	}
	expectedResult := "hbase-vineyard-values"
	if engine.getConfigmapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
