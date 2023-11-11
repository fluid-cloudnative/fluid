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

package base

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPhysicalDatasetFromMounts(t *testing.T) {
	tests := []struct {
		virtualDataset *datav1alpha1.Dataset
		want           int
	}{
		{
			virtualDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
						{
							MountPoint: "dataset://ns-b/n-b",
						},
					},
				},
			},
			want: 2,
		},
		{
			virtualDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
						{
							MountPoint: "http://ns-b/n-b",
						},
					},
				},
			},
			want: 1,
		},
		{
			virtualDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		if got := GetPhysicalDatasetFromMounts(tt.virtualDataset.Spec.Mounts); len(got) != tt.want {
			t.Errorf("GetPhysicalDatasetFromMounts() len = %v, want %v", got, tt.want)
		}
	}
}

func TestGetDatasetRefName(t *testing.T) {
	refNameA := GetDatasetRefName("a-b", "c")
	refNameB := GetDatasetRefName("a", "bc")

	if refNameB == refNameA {
		t.Errorf("RefName is equal for different name and namespace")
	}
}

func TestCheckReferenceDataset(t *testing.T) {

	tests := []struct {
		name      string
		dataset   *datav1alpha1.Dataset
		wantCheck bool
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "dataset_with_two_datasetmounts",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
						{
							MountPoint: "dataset://ns-b/n-b",
						},
					},
				},
			},
			wantCheck: false,
			wantErr:   true,
		},
		{
			name: "dataset_with_two_datasetmounts",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
						{
							MountPoint: "http://ns-b/n-b",
						},
					},
				},
			},
			wantCheck: false,
			wantErr:   true,
		},
		{
			name: "referenced_dataset",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://ns-a/n-a",
						},
					},
				},
			},
			wantCheck: true,
			wantErr:   false,
		},
		{
			name: "no_mounts",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
			wantCheck: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCheck, err := CheckReferenceDataset(tt.dataset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Testcase %v CheckReferenceDataset() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotCheck != tt.wantCheck {
				t.Errorf("Testcase %v CheckReferenceDataset() = %v, want %v", tt.name, gotCheck, tt.wantCheck)
			}
		})
	}
}

func TestGetPhysicalDatasetSubPath(t *testing.T) {
	type args struct {
		dataset *datav1alpha1.Dataset
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "non empty sub path",
			args: args{
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "dataset://ns-a/ns-b/sub-c/sub-d",
							},
						},
					},
				},
			},
			want: []string{"sub-c/sub-d"},
		},
		{
			name: "empty sub path",
			args: args{
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "dataset://ns-a/ns-b/",
							},
						},
					},
				},
			},
			want: []string{""},
		},
		{
			name: "no sub path",
			args: args{
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "dataset://ns-a/ns-b",
							},
						},
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPhysicalDatasetSubPath(tt.args.dataset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPhysicalDatasetSubPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
