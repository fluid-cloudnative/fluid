/*
Copyright 2021 The Fluid Authors.

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

package base

import (
	"reflect"
	"testing"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"

	fakeutils "github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_convertToTieredstoreInfo(t *testing.T) {
	type args struct {
		tieredstore v1alpha1.TieredStore
	}

	quota20Gi := resource.MustParse("20Gi")
	quota10Gi := resource.MustParse("10Gi")

	tests := []struct {
		name    string
		args    args
		want    TieredStoreInfo
		wantErr bool
	}{
		{
			name: "Test: No quota config set err",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Quota:     nil,
					QuotaList: "",
				},
			}}},
			want:    TieredStoreInfo{},
			wantErr: true,
		},
		{
			name: "Test: Inconsistent length of quotas and paths",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Path:      "/path/to/cache1/,/path/to/cache2",
					QuotaList: "10Gi,20Gi,30Gi",
				},
			}}},
			want:    TieredStoreInfo{},
			wantErr: true,
		},
		{
			name: "Test: Only quota is set, divide quota equally",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Path:  "/path/to/cache1/,/path/to/cache2",
					Quota: resource.NewQuantity(1024, resource.BinarySI),
				},
			}}},
			want: TieredStoreInfo{Levels: []Level{
				{
					CachePaths: []CachePath{
						{
							Path:  "/path/to/cache1",
							Quota: resource.NewQuantity(512, resource.BinarySI),
						},
						{
							Path:  "/path/to/cache2",
							Quota: resource.NewQuantity(512, resource.BinarySI),
						},
					},
				},
			}},
			wantErr: false,
		},
		{
			name: "Test: quotaList for configs",
			args: args{tieredstore: v1alpha1.TieredStore{Levels: []v1alpha1.Level{
				{
					Path: "/path/to/cache1/,/path/to/cache2/",
					// QuotaList Overwrites Quota
					Quota:     resource.NewQuantity(124, resource.BinarySI),
					QuotaList: "10Gi,20Gi",
				},
			}}},
			want: TieredStoreInfo{Levels: []Level{
				{
					CachePaths: []CachePath{
						{
							Path:  "/path/to/cache1",
							Quota: &quota10Gi,
						},
						{
							Path:  "/path/to/cache2",
							Quota: &quota20Gi,
						},
					},
				},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToTieredstoreInfo(tt.args.tieredstore)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToTieredstoreInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToTieredstoreInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildRuntimeInfo(t *testing.T) {
	const runtimetype = "alluxio"
	type args struct {
		name        string
		namespace   string
		runtimeType string
		tieredstore v1alpha1.TieredStore
	}

	tieredstore := v1alpha1.TieredStore{
		Levels: []v1alpha1.Level{
			{
				MediumType: "MEM",
				Path:       "/dev/shm/cache/cache1,/dev/shm/cache/cache2/",
				Quota:      nil,
				QuotaList:  "3Gi,2Gi",
				High:       "0.95",
				Low:        "0.7",
			},
			{
				MediumType: "SSD",
				Path:       "/mnt/cache",
				// 1 << 29 == 1Gi
				Quota:     resource.NewQuantity(1024*1024*1024, resource.BinarySI),
				QuotaList: "",
				High:      "0.99",
				Low:       "0.8",
			},
		},
	}

	tests := []struct {
		name        string
		args        args
		wantRuntime RuntimeInfoInterface
		wantErr     bool
	}{
		{
			name: "TestBuildRuntimeInfo",
			args: args{
				name:        "dataset",
				namespace:   "default",
				runtimeType: runtimetype,
				tieredstore: tieredstore,
			},
			wantRuntime: &RuntimeInfo{
				name:        "dataset",
				namespace:   "default",
				runtimeType: runtimetype,
				exclusive:   false,
				//setup:       false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRuntime, err := BuildRuntimeInfo(tt.args.name, tt.args.namespace, tt.args.runtimeType, WithTieredStore(tt.args.tieredstore))
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tieredstoreInfo, err := convertToTieredstoreInfo(tieredstore)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToTieredstoreInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotRuntime.GetName() != tt.wantRuntime.GetName() ||
				gotRuntime.GetNamespace() != tt.wantRuntime.GetNamespace() ||
				gotRuntime.GetRuntimeType() != tt.wantRuntime.GetRuntimeType() ||
				!reflect.DeepEqual(gotRuntime.GetTieredStoreInfo(), tieredstoreInfo) {
				t.Errorf("BuildRuntimeInfo() gotRuntime = %v, want %v", gotRuntime, tt.wantRuntime)
			}

		})
	}
}

func TestCleanPolicyAndLaunchMode(t *testing.T) {
	s := runtime.NewScheme()

	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.AlluxioRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JindoRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JuiceFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.GooseFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.Dataset{})

	// Test Alluxio Runtime
	alluxioRuntimeDefaultCleanPolicy := v1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_alluxio",
			Namespace: "default",
		},
		Spec: v1alpha1.AlluxioRuntimeSpec{
			Fuse: v1alpha1.AlluxioFuseSpec{},
		},
	}

	dataAlluxioDefaultCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_alluxio",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "default_policy_alluxio",
					Namespace: "default",
					Type:      common.AlluxioRuntime,
				},
			},
		},
	}

	alluxioRuntimeOnDemandCleanPolicy := v1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_alluxio",
			Namespace: "default",
		},
		Spec: v1alpha1.AlluxioRuntimeSpec{
			Fuse: v1alpha1.AlluxioFuseSpec{
				CleanPolicy: v1alpha1.OnDemandCleanPolicy,
				LaunchMode:  v1alpha1.LazyMode,
			},
		},
	}

	dataAlluxioOnDemandCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_alluxio",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_demand_policy_alluxio",
					Namespace: "default",
					Type:      common.AlluxioRuntime,
				},
			},
		},
	}

	alluxioRuntimeOnRuntimeDeletedCleanPolicy := v1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_alluxio",
			Namespace: "default",
		},
		Spec: v1alpha1.AlluxioRuntimeSpec{
			Fuse: v1alpha1.AlluxioFuseSpec{
				CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
				LaunchMode:  v1alpha1.EagerMode,
			},
		},
	}

	dataAlluxioOnRuntimeDeletedCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_alluxio",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_runtime_deleted_policy_alluxio",
					Namespace: "default",
					Type:      common.AlluxioRuntime,
				},
			},
		},
	}

	alluxioRuntimeObjs := []runtime.Object{}
	alluxioRuntimeObjs = append(alluxioRuntimeObjs, &alluxioRuntimeDefaultCleanPolicy, &dataAlluxioDefaultCleanPolicy)
	alluxioRuntimeObjs = append(alluxioRuntimeObjs, &alluxioRuntimeOnDemandCleanPolicy, &dataAlluxioOnDemandCleanPolicy)
	alluxioRuntimeObjs = append(alluxioRuntimeObjs, &alluxioRuntimeOnRuntimeDeletedCleanPolicy, &dataAlluxioOnRuntimeDeletedCleanPolicy)

	// Test JindoFs Runtime
	jindoRuntimeDefaultCleanPolicy := v1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_jindo",
			Namespace: "default",
		},
		Spec: v1alpha1.JindoRuntimeSpec{
			Fuse: v1alpha1.JindoFuseSpec{},
		},
	}

	dataJindoDefaultCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_jindo",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "default_policy_jindo",
					Namespace: "default",
					Type:      common.JindoRuntime,
				},
			},
		},
	}

	jindoRuntimeOnDemandCleanPolicy := v1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_jindo",
			Namespace: "default",
		},
		Spec: v1alpha1.JindoRuntimeSpec{
			Fuse: v1alpha1.JindoFuseSpec{
				CleanPolicy: v1alpha1.OnDemandCleanPolicy,
				LaunchMode:  v1alpha1.LazyMode,
			},
		},
	}

	dataJindoOnDemandCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_jindo",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_demand_policy_jindo",
					Namespace: "default",
					Type:      common.JindoRuntime,
				},
			},
		},
	}

	jindoRuntimeOnRuntimeDeletedCleanPolicy := v1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_jindo",
			Namespace: "default",
		},
		Spec: v1alpha1.JindoRuntimeSpec{
			Fuse: v1alpha1.JindoFuseSpec{
				CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
				LaunchMode:  v1alpha1.EagerMode,
			},
		},
	}

	dataJindoOnRuntimeDeletedCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_jindo",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_runtime_deleted_policy_jindo",
					Namespace: "default",
					Type:      common.JindoRuntime,
				},
			},
		},
	}

	jindoRuntimeObjs := []runtime.Object{}
	jindoRuntimeObjs = append(jindoRuntimeObjs, &jindoRuntimeDefaultCleanPolicy, &dataJindoDefaultCleanPolicy)
	jindoRuntimeObjs = append(jindoRuntimeObjs, &jindoRuntimeOnDemandCleanPolicy, &dataJindoOnDemandCleanPolicy)
	jindoRuntimeObjs = append(jindoRuntimeObjs, &jindoRuntimeOnRuntimeDeletedCleanPolicy, &dataJindoOnRuntimeDeletedCleanPolicy)

	// Test JuiceFs Runtime
	juiceRuntimeDefaultCleanPolicy := v1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_juicefs",
			Namespace: "default",
		},
		Spec: v1alpha1.JuiceFSRuntimeSpec{
			Fuse: v1alpha1.JuiceFSFuseSpec{},
		},
	}

	dataJuiceDefaultCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_juicefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "default_policy_juicefs",
					Namespace: "default",
					Type:      common.JuiceFSRuntime,
				},
			},
		},
	}

	juiceRuntimeOnDemandCleanPolicy := v1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_juicefs",
			Namespace: "default",
		},
		Spec: v1alpha1.JuiceFSRuntimeSpec{
			Fuse: v1alpha1.JuiceFSFuseSpec{
				CleanPolicy: v1alpha1.OnDemandCleanPolicy,
				LaunchMode:  v1alpha1.LazyMode,
			},
		},
	}

	dataJuiceOnDemandCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_juicefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_demand_policy_juicefs",
					Namespace: "default",
					Type:      common.JuiceFSRuntime,
				},
			},
		},
	}

	juiceRuntimeOnRuntimeDeletedCleanPolicy := v1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_juicefs",
			Namespace: "default",
		},
		Spec: v1alpha1.JuiceFSRuntimeSpec{
			Fuse: v1alpha1.JuiceFSFuseSpec{
				CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
				LaunchMode:  v1alpha1.EagerMode,
			},
		},
	}

	dataJuiceOnRuntimeDeletedCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_juicefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_runtime_deleted_policy_juicefs",
					Namespace: "default",
					Type:      common.JuiceFSRuntime,
				},
			},
		},
	}

	juiceRuntimeObjs := []runtime.Object{}
	juiceRuntimeObjs = append(juiceRuntimeObjs, &juiceRuntimeDefaultCleanPolicy, &dataJuiceDefaultCleanPolicy)
	juiceRuntimeObjs = append(juiceRuntimeObjs, &juiceRuntimeOnDemandCleanPolicy, &dataJuiceOnDemandCleanPolicy)
	juiceRuntimeObjs = append(juiceRuntimeObjs, &juiceRuntimeOnRuntimeDeletedCleanPolicy, &dataJuiceOnRuntimeDeletedCleanPolicy)

	// Test GooseFs Runtime
	goosefsRuntimeDefaultCleanPolicy := v1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_goosefs",
			Namespace: "default",
		},
		Spec: v1alpha1.GooseFSRuntimeSpec{
			Fuse: v1alpha1.GooseFSFuseSpec{},
		},
	}

	dataGooseFSDefaultCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default_policy_goosefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "default_policy_goosefs",
					Namespace: "default",
					Type:      common.GooseFSRuntime,
				},
			},
		},
	}

	goosefsRuntimeOnDemandCleanPolicy := v1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_goosefs",
			Namespace: "default",
		},
		Spec: v1alpha1.GooseFSRuntimeSpec{
			Fuse: v1alpha1.GooseFSFuseSpec{
				CleanPolicy: v1alpha1.OnDemandCleanPolicy,
				LaunchMode:  v1alpha1.LazyMode,
			},
		},
	}

	dataGooseFSOnDemandCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_demand_policy_goosefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_demand_policy_goosefs",
					Namespace: "default",
					Type:      common.GooseFSRuntime,
				},
			},
		},
	}

	goosefsRuntimeOnRuntimeDeletedCleanPolicy := v1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_goosefs",
			Namespace: "default",
		},
		Spec: v1alpha1.GooseFSRuntimeSpec{
			Fuse: v1alpha1.GooseFSFuseSpec{
				CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
				LaunchMode:  v1alpha1.EagerMode,
			},
		},
	}

	dataGooseFSOnRuntimeDeletedCleanPolicy := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "on_runtime_deleted_policy_goosefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "on_runtime_deleted_policy_goosefs",
					Namespace: "default",
					Type:      common.GooseFSRuntime,
				},
			},
		},
	}

	goosefsRuntimeObjs := []runtime.Object{}
	goosefsRuntimeObjs = append(goosefsRuntimeObjs, &goosefsRuntimeDefaultCleanPolicy, &dataGooseFSDefaultCleanPolicy)
	goosefsRuntimeObjs = append(goosefsRuntimeObjs, &goosefsRuntimeOnDemandCleanPolicy, &dataGooseFSOnDemandCleanPolicy)
	goosefsRuntimeObjs = append(goosefsRuntimeObjs, &goosefsRuntimeOnRuntimeDeletedCleanPolicy, &dataGooseFSOnRuntimeDeletedCleanPolicy)

	type args struct {
		client    client.Client
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		want    RuntimeInfoInterface
		wantErr bool
	}{
		{
			name: "default_test_alluxio",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:      "default_policy_alluxio",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "default_policy_alluxio",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_demand_test_alluxio",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:      "on_demand_policy_alluxio",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_demand_policy_alluxio",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_runtime_deleted_test-alluxio",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:      "on_runtime_deleted_policy_alluxio",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_runtime_deleted_policy_alluxio",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.EagerMode,
				},
			},
			wantErr: false,
		},
		{
			name: "default_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:      "default_policy_jindo",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "default_policy_jindo",
				namespace:   "default",
				runtimeType: common.JindoRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_demand_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:      "on_demand_policy_jindo",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_demand_policy_jindo",
				namespace:   "default",
				runtimeType: common.JindoRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_runtime_deleted_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:      "on_runtime_deleted_policy_jindo",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_runtime_deleted_policy_jindo",
				namespace:   "default",
				runtimeType: common.JindoRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.EagerMode,
				},
			},
			wantErr: false,
		},
		{
			name: "default_test-juicefs",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, juiceRuntimeObjs...),
				name:      "default_policy_juicefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "default_policy_juicefs",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_demand_test-juicefs",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, juiceRuntimeObjs...),
				name:      "on_demand_policy_juicefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_demand_policy_juicefs",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_runtime_deleted_test-juicefs",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, juiceRuntimeObjs...),
				name:      "on_runtime_deleted_policy_juicefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_runtime_deleted_policy_juicefs",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.EagerMode,
				},
			},
			wantErr: false,
		},
		{
			name: "default_test_goosefs",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:      "default_policy_goosefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "default_policy_goosefs",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_demand_test_goosefs",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:      "on_demand_policy_goosefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_demand_policy_goosefs",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "on_runtime_deleted_test-goosefs",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:      "on_runtime_deleted_policy_goosefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "on_runtime_deleted_policy_goosefs",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.EagerMode,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SetupFuseCleanPolicy will be called in GetRuntimeInfo()
			got, err := GetRuntimeInfo(tt.args.client, tt.args.name, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got.GetFuseCleanPolicy(), tt.want.GetFuseCleanPolicy()) {
				t.Errorf("GetRuntimeInfo() = %#v, want %#v", got, tt.want)
			}
			if !tt.wantErr && !reflect.DeepEqual(got.GetFuseLaunchMode(), tt.want.GetFuseLaunchMode()) {
				t.Errorf("GetFuseLaunchMode() = %#v, want %#v", got.GetFuseLaunchMode(), tt.want.GetFuseLaunchMode())
			}
		})
	}
}

func TestGetRuntimeInfo(t *testing.T) {
	s := runtime.NewScheme()

	alluxioRuntime := v1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: "default",
		},
	}

	dataAlluxio := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "alluxio",
					Namespace: "default",
					Type:      common.AlluxioRuntime,
				},
			},
		},
	}

	goosefsRuntime := v1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "goosefs",
			Namespace: "default",
		},
	}

	dataGooseFS := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "goosefs",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "goosefs",
					Namespace: "default",
					Type:      common.GooseFSRuntime,
				},
			},
		},
	}

	jindoRuntime := v1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jindo",
			Namespace: "default",
		},
	}

	dataJindo := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jindo",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "jindo",
					Namespace: "default",
					Type:      common.JindoRuntime,
				},
			},
		},
	}

	juicefsRuntime := v1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "juice",
			Namespace: "default",
		},
	}
	dataJuice := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "juice",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "juice",
					Namespace: "default",
					Type:      common.JuiceFSRuntime,
				},
			},
		},
	}

	efcRuntime := v1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efc",
			Namespace: "default",
		},
	}
	dataEFC := v1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efc",
			Namespace: "default",
		},
		Status: v1alpha1.DatasetStatus{
			Runtimes: []v1alpha1.Runtime{
				{
					Name:      "efc",
					Namespace: "default",
					Type:      common.EFCRuntime,
				},
			},
		},
	}
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.AlluxioRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.GooseFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JindoRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JuiceFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.EFCRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.Dataset{})
	_ = v1.AddToScheme(s)
	alluxioRuntimeObjs := []runtime.Object{}
	goosefsRuntimeObjs := []runtime.Object{}
	jindoRuntimeObjs := []runtime.Object{}
	juicefsRuntimeObjs := []runtime.Object{}
	efcRuntimeObjs := []runtime.Object{}

	alluxioRuntimeObjs = append(alluxioRuntimeObjs, &alluxioRuntime, &dataAlluxio)
	goosefsRuntimeObjs = append(goosefsRuntimeObjs, &goosefsRuntime, &dataGooseFS)
	jindoRuntimeObjs = append(jindoRuntimeObjs, &jindoRuntime, &dataJindo)
	juicefsRuntimeObjs = append(juicefsRuntimeObjs, &juicefsRuntime, &dataJuice)
	efcRuntimeObjs = append(efcRuntimeObjs, &efcRuntime, &dataEFC)
	type args struct {
		client    client.Client
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		args    args
		want    RuntimeInfoInterface
		wantErr bool
	}{
		{
			name: "alluxio_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:      "alluxio",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "alluxio",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "goosefs_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:      "goosefs",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "goosefs",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "goosefs_test_fake",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:      "goosefs-fake",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "goosefs-fake",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: true,
		},
		{
			name: "jindo_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:      "jindo",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "jindo",
				namespace:   "default",
				runtimeType: common.JindoRuntime,
				fuse: Fuse{
					CleanPolicy:         v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:          v1alpha1.LazyMode,
					MetricsScrapeTarget: mountModeSelector{},
				},
			},
			wantErr: false,
		},
		{
			name: "juicefs_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, juicefsRuntimeObjs...),
				name:      "juice",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "juice",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "juicefs_test_err",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, juicefsRuntimeObjs...),
				name:      "juice-fake",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "juice-fake",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: true,
		},
		{
			name: "efc_test",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, efcRuntimeObjs...),
				name:      "efc",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "efc",
				namespace:   "default",
				runtimeType: common.EFCRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnRuntimeDeletedCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: false,
		},
		{
			name: "efc_test_err",
			args: args{
				client:    fakeutils.NewFakeClientWithScheme(s, efcRuntimeObjs...),
				name:      "efc-fake",
				namespace: "default",
			},
			want: &RuntimeInfo{
				name:        "efc-fake",
				namespace:   "default",
				runtimeType: common.EFCRuntime,
				fuse: Fuse{
					CleanPolicy: v1alpha1.OnDemandCleanPolicy,
					LaunchMode:  v1alpha1.LazyMode,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRuntimeInfo(tt.args.client, tt.args.name, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				got.SetClient(nil)
			}

			if tt.want != nil {
				tt.want.SetClient(nil)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRuntimeInfo() = %#v\n, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetRuntimeStatus(t *testing.T) {
	s := runtime.NewScheme()

	alluxioRuntime := v1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio",
			Namespace: "default",
		},
	}

	goosefsRuntime := v1alpha1.GooseFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "goosefs",
			Namespace: "default",
		},
	}

	jindoRuntime := v1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jindo",
			Namespace: "default",
		},
	}

	juicefsRuntime := v1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "juice",
			Namespace: "default",
		},
	}

	efcRuntime := v1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "efc",
			Namespace: "default",
		},
	}

	thinRuntime := v1alpha1.ThinRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "thin",
			Namespace: "default",
		},
	}

	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.AlluxioRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.GooseFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JindoRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.JuiceFSRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.EFCRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ThinRuntime{})
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.Dataset{})

	_ = v1.AddToScheme(s)
	alluxioRuntimeObjs := []runtime.Object{}
	goosefsRuntimeObjs := []runtime.Object{}
	jindoRuntimeObjs := []runtime.Object{}
	juicefsRuntimeObjs := []runtime.Object{}
	efcRuntimeObjs := []runtime.Object{}
	thinRuntimeObjs := []runtime.Object{}

	alluxioRuntimeObjs = append(alluxioRuntimeObjs, &alluxioRuntime)
	goosefsRuntimeObjs = append(goosefsRuntimeObjs, &goosefsRuntime)
	jindoRuntimeObjs = append(jindoRuntimeObjs, &jindoRuntime)
	juicefsRuntimeObjs = append(juicefsRuntimeObjs, &juicefsRuntime)
	efcRuntimeObjs = append(efcRuntimeObjs, &efcRuntime)
	thinRuntimeObjs = append(thinRuntimeObjs, &thinRuntime)
	type args struct {
		client      client.Client
		name        string
		namespace   string
		runtimeType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "alluxio_test",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:        "alluxio",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},
			wantErr: false,
		},
		{
			name: "alluxio_test_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, alluxioRuntimeObjs...),
				name:        "alluxio-error",
				namespace:   "default",
				runtimeType: common.AlluxioRuntime,
			},
			wantErr: true,
		},
		{
			name: "goosefs_test",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:        "goosefs",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
			},
			wantErr: false,
		},
		{
			name: "goosefs_test_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, goosefsRuntimeObjs...),
				name:        "goosefs-error",
				namespace:   "default",
				runtimeType: common.GooseFSRuntime,
			},
			wantErr: true,
		},
		{
			name: "jindo_test",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:        "jindo",
				namespace:   "default",
				runtimeType: common.JindoRuntime,
			},
			wantErr: false,
		},
		{
			name: "jindo_test_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, jindoRuntimeObjs...),
				name:        "jindo-error",
				namespace:   "default",
				runtimeType: common.JindoRuntime,
			},
			wantErr: true,
		},
		{
			name: "juicefs_test",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, juicefsRuntimeObjs...),
				name:        "juice",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
			},
			wantErr: false,
		},
		{
			name: "juicefs_test_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, juicefsRuntimeObjs...),
				name:        "juice-error",
				namespace:   "default",
				runtimeType: common.JuiceFSRuntime,
			},
			wantErr: true,
		},
		{
			name: "efc_test",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, efcRuntimeObjs...),
				name:        "efc",
				namespace:   "default",
				runtimeType: common.EFCRuntime,
			},
			wantErr: false,
		},
		{
			name: "efc_test_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, efcRuntimeObjs...),
				name:        "efc-error",
				namespace:   "default",
				runtimeType: common.EFCRuntime,
			},
			wantErr: true,
		},
		{
			name: "thin_test",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, thinRuntimeObjs...),
				name:        "thin",
				namespace:   "default",
				runtimeType: common.ThinRuntime,
			},
			wantErr: false,
		},
		{
			name: "thin_test_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, thinRuntimeObjs...),
				name:        "thin-error",
				namespace:   "default",
				runtimeType: common.ThinRuntime,
			},
			wantErr: true,
		},
		{
			name: "default_error",
			args: args{
				client:      fakeutils.NewFakeClientWithScheme(s, thinRuntimeObjs...),
				name:        "thin-not-exit",
				namespace:   "default",
				runtimeType: "thin-not-exit",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetRuntimeStatus(tt.args.client, tt.args.runtimeType, tt.args.name, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetSyncRetryDuration(t *testing.T) {

	_, err := getSyncRetryDuration()
	if err != nil {
		t.Errorf("Failed to getSyncRetryDuration %v", err)
	}

	t.Setenv(syncRetryDurationEnv, "s")
	_, err = getSyncRetryDuration()
	if err == nil {
		t.Errorf("Expect to get err, but got nil")
	}

	t.Setenv(syncRetryDurationEnv, "3s")
	d, err := getSyncRetryDuration()
	if err != nil {
		t.Errorf("Failed to getSyncRetryDuration %v", err)
	}
	if d == nil {
		t.Errorf("Failed to set the duration, expect %v, got %v", time.Duration(3*time.Second), d)
	}
}

func TestPermitSync(t *testing.T) {

	id := "test id"
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      "hbase",
			Namespace: "fluid",
		},
		Log:     fakeutils.NullLogger(),
		Runtime: &v1alpha1.AlluxioRuntime{},
	}

	templateEngine := NewTemplateEngine(nil, id, ctx)
	permit := templateEngine.permitSync(types.NamespacedName{Namespace: ctx.Namespace, Name: ctx.Namespace})
	if !permit {
		t.Errorf("expect permit, but got %v", permit)
	}

	templateEngine.setTimeOfLastSync()
	permit = templateEngine.permitSync(types.NamespacedName{Namespace: ctx.Namespace, Name: ctx.Namespace})
	if permit {
		t.Errorf("expect not permit, but got %v", permit)
	}

	templateEngine.setTimeOfLastSync()
	templateEngine.syncRetryDuration = 1 * time.Microsecond
	time.Sleep(1 * time.Second)
	permit = templateEngine.permitSync(types.NamespacedName{Namespace: ctx.Namespace, Name: ctx.Namespace})
	if !permit {
		t.Errorf("expect permit, but got %v", permit)
	}
}
