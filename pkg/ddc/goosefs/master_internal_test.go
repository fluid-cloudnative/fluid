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

package goosefs

import (
	"fmt"

	"github.com/brahma-adshonor/gohook"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"

	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/util/net"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	"testing"
)

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

	allixioruntime := &datav1alpha1.GooseFSRuntime{

		ObjectMeta: metav1.ObjectMeta{

			Name: "hbase",

			Namespace: "fluid",
		},
	}

	testObjs := []runtime.Object{}

	testObjs = append(testObjs, (*allixioruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{

		{

			ObjectMeta: metav1.ObjectMeta{

				Name: "hbase",

				Namespace: "fluid",
			},
		},
	}

	for _, datasetInput := range datasetInputs {

		testObjs = append(testObjs, datasetInput.DeepCopy())

	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := GooseFSEngine{

		name: "hbase",

		namespace: "fluid",

		Client: client,

		Log: fake.NullLogger(),

		runtime: &datav1alpha1.GooseFSRuntime{

			Spec: datav1alpha1.GooseFSRuntimeSpec{

				APIGateway: datav1alpha1.GooseFSCompTemplateSpec{

					Enabled: false,
				},

				Master: datav1alpha1.GooseFSCompTemplateSpec{

					Replicas: 2,
				},
			},
		},
	}

	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
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

	fmt.Println(err)

	if err != nil {

		t.Errorf("fail to install release")

	}

	wrappedUnhookInstallRelease()

	wrappedUnhookCheckRelease()

}

func TestGenerateGooseFSValueFile(t *testing.T) {
	allixioruntime := &datav1alpha1.GooseFSRuntime{

		ObjectMeta: metav1.ObjectMeta{

			Name: "hbase",

			Namespace: "fluid",
		},
	}

	testObjs := []runtime.Object{}

	testObjs = append(testObjs, (*allixioruntime).DeepCopy())

	datasetInputs := []datav1alpha1.Dataset{

		{

			ObjectMeta: metav1.ObjectMeta{

				Name: "hbase",

				Namespace: "fluid",
			},
		},
	}

	for _, datasetInput := range datasetInputs {

		testObjs = append(testObjs, datasetInput.DeepCopy())

	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := GooseFSEngine{

		name: "hbase",

		namespace: "fluid",

		Client: client,

		Log: fake.NullLogger(),

		runtime: &datav1alpha1.GooseFSRuntime{

			Spec: datav1alpha1.GooseFSRuntimeSpec{

				APIGateway: datav1alpha1.GooseFSCompTemplateSpec{

					Enabled: false,
				},

				Master: datav1alpha1.GooseFSCompTemplateSpec{

					Replicas: 2,
				},
			},
		},
	}

	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = engine.generateGooseFSValueFile(allixioruntime)

	if err != nil {

		t.Errorf("fail to exec the function")

	}

}

func TestGetConfigmapName(t *testing.T) {

	engine := GooseFSEngine{

		name: "hbase",

		engineImpl: "goosefs",
	}

	expectedResult := "hbase-goosefs-values"

	if engine.getHelmValuesConfigMapName() != expectedResult {

		t.Errorf("fail to get the configmap name")

	}

}
