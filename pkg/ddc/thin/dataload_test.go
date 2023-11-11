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

package thin

import (
	"errors"
	"testing"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

func TestThinEngine_CreateDataLoadJob(t *testing.T) {
	mockExecCheckReleaseCommon := func(name string, namespace string) (exist bool, err error) {
		return false, nil
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

	targetDataLoad := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
		},
	}
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "fluid",
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := ThinEngine{
		name: "test",
	}
	ctx := cruntime.ReconcileRequestContext{
		Log:      fake.NullLogger(),
		Client:   client,
		Recorder: record.NewFakeRecorder(1),
	}

	err := gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err != nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookInstallRelease()

	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookCheckRelease()
}

func TestThinEngine_CheckExistenceOfPath(t *testing.T) {
	mockExecNotExist := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "does not exist", "", errors.New("other error")
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	engine := ThinEngine{
		namespace: "fluid",
		name:      "test",
		Log:       fake.NullLogger(),
	}

	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExecNotExist, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	targetDataload := datav1alpha1.DataLoad{
		Spec: datav1alpha1.DataLoadSpec{
			Target: []datav1alpha1.TargetPath{{
				Path:     "/tmp",
				Replicas: 1,
			}},
		},
	}
	notExist, err := engine.CheckExistenceOfPath(targetDataload)
	if !(err == nil && notExist == true) {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhook()
}
