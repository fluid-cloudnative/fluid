/*
Copyright 2022 The Fluid Authors.

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

package jindofsx

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
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

	allixioruntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*allixioruntime).DeepCopy())

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

	engine := JindoFSxEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
	}
	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
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
	_ = engine.setupMasterInernal()
	wrappedUnhookCheckRelease()

	// check release error
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
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
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookInstallRelease()

	// install release successfully
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = engine.setupMasterInernal()
	wrappedUnhookInstallRelease()
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGenerateJindoValueFile(t *testing.T) {
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

	jindoruntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*jindoruntime).DeepCopy())

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
	result := resource.MustParse("20Gi")
	engine := JindoFSxEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 2,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
			},
		},
	}

	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = engine.generateJindoValueFile()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCreateConfigMapFromFile()

	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGetConfigmapName(t *testing.T) {
	engine := JindoFSxEngine{
		name:       "hbase",
		engineImpl: "jindo",
	}
	expectedResult := "hbase-jindo-values"
	if engine.getHelmValuesConfigMapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
