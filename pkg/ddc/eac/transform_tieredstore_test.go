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
		eacValue       *EAC
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
			eacValue:       &EAC{},
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
			eacValue:       &EAC{},
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
		engine := &EACEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformWorkerTieredStore(test.runtime, test.eacValue)
		if err != nil {
			t.Error(err)
		}
		if test.wantType != test.eacValue.Worker.TieredStore.Levels[0].Type {
			t.Errorf("expected %v, got %v", test.wantType, test.eacValue.Worker.TieredStore.Levels[0].Type)
		}
		if test.wantPath != test.eacValue.Worker.TieredStore.Levels[0].Path {
			t.Errorf("expected %v, got %v", test.wantPath, test.eacValue.Worker.TieredStore.Levels[0].Path)
		}
		if test.wantMediumType != test.eacValue.getTiredStoreLevel0MediumType() {
			t.Errorf("expected %v, got %v", test.wantMediumType, test.eacValue.getTiredStoreLevel0MediumType())
		}
		if test.wantQuota != test.eacValue.getTiredStoreLevel0Quota() {
			t.Errorf("expected %v, got %v", test.wantQuota, test.eacValue.getTiredStoreLevel0Quota())
		}
	}
}

func TestTransformMasterTieredStore(t *testing.T) {
	var tests = []struct {
		runtime        *datav1alpha1.EFCRuntime
		dataset        *datav1alpha1.Dataset
		eacValue       *EAC
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
			eacValue:       &EAC{},
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
		engine := &EACEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformMasterTieredStore(test.runtime, test.eacValue)
		if err != nil {
			t.Error(err)
		}
		if test.wantType != test.eacValue.Master.TieredStore.Levels[0].Type {
			t.Errorf("expected %v, got %v", test.wantType, test.eacValue.Master.TieredStore.Levels[0].Type)
		}
		if test.wantPath != test.eacValue.Master.TieredStore.Levels[0].Path {
			t.Errorf("expected %v, got %v", test.wantPath, test.eacValue.Master.TieredStore.Levels[0].Path)
		}
		if test.wantMediumType != test.eacValue.Master.TieredStore.Levels[0].MediumType {
			t.Errorf("expected %v, got %v", test.wantMediumType, test.eacValue.Master.TieredStore.Levels[0].MediumType)
		}
	}
}

func TestTransformFuseTieredStore(t *testing.T) {
	var tests = []struct {
		runtime        *datav1alpha1.EFCRuntime
		dataset        *datav1alpha1.Dataset
		eacValue       *EAC
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
			eacValue:       &EAC{},
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
		engine := &EACEngine{
			Log:    fake.NullLogger(),
			Client: client,
			name:   test.runtime.Name,
		}
		err := engine.transformFuseTieredStore(test.runtime, test.eacValue)
		if err != nil {
			t.Error(err)
		}
		if test.wantType != test.eacValue.Fuse.TieredStore.Levels[0].Type {
			t.Errorf("expected %v, got %v", test.wantType, test.eacValue.Fuse.TieredStore.Levels[0].Type)
		}
		if test.wantPath != test.eacValue.Fuse.TieredStore.Levels[0].Path {
			t.Errorf("expected %v, got %v", test.wantPath, test.eacValue.Fuse.TieredStore.Levels[0].Path)
		}
		if test.wantMediumType != test.eacValue.Fuse.TieredStore.Levels[0].MediumType {
			t.Errorf("expected %v, got %v", test.wantMediumType, test.eacValue.Fuse.TieredStore.Levels[0].MediumType)
		}
	}
}
