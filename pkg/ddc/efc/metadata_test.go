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

	"github.com/agiledragon/gomonkey/v2"
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

	mockTotalFileNumsCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
	}
	mockTotalFileNumsError := func(e *EFCEngine) (int64, error) {
		return 0, errors.New("other error")
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

	patches := gomonkey.ApplyMethod(engine, "TotalStorageBytes", mockTotalStorageBytesError)
	defer patches.Reset()

	err = engine.syncMetadataInternal()
	if err == nil {
		t.Errorf("fail to exec the function")
	}

	patches.ApplyMethod(engine, "TotalStorageBytes", mockTotalStorageBytesCommon)
	patches.ApplyMethod(engine, "TotalFileNums", mockTotalFileNumsError)

	err = engine.syncMetadataInternal()
	if err == nil {
		t.Errorf("fail to exec the function")
	}

	patches.ApplyMethod(engine, "TotalStorageBytes", mockTotalStorageBytesCommon)
	patches.ApplyMethod(engine, "TotalFileNums", mockTotalFileNumsCommon)

	err = engine.syncMetadataInternal()
	if err != nil {
		t.Errorf("fail to exec the function")
	}
}

func TestSyncMetadata(t *testing.T) {
	mockShouldCheckUFSCommon := func(e *EFCEngine) (should bool, err error) {
		return true, nil
	}

	mockTotalStorageBytesCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
	}

	mockTotalFileNumsCommon := func(e *EFCEngine) (int64, error) {
		return 0, nil
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

	patches := gomonkey.ApplyMethod(engine, "ShouldCheckUFS", mockShouldCheckUFSCommon)
	patches.ApplyMethod(engine, "TotalStorageBytes", mockTotalStorageBytesCommon)
	patches.ApplyMethod(engine, "TotalFileNums", mockTotalFileNumsCommon)
	defer patches.Reset()

	err := engine.SyncMetadata()
	if err != nil {
		t.Errorf("fail to exec the function")
	}
}
