/*
Copyright 2023 The Fluid Authors.

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

package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func TestGetRuntimeByCategory(t *testing.T) {
	testCases := map[string]struct {
		runtimes  []datav1alpha1.Runtime
		wantIndex int
	}{
		"test get runtime by category case 1": {
			runtimes:  mockThreeRuntimes(2, common.AccelerateCategory),
			wantIndex: 2,
		},
		"test get runtime by category case 2": {
			runtimes:  mockThreeRuntimes(0, common.AccelerateCategory),
			wantIndex: 0,
		},
		"test get runtime by category case 3": {
			runtimes:  mockThreeRuntimes(4, common.AccelerateCategory),
			wantIndex: -1,
		},
		"test get runtime by category case 4": {
			runtimes:  mockThreeRuntimes(1, common.AccelerateCategory),
			wantIndex: 1,
		},
		"test get runtime by category case 5": {
			runtimes:  nil,
			wantIndex: -1,
		},
	}

	for k, item := range testCases {
		gotIndex, _ := GetRuntimeByCategory(item.runtimes, common.AccelerateCategory)
		if gotIndex != item.wantIndex {
			t.Errorf("%s check failure, want index:%v,got index:%v", k, item.wantIndex, gotIndex)
		}

	}
}

func mockThreeRuntimes(index int, category common.Category) []datav1alpha1.Runtime {
	list := make([]datav1alpha1.Runtime, 0)

	r1 := datav1alpha1.Runtime{}
	list = append(list, r1)

	r2 := datav1alpha1.Runtime{}
	list = append(list, r2)

	r3 := datav1alpha1.Runtime{}
	list = append(list, r3)

	if index < len(list) && index >= 0 {
		list[index].Category = category
	}

	return list
}

func TestCreateRuntimeForReferenceDatasetIfNotExist(t *testing.T) {

	deletionTimestamp := v1.Now()
	thinRuntimes := []*datav1alpha1.ThinRuntime{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ThinRuntimeExists",
				Namespace: "default",
				OwnerReferences: []v1.OwnerReference{
					{
						// Kind:       "Dataset",
						// APIVersion: "data.fluid.io/v1alpha1",
						Name:       "ThinRuntimeExists",
						Controller: ptr.To(true),
						UID:        "3e108dcc-9aab-4d0b-99dc-9976d5cd6d5a",
					},
				},
			},
		}, {
			ObjectMeta: v1.ObjectMeta{
				Name:      "ThinRuntimeExistWithOwnerReference",
				Namespace: "default",
			},
		}, {
			// A leftover runtime which is stuck in Terminating, e.g. because its controller is
			// scaled to 0 and can not remove the finalizer. The finalizer is also required to keep
			// the object in the fake client's tracker once a deletionTimestamp is set.
			ObjectMeta: v1.ObjectMeta{
				Name:              "ThinRuntimeTerminating",
				Namespace:         "default",
				DeletionTimestamp: &deletionTimestamp,
				Finalizers:        []string{"thin-runtime-controller-finalizer"},
			},
		},
	}
	objs := []runtime.Object{}
	for _, thinRuntime := range thinRuntimes {
		objs = append(objs, thinRuntime.DeepCopy())
	}
	datasetScheme := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(datasetScheme)
	fakeClient := fake.NewFakeClientWithScheme(datasetScheme, objs...)

	tests := []struct {
		name    string
		dataset *datav1alpha1.Dataset
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "ThinRuntimeExists",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      "ThinRuntimeExists",
					Namespace: "default",
					UID:       "3e108dcc-9aab-4d0b-99dc-9976d5cd6d5a",
				},
			},
			wantErr: false,
		}, {
			name: "ThinRuntimeExistWithOwnerReference",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      "ThinRuntimeExistWithOwnerReference",
					Namespace: "default",
				},
			},
			wantErr: false,
		}, {
			name: "ThinRuntimeDoesnotExist",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      "ThinRuntimeDoesnotExist",
					Namespace: "default",
				},
			},
			wantErr: false,
		}, {
			// The runtime of the same name is still terminating, it can neither be adopted nor be
			// re-created, so an error is expected to make the caller requeue.
			name: "ThinRuntimeTerminating",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      "ThinRuntimeTerminating",
					Namespace: "default",
					UID:       "5b7bd2c9-e6e8-4c1e-9c9a-5d2b6ff6bd11",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateRuntimeForReferenceDatasetIfNotExist(fakeClient, tt.dataset); (err != nil) != tt.wantErr {
				t.Errorf("Testcase %v CreateRuntimeForReferenceDatasetIfNotExist() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}

	// The terminating runtime must be left untouched, especially it must not be adopted by the dataset.
	terminatingRuntime, err := GetThinRuntime(fakeClient, "ThinRuntimeTerminating", "default")
	if err != nil {
		t.Fatalf("failed to get the terminating thinRuntime: %v", err)
	}
	if !HasDeletionTimestamp(terminatingRuntime.ObjectMeta) {
		t.Errorf("expected the thinRuntime ThinRuntimeTerminating to be still terminating")
	}
	if len(terminatingRuntime.GetOwnerReferences()) != 0 {
		t.Errorf("expected no ownerReference set on the terminating thinRuntime, but got %v", terminatingRuntime.GetOwnerReferences())
	}
}

func TestDatasetControllerOwnerReference(t *testing.T) {
	testCases := map[string]struct {
		dataset            *datav1alpha1.Dataset
		expectedKind       string
		expectedAPIVersion string
	}{
		"dataset with complete TypeMeta": {
			dataset: &datav1alpha1.Dataset{
				TypeMeta: v1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "complete-dataset",
					UID:  "uid-complete",
				},
			},
			expectedKind:       "Dataset",
			expectedAPIVersion: "data.fluid.io/v1alpha1",
		},
		"dataset with empty TypeMeta": {
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name: "empty-typemeta-dataset",
					UID:  "uid-empty",
				},
			},
			expectedKind:       "Dataset",
			expectedAPIVersion: "data.fluid.io/v1alpha1",
		},
		"dataset with only Kind in TypeMeta": {
			dataset: &datav1alpha1.Dataset{
				TypeMeta: v1.TypeMeta{
					Kind: "Dataset",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "kind-only-dataset",
					UID:  "uid-kind",
				},
			},
			expectedKind:       "Dataset",
			expectedAPIVersion: "data.fluid.io/v1alpha1",
		},
		"dataset with only APIVersion in TypeMeta": {
			dataset: &datav1alpha1.Dataset{
				TypeMeta: v1.TypeMeta{
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "apiversion-only-dataset",
					UID:  "uid-apiversion",
				},
			},
			expectedKind:       "Dataset",
			expectedAPIVersion: "data.fluid.io/v1alpha1",
		},
		"dataset with custom APIVersion in TypeMeta": {
			dataset: &datav1alpha1.Dataset{
				TypeMeta: v1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1beta1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "custom-apiversion-dataset",
					UID:  "uid-custom",
				},
			},
			expectedKind:       "Dataset",
			expectedAPIVersion: "data.fluid.io/v1beta1",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ownerRef := datasetControllerOwnerReference(tc.dataset)
			if ownerRef.Kind != tc.expectedKind {
				t.Errorf("expected Kind %s, got %s", tc.expectedKind, ownerRef.Kind)
			}
			if ownerRef.APIVersion != tc.expectedAPIVersion {
				t.Errorf("expected APIVersion %s, got %s", tc.expectedAPIVersion, ownerRef.APIVersion)
			}
			if ownerRef.Name != tc.dataset.GetName() {
				t.Errorf("expected Name %s, got %s", tc.dataset.GetName(), ownerRef.Name)
			}
			if ownerRef.UID != tc.dataset.GetUID() {
				t.Errorf("expected UID %s, got %s", tc.dataset.GetUID(), ownerRef.UID)
			}
			if ownerRef.Controller == nil || !*ownerRef.Controller {
				t.Errorf("expected Controller to be true, got %v", ownerRef.Controller)
			}
		})
	}
}
