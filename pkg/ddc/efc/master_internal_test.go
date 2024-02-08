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

package efc

import (
	"errors"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"k8s.io/apimachinery/pkg/util/net"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
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

	efcruntime := &datav1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
	}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, (*efcruntime).DeepCopy())

	var datasetInputs = []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := EFCEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime:   efcruntime,
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
	if err != nil {
		t.Errorf("fail to install release")
	}
	wrappedUnhookInstallRelease()
	wrappedUnhookCheckRelease()
}

func TestGenerateEFCValueFile(t *testing.T) {
	testObjs := []runtime.Object{}
	efcruntime := &datav1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.EFCRuntimeSpec{},
	}
	testObjs = append(testObjs, (*efcruntime).DeepCopy())

	var datasetInputs = []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := EFCEngine{
		name:      "test",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime:   efcruntime,
	}

	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = engine.generateEFCValueFile(efcruntime)
	if err != nil {
		t.Errorf("fail to exec the function: %v", err)
	}

}
