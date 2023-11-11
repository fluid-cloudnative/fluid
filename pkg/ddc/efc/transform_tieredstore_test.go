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
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTransformWorkerTieredStore(t *testing.T) {
	test2_quota := resource.MustParse("2Gi")
	var tests = []struct {
		runtime        *datav1alpha1.EFCRuntime
		dataset        *datav1alpha1.Dataset
		efcValue       *EFC
		wantType       string
		wantPath       string
		wantMediumType string
		wantQuota      string
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{},
			},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			efcValue:       &EFC{},
			wantType:       string(common.VolumeTypeEmptyDir),
			wantPath:       "/cache_dir//test",
			wantMediumType: string(common.Memory),
			wantQuota:      "1GB",
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tes2",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{
					TieredStore: datav1alpha1.TieredStore{
						Levels: []datav1alpha1.Level{
							{
								MediumType: "MEM",
								VolumeType: "emptyDir",
								Path:       "/cache_dir2",
								Quota:      &test2_quota,
							},
						},
					},
				},
			},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tes2",
				},
			},
			efcValue:       &EFC{},
			wantType:       string(common.VolumeTypeEmptyDir),
			wantPath:       "/cache_dir2//tes2",
			wantMediumType: string(common.Memory),
			wantQuota:      "2GB",
		},
	}
	for _, test := range tests {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, test.runtime.DeepCopy())
		testObjs = append(testObjs, test.dataset.DeepCopy())
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
		engine := &EFCEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformWorkerTieredStore(test.runtime, test.efcValue)
		if err != nil {
			t.Error(err)
		}
		if test.wantType != test.efcValue.Worker.TieredStore.Levels[0].Type {
			t.Errorf("expected %v, got %v", test.wantType, test.efcValue.Worker.TieredStore.Levels[0].Type)
		}
		if test.wantPath != test.efcValue.Worker.TieredStore.Levels[0].Path {
			t.Errorf("expected %v, got %v", test.wantPath, test.efcValue.Worker.TieredStore.Levels[0].Path)
		}
		if test.wantMediumType != test.efcValue.getTiredStoreLevel0MediumType() {
			t.Errorf("expected %v, got %v", test.wantMediumType, test.efcValue.getTiredStoreLevel0MediumType())
		}
		if test.wantQuota != test.efcValue.getTiredStoreLevel0Quota() {
			t.Errorf("expected %v, got %v", test.wantQuota, test.efcValue.getTiredStoreLevel0Quota())
		}
	}
}

func TestTransformMasterTieredStore(t *testing.T) {
	var tests = []struct {
		runtime        *datav1alpha1.EFCRuntime
		dataset        *datav1alpha1.Dataset
		efcValue       *EFC
		wantType       string
		wantPath       string
		wantMediumType string
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{},
			},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			efcValue:       &EFC{},
			wantType:       string(common.VolumeTypeEmptyDir),
			wantPath:       "/dev/shm",
			wantMediumType: string(common.Memory),
		},
	}
	for _, test := range tests {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, test.runtime.DeepCopy())
		testObjs = append(testObjs, test.dataset.DeepCopy())
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
		engine := &EFCEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformMasterTieredStore(test.runtime, test.efcValue)
		if err != nil {
			t.Error(err)
		}
		if test.wantType != test.efcValue.Master.TieredStore.Levels[0].Type {
			t.Errorf("expected %v, got %v", test.wantType, test.efcValue.Master.TieredStore.Levels[0].Type)
		}
		if test.wantPath != test.efcValue.Master.TieredStore.Levels[0].Path {
			t.Errorf("expected %v, got %v", test.wantPath, test.efcValue.Master.TieredStore.Levels[0].Path)
		}
		if test.wantMediumType != test.efcValue.Master.TieredStore.Levels[0].MediumType {
			t.Errorf("expected %v, got %v", test.wantMediumType, test.efcValue.Master.TieredStore.Levels[0].MediumType)
		}
	}
}

func TestTransformFuseTieredStore(t *testing.T) {
	var tests = []struct {
		runtime        *datav1alpha1.EFCRuntime
		dataset        *datav1alpha1.Dataset
		efcValue       *EFC
		wantType       string
		wantPath       string
		wantMediumType string
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: datav1alpha1.EFCRuntimeSpec{},
			},
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			efcValue:       &EFC{},
			wantType:       string(common.VolumeTypeEmptyDir),
			wantPath:       "/dev/shm",
			wantMediumType: string(common.Memory),
		},
	}
	for _, test := range tests {
		testObjs := []runtime.Object{}
		testObjs = append(testObjs, test.runtime.DeepCopy())
		testObjs = append(testObjs, test.dataset.DeepCopy())
		client := fake.NewFakeClientWithScheme(testScheme, testObjs...)
		engine := &EFCEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformFuseTieredStore(test.runtime, test.efcValue)
		if err != nil {
			t.Error(err)
		}
		if test.wantType != test.efcValue.Fuse.TieredStore.Levels[0].Type {
			t.Errorf("expected %v, got %v", test.wantType, test.efcValue.Fuse.TieredStore.Levels[0].Type)
		}
		if test.wantPath != test.efcValue.Fuse.TieredStore.Levels[0].Path {
			t.Errorf("expected %v, got %v", test.wantPath, test.efcValue.Fuse.TieredStore.Levels[0].Path)
		}
		if test.wantMediumType != test.efcValue.Fuse.TieredStore.Levels[0].MediumType {
			t.Errorf("expected %v, got %v", test.wantMediumType, test.efcValue.Fuse.TieredStore.Levels[0].MediumType)
		}
	}
}
