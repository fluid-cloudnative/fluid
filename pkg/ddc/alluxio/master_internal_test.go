/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package alluxio

import (
	"fmt"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
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
	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
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
	fmt.Println(err)
	if err != nil {
		t.Errorf("fail to install release")
	}
	wrappedUnhookInstallRelease()
	wrappedUnhookCheckRelease()
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGenerateAlluxioValueFile(t *testing.T) {
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

	err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 50}, "bitmap", GetReservedPorts)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = engine.generateAlluxioValueFile(allixioruntime)
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCreateConfigMapFromFile()

	err = gohook.Hook(kubectl.CreateConfigMapFromFile, mockExecCreateConfigMapFromFileCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = engine.generateAlluxioValueFile(allixioruntime)
	if err != nil {
		t.Errorf("fail to generateAlluxioValueFile %v", err)
	}
	wrappedUnhookCreateConfigMapFromFile()
}

func TestGetConfigmapName(t *testing.T) {
	engine := AlluxioEngine{
		name:        "hbase",
		runtimeType: "alluxio",
	}
	expectedResult := "hbase-alluxio-values"
	if engine.getConfigmapName() != expectedResult {
		t.Errorf("fail to get the configmap name")
	}
}
