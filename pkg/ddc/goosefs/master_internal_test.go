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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
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

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "goosefs")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

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
		runtimeInfo: runtimeInfo,
	}

	err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}

	// check release found
	patches := gomonkey.ApplyFunc(helm.CheckRelease, mockExecCheckReleaseCommonFound)
	defer patches.Reset()

	err = engine.setupMasterInternal()

	if err != nil {

		t.Errorf("fail to exec check helm release")

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

	fmt.Println(err)

	if err != nil {

		t.Errorf("fail to install release")

	}

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

	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "fluid", "goosefs")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

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
		runtimeInfo: runtimeInfo,
	}

	err = portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, "bitmap", GetReservedPorts)
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
