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

package efc

import (
	"errors"
	"testing"

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc/operations"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
)

func TestEFCEngine_CreateDataLoadJob(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efcdemo-efc-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}

	mockExecCheckReleaseCommon := func(name string, namespace string) (exist bool, err error) {
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

	targetDataLoad := datav1alpha1.DataLoad{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efcdemo-dataload",
			Namespace: "fluid",
		},
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "efcdemo",
				Namespace: "fluid",
			},
		},
	}
	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "efcdemo",
				Namespace: "fluid",
			},
		},
	}
	statefulsetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "efcdemo-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"a": "b"},
				},
			},
		},
	}
	podListInputs := []v1.PodList{{
		Items: []v1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"a": "b"},
			},
		}},
	}}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, configMap)
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, statefulsetInput := range statefulsetInputs {
		testObjs = append(testObjs, statefulsetInput.DeepCopy())
	}
	for _, podInput := range podListInputs {
		testObjs = append(testObjs, podInput.DeepCopy())
	}
	testScheme.AddKnownTypes(v1.SchemeGroupVersion, configMap)
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
	engine := &EFCEngine{
		name:      "efcdemo",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}
	ctx := cruntime.ReconcileRequestContext{
		Log:      fake.NullLogger(),
		Client:   client,
		Recorder: record.NewFakeRecorder(1),
	}

	err := gohook.Hook(helm.CheckRelease, mockExecCheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err == nil {
		t.Errorf("fail to catch the error: %v", err)
	}
	wrappedUnhookCheckRelease()

	err = gohook.Hook(helm.CheckRelease, mockExecCheckReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err == nil {
		t.Errorf("fail to catch the error: %v", err)
	}
	wrappedUnhookInstallRelease()

	err = gohook.Hook(helm.InstallRelease, mockExecInstallReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.CreateDataLoadJob(ctx, targetDataLoad)
	if err != nil {
		t.Errorf("fail to exec the function: %v", err)
	}
	wrappedUnhookCheckRelease()
}

func TestEFCEngine_GenerateDataLoadValueFileWithRuntimeHDD(t *testing.T) {
}

func TestEFCEngine_GenerateDataLoadValueFileWithRuntime(t *testing.T) {
}

func TestEFCEngine_CheckExistenceOfPath(t *testing.T) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efcdemo-efc-values",
			Namespace: "fluid",
		},
		Data: map[string]string{
			"data": valuesConfigMapData,
		},
	}

	datasetInputs := []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "efcdemo",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		},
	}

	statefulsetInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "efcdemo-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"a": "b"},
				},
			},
		},
	}
	podListInputs := []v1.PodList{{
		Items: []v1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "fluid",
				Labels:    map[string]string{"a": "b"},
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				Conditions: []v1.PodCondition{{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				}},
			},
		}},
	}}
	testObjs := []runtime.Object{}
	testObjs = append(testObjs, configMap)
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}
	for _, statefulsetInput := range statefulsetInputs {
		testObjs = append(testObjs, statefulsetInput.DeepCopy())
	}
	for _, podInput := range podListInputs {
		testObjs = append(testObjs, podInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	mockNotExist := func(a operations.EFCFileUtils, efcSubPath string) (found bool, err error) {
		return false, errors.New("other error")
	}
	mockExist := func(a operations.EFCFileUtils, efcSubPath string) (found bool, err error) {
		return true, nil
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(operations.EFCFileUtils.IsExist)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	engine := EFCEngine{
		namespace: "fluid",
		Log:       fake.NullLogger(),
		name:      "efcdemo",
		Client:    client,
	}

	targetDataload := datav1alpha1.DataLoad{
		Spec: datav1alpha1.DataLoadSpec{
			Dataset: datav1alpha1.TargetDataset{
				Name:      "efcdemo",
				Namespace: "fluid",
			},
			Target: []datav1alpha1.TargetPath{
				{
					Path: "tmp",
				},
			},
		},
	}

	err := gohook.Hook(operations.EFCFileUtils.IsExist, mockNotExist, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	notExist, err := engine.CheckExistenceOfPath(targetDataload)
	if !(err != nil && notExist == true) {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhook()

	err = gohook.Hook(operations.EFCFileUtils.IsExist, mockExist, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	notExist, err = engine.CheckExistenceOfPath(targetDataload)
	if !(err == nil && notExist == false) {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhook()
}

func TestEFCEngine_CheckRuntimeReady(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		fields    fields
		sts       appsv1.StatefulSet
		podList   v1.PodList
		wantReady bool
	}{
		{
			name: "efc-test",
			fields: fields{
				name:      "efc-test",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc-test-worker",
					Namespace: "fluid",
					UID:       "uid1",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"a": "b"},
					},
				},
			},
			podList: v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fluid",
						Name:      "efc-test-worker-0",
						Labels:    map[string]string{"a": "b"},
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "efc-test-worker",
							UID:        "uid1",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						}},
					},
				}},
			},
			wantReady: true,
		},
		{
			name: "efc-test-err",
			fields: fields{
				name:      "efc-test-err",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc-test-err-worker",
					Namespace: "fluid",
					UID:       "uid2",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"a": "b"},
					},
				},
			},
			podList: v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fluid",
						Name:      "efc-test-err-worker-0",
						Labels:    map[string]string{"a": "b"},
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "efc-test-err-worker",
							UID:        "uid2",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionFalse,
						}},
					},
				}},
			},
			wantReady: false,
		},
		{
			name: "efc-test-err2",
			fields: fields{
				name:      "efc-test-err2",
				namespace: "fluid",
			},
			sts: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "efc-test-err2-worker",
					Namespace: "fluid",
					UID:       "uid3",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"a": "b"},
					},
				},
			},
			podList: v1.PodList{
				Items: []v1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fluid",
						Name:      "efc-test-err2-worker-0",
						Labels:    map[string]string{"nota": "notb"},
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "efc-test-err2-worker",
							UID:        "uid3",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{{
							Type:   v1.PodReady,
							Status: v1.ConditionTrue,
						}},
					},
				}},
			},
			wantReady: false,
		},
	}

	ReadyCommon := func(a operations.EFCFileUtils) (ready bool) {
		return true
	}
	wrappedUnhookReady := func() {
		err := gohook.UnHook(operations.EFCFileUtils.Ready)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(operations.EFCFileUtils.Ready, ReadyCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, tt := range tests {
		testObjs := []runtime.Object{}
		t.Run(tt.name, func(t *testing.T) {
			testObjs = append(testObjs, tt.sts.DeepCopy())
			testObjs = append(testObjs, tt.podList.DeepCopy())
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			e := &EFCEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}
			if gotReady := e.CheckRuntimeReady(); gotReady != tt.wantReady {
				t.Errorf("CheckRuntimeReady() = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}

	wrappedUnhookReady()
}
