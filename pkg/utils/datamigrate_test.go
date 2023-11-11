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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestGetDataMigrate(t *testing.T) {
	mockDataMigrateName := "fluid-test-data-migrate"
	mockDataMigrateNamespace := "default"
	initDataMigrate := &datav1alpha1.DataMigrate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDataMigrateName,
			Namespace: mockDataMigrateNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, initDataMigrate)

	fakeClient := fake.NewFakeClientWithScheme(s, initDataMigrate)

	testCases := map[string]struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		"test get DataMigrate case 1": {
			name:      mockDataMigrateName,
			namespace: mockDataMigrateNamespace,
			wantName:  mockDataMigrateName,
			notFound:  false,
		},
		"test get DataMigrate case 2": {
			name:      mockDataMigrateName + "not-exist",
			namespace: mockDataMigrateNamespace,
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range testCases {
		gotDataMigrate, err := GetDataMigrate(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil && gotDataMigrate != nil {
				t.Errorf("%s check failure, want get err, but get nil", k)
			}
		} else {
			if gotDataMigrate.Name != item.wantName {
				t.Errorf("%s check failure, want DataMigrate name:%s, got DataMigrate name:%s", k, item.wantName, gotDataMigrate.Name)
			}
		}
	}
}

func TestGetDataMigrateJobName(t *testing.T) {
	type args struct {
		releaseName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				releaseName: "test",
			},
			want: "test-migrate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataMigrateJobName(tt.args.releaseName); got != tt.want {
				t.Errorf("GetDataMigrateJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDataMigrateReleaseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				name: "test",
			},
			want: "test-migrate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataMigrateReleaseName(tt.args.name); got != tt.want {
				t.Errorf("GetDataMigrateReleaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTargetDatasetOfMigrate(t *testing.T) {
	dataSet := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "fluid",
		},
		Status: datav1alpha1.DatasetStatus{
			Runtimes: []datav1alpha1.Runtime{
				{
					Name:      "juicefs-runtime",
					Namespace: "fluid",
					Type:      "juicefs",
					Category:  common.AccelerateCategory,
				},
			},
		},
	}
	juicefsRuntime := &datav1alpha1.JuiceFSRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dataset",
			Namespace: "fluid",
		},
	}
	runtimeObjs := []runtime.Object{}
	runtimeObjs = append(runtimeObjs, dataSet.DeepCopy(), juicefsRuntime.DeepCopy())
	fakeClient := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)

	type args struct {
		dataMigrate *datav1alpha1.DataMigrate
	}
	tests := []struct {
		name        string
		args        args
		wantDataset *datav1alpha1.Dataset
		wantErr     bool
	}{
		{
			name: "test-from",
			args: args{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						From: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name:      "test-dataset",
								Namespace: "fluid",
							},
						},
					},
				},
			},
			wantDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
			},
			wantErr: false,
		},
		{
			name: "test-to",
			args: args{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						To: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name:      "test-dataset",
								Namespace: "fluid",
							},
						},
					},
				},
			},
			wantDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
			},
			wantErr: false,
		},
		{
			name: "test-not-exist",
			args: args{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						To: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name:      "not-exist-dataset",
								Namespace: "fluid",
							},
						},
					},
				},
			},
			wantDataset: nil,
			wantErr:     true,
		},
		{
			name: "test-to-type",
			args: args{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						To: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name:      "test-dataset",
								Namespace: "fluid",
							},
						},
						RuntimeType: "juicefs",
					},
				},
			},
			wantDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "fluid",
				},
			},
			wantErr: false,
		},
		{
			name: "test-wrong-type",
			args: args{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						To: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name:      "test-dataset",
								Namespace: "fluid",
							},
						},
						RuntimeType: "not-exist-runtime",
					},
				},
			},
			wantDataset: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDataset, err := GetTargetDatasetOfMigrate(fakeClient, tt.args.dataMigrate)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetDatasetOfMigrate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDataset != nil || tt.wantDataset != nil {
				if gotDataset.Name != tt.wantDataset.Name || gotDataset.Namespace != tt.wantDataset.Namespace {
					t.Errorf("GetTargetDatasetOfMigrate() gotDataset = %v, want %v", gotDataset, tt.wantDataset)
				}
			}
		})
	}
}
