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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestSyncMetadataInternal(t *testing.T) {
	mockTotalStorageBytesCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
	}
	mockTotalStorageBytesError := func(e *EFCEngine) (int64, error) {
		return 0, errors.New("other error")
	}
	wrappedUnhookTotalStorageBytes := func(e *EFCEngine) {
		err := gohook.UnHookMethod(e, "TotalStorageBytes")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	mockTotalFileNumsCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
	}
	mockTotalFileNumsError := func(e *EFCEngine) (int64, error) {
		return 0, errors.New("other error")
	}
	wrappedUnhookTotalFileNums := func(e *EFCEngine) {
		err := gohook.UnHookMethod(e, "TotalFileNums")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	testObjs := []runtime.Object{}
	EFCRuntimeInputs := []datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
	}
	for _, EFCRuntime := range EFCRuntimeInputs {
		testObjs = append(testObjs, EFCRuntime.DeepCopy())
	}

	var datasetInputs = []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &EFCEngine{
		name:      "spark",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	err := gohook.HookMethod(engine, "TotalStorageBytes", mockTotalStorageBytesError, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.syncMetadataInternal()
	if err == nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookTotalStorageBytes(engine)

	err = gohook.HookMethod(engine, "TotalStorageBytes", mockTotalStorageBytesCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.HookMethod(engine, "TotalFileNums", mockTotalFileNumsError, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.syncMetadataInternal()
	if err == nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookTotalStorageBytes(engine)
	wrappedUnhookTotalFileNums(engine)

	err = gohook.HookMethod(engine, "TotalStorageBytes", mockTotalStorageBytesCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.HookMethod(engine, "TotalFileNums", mockTotalFileNumsCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.syncMetadataInternal()
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookTotalStorageBytes(engine)
	wrappedUnhookTotalFileNums(engine)
}

func TestSyncMetadata(t *testing.T) {
	mockShouldCheckUFSCommon := func(e *EFCEngine) (should bool, err error) {
		return true, nil
	}
	wrappedUnhookShouldCheckUFS := func(e *EFCEngine) {
		err := gohook.UnHookMethod(e, "ShouldCheckUFS")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	mockTotalStorageBytesCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
	}
	wrappedUnhookTotalStorageBytes := func(e *EFCEngine) {
		err := gohook.UnHookMethod(e, "TotalStorageBytes")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	mockTotalFileNumsCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
	}
	wrappedUnhookTotalFileNums := func(e *EFCEngine) {
		err := gohook.UnHookMethod(e, "TotalFileNums")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	testObjs := []runtime.Object{}
	EFCRuntimeInputs := []datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
	}
	for _, EFCRuntime := range EFCRuntimeInputs {
		testObjs = append(testObjs, EFCRuntime.DeepCopy())
	}

	var datasetInputs = []datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, datasetInput := range datasetInputs {
		testObjs = append(testObjs, datasetInput.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engine := &EFCEngine{
		name:      "spark",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	err := gohook.HookMethod(engine, "ShouldCheckUFS", mockShouldCheckUFSCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.HookMethod(engine, "TotalStorageBytes", mockTotalStorageBytesCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.HookMethod(engine, "TotalFileNums", mockTotalFileNumsCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = engine.SyncMetadata()
	if err != nil {
		t.Errorf("fail to exec the function")
	}

	wrappedUnhookTotalFileNums(engine)
	wrappedUnhookTotalStorageBytes(engine)
	wrappedUnhookShouldCheckUFS(engine)
}
