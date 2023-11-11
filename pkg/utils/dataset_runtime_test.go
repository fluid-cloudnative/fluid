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

package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
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
						Controller: utilpointer.BoolPtr(true),
						UID:        "3e108dcc-9aab-4d0b-99dc-9976d5cd6d5a",
					},
				},
			},
		}, {
			ObjectMeta: v1.ObjectMeta{
				Name:      "ThinRuntimeExistWithOwnerReference",
				Namespace: "default",
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateRuntimeForReferenceDatasetIfNotExist(fakeClient, tt.dataset); (err != nil) != tt.wantErr {
				t.Errorf("Testcase %v CreateRuntimeForReferenceDatasetIfNotExist() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
