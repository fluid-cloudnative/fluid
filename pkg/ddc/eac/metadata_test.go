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

package eac

import (
	"errors"
	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestSyncMetadataInternal(t *testing.T) {
	mockTotalStorageBytesCommon := func() (int64, error) {
		return 0, nil
	}
	mockTotalStorageBytesError := func() (int64, error) {
		return 0, errors.New("other error")
	}
	wrappedUnhookTotalStorageBytes := func(e *EACEngine) {
		err := gohook.UnHook(e.TotalStorageBytes)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	mockTotalFileNumsCommon := func() (int64, error) {
		return 0, nil
	}
	mockTotalFileNumsError := func() (int64, error) {
		return 0, errors.New("other error")
	}
	wrappedUnhookTotalFileNums := func(e *EACEngine) {
		err := gohook.UnHook(e.TotalFileNums)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	testObjs := []runtime.Object{}
	EACRuntimeInputs := []datav1alpha1.EACRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EACRuntimeSpec{
				Master: datav1alpha1.EACCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
	}
	for _, EACRuntime := range EACRuntimeInputs {
		testObjs = append(testObjs, EACRuntime.DeepCopy())
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

	engine := &EACEngine{
		name:      "spark",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	err := gohook.Hook(engine.TotalStorageBytes, mockTotalStorageBytesError, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.syncMetadataInternal()
	if err == nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookTotalStorageBytes(engine)

	err = gohook.Hook(engine.TotalStorageBytes, mockTotalStorageBytesCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(engine.TotalStorageBytes, mockTotalFileNumsError, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = engine.syncMetadataInternal()
	if err == nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookTotalStorageBytes(engine)
	wrappedUnhookTotalFileNums(engine)

	err = gohook.Hook(engine.TotalStorageBytes, mockTotalStorageBytesCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(engine.TotalStorageBytes, mockTotalFileNumsCommon, nil)
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
	mockShouldCheckUFSCommon := func() (should bool, err error) {
		return true, nil
	}
	wrappedUnhookShouldCheckUFS := func(e *EACEngine) {
		err := gohook.UnHook(e.ShouldCheckUFS)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	mockTotalStorageBytesCommon := func() (int64, error) {
		return 0, nil
	}
	wrappedUnhookTotalStorageBytes := func(e *EACEngine) {
		err := gohook.UnHook(e.TotalStorageBytes)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	mockTotalFileNumsCommon := func() (int64, error) {
		return 0, nil
	}
	wrappedUnhookTotalFileNums := func(e *EACEngine) {
		err := gohook.UnHook(e.TotalFileNums)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	testObjs := []runtime.Object{}
	EACRuntimeInputs := []datav1alpha1.EACRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EACRuntimeSpec{
				Master: datav1alpha1.EACCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
	}
	for _, EACRuntime := range EACRuntimeInputs {
		testObjs = append(testObjs, EACRuntime.DeepCopy())
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

	engine := &EACEngine{
		name:      "spark",
		namespace: "fluid",
		Client:    client,
		Log:       fake.NullLogger(),
	}

	err := gohook.Hook(engine.ShouldCheckUFS, mockShouldCheckUFSCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(engine.TotalStorageBytes, mockTotalStorageBytesCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(engine.TotalStorageBytes, mockTotalFileNumsCommon, nil)
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
