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

	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
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

	runtimeInfo, err := base.BuildRuntimeInfo("spark", "fluid", "efc")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	engine := &EFCEngine{
		name:        "spark",
		namespace:   "fluid",
		Client:      client,
		Log:         fake.NullLogger(),
		runtimeInfo: runtimeInfo,
	}

	err = gohook.HookMethod(engine, "TotalStorageBytes", mockTotalStorageBytesError, nil)
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
