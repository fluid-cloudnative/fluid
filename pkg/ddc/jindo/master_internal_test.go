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

package jindo

import (
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

// TestSetupMasterInternal is a unit test function that verifies the behavior of the setupMasterInternal function in the JindoEngine.
// It leverages mock functions and gohook to simulate various scenarios related to checking and installing Helm releases.
// The test cases include:
// 1. Handling errors when creating runtime information.
// 2. Handling different responses when checking for Helm releases, including:
//    - The release is found.
//    - The release is not found.
//    - An error occurs while checking the release.
// 3. Handling different responses when installing a Helm release, including:
//    - Successful installation.
//    - Failure to install the release.
// 4. Setting up a fake Kubernetes client with test objects to simulate real interactions.
// 5. Using gohook to dynamically hook and unhook function calls to test various outcomes.
func TestSetupMasterInternal(t *testing.T) {
	mockExecCheckReleaseCommonFound := func(name string, namespace string) (exist bool, err error) {
		// Simulates that the Helm release exists and no error occurs.
		return true, nil
	}
	mockExecCheckReleaseCommonNotFound := func(name string, namespace string) (exist bool, err error) {
		// Simulates that the Helm release does not exist.
		return false, nil
	}
	mockExecCheckReleaseErr := func(name string, namespace string) (exist bool, err error) {
		// Simulates an error while checking for the Helm release.
		return false, errors.New("fail to check release")
	}
	mockExecInstallReleaseCommon := func(name string, namespace string, valueFile string, chartName string) error {
		// Simulates a successful Helm release installation.
		return nil
	}
	mockExecInstallReleaseErr := func(name string, namespace string, valueFile string, chartName string) error {
		// Simulates a failure in installing the Helm release.
		return errors.New("fail to install dataload chart")
	}

	wrappedUnhookCheckRelease := func() {
		// Unhooks the mock function for helm.CheckRelease and logs a fatal error if it fails.
		err := gohook.UnHook(helm.CheckRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookInstallRelease := func() {
		// Unhooks the mock function for helm.InstallRelease and logs a fatal error if it fails.
		err := gohook.UnHook(helm.InstallRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	// Creates a mock JindoRuntime object in the "fluid" namespace.
	allixioruntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
	}
	// Initializes a slice of runtime objects with the JindoRuntime instance.
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*allixioruntime).DeepCopy())

	// Creates dataset objects and adds them to testObjs for simulating Kubernetes resources.
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

	// Initializes a fake Kubernetes client with test objects.
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	// Builds runtime information for the JindoRuntime.
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "jindo")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Initializes the JindoEngine instance with test parameters.
	engine := JindoEngine{
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
		runtimeInfo: runtimeInfo,
	}

	// Sets up the runtime port allocator.
	err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test case: Ensures setupMasterInternal catches an error when no release is found.
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}

	// Test case: Hook helm.CheckRelease with mockExecCheckReleaseCommonFound to simulate the release being found.
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = engine.setupMasterInernal()
	wrappedUnhookCheckRelease()

	// Test case: Hook helm.CheckRelease with mockExecCheckReleaseErr to simulate an error while checking the release.
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCheckRelease()

	// Test case: Hook helm.CheckRelease with mockExecCheckReleaseCommonNotFound to simulate the release not being found.
	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommonNotFound, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Test case: Hook helm.InstallRelease with mockExecInstallReleaseErr to simulate a failed installation.
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.setupMasterInernal()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookInstallRelease()

	// Test case: Hook helm.InstallRelease with mockExecInstallReleaseCommon to simulate a successful installation.
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = engine.setupMasterInernal()
	wrappedUnhookInstallRelease()
}

func TestGenerateJindoValueFile(t *testing.T) {
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
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "jindo")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}
	result := resource.MustParse("20Gi")
	engine := JindoEngine{
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
		runtimeInfo: runtimeInfo,
	}

	err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = engine.generateJindoValueFile()
	if err != nil {
		t.Errorf("fail to catch the error")
	}
}

func TestGetConfigmapName(t *testing.T) {
	engine := JindoEngine{
		name:       "hbase",
		engineImpl: "jindo",
	}
	expectedResult := "hbase-jindo-values"
	if engine.getHelmValuesConfigmapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
