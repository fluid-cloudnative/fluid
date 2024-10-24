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

package alluxio

import (
	"fmt"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

	allixioruntime := &datav1alpha1.AlluxioRuntime{
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

	engine := AlluxioEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
					Enabled: false,
				},
				Master: datav1alpha1.AlluxioCompTemplateSpec{
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

func TestGenerateAlluxioValueFile(t *testing.T) {
	allixioruntime := &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.AlluxioRuntimeSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-runtime"}},
			Master: datav1alpha1.AlluxioCompTemplateSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-master"}},
			},
			Worker: datav1alpha1.AlluxioCompTemplateSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-worker"}},
			},
			Fuse: datav1alpha1.AlluxioFuseSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: "secret-fuse"}},
			},
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

	engine := AlluxioEngine{
		name:      "hbase",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
		runtime: &datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
					Enabled: false,
				},
				Master: datav1alpha1.AlluxioCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
	}

	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = engine.generateAlluxioValueFile(allixioruntime)
	if err != nil {
		t.Errorf("fail to catch the error")
	}

	// create it again still success
	_, err = engine.generateAlluxioValueFile(allixioruntime)
	if err != nil {
		t.Errorf("fail to generateAlluxioValueFile %v", err)
	}
}

func TestGetConfigmapName(t *testing.T) {
	engine := AlluxioEngine{
		name:       "hbase",
		engineImpl: "alluxio",
	}
	expectedResult := "hbase-alluxio-values"
	if engine.getHelmValuesConfigMapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
