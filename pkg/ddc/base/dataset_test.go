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

package base

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func TestGetOwnerDatasetUIDFromRuntimeMeta(t *testing.T) {
	uid := types.UID("dataset-uid-1")
	tests := []struct {
		name      string
		meta      metav1.ObjectMeta
		wantUID   types.UID
		expectErr bool
	}{
		{
			name: "single dataset owner with same name",
			meta: metav1.ObjectMeta{
				Name: "sample-runtime",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: datav1alpha1.Datasetkind,
						Name: "sample-runtime",
						UID:  uid,
					},
				},
			},
			wantUID:   uid,
			expectErr: false,
		},
		{
			name: "dataset owner name mismatch",
			meta: metav1.ObjectMeta{
				Name: "sample-runtime",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: datav1alpha1.Datasetkind,
						Name: "another-name",
						UID:  uid,
					},
				},
			},
			wantUID:   "",
			expectErr: true,
		},
		{
			name: "multiple dataset owners",
			meta: metav1.ObjectMeta{
				Name: "sample-runtime",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: datav1alpha1.Datasetkind,
						Name: "sample-runtime",
						UID:  uid,
					},
					{
						Kind: datav1alpha1.Datasetkind,
						Name: "sample-runtime",
						UID:  types.UID("dataset-uid-2"),
					},
				},
			},
			wantUID:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUID, err := GetOwnerDatasetUIDFromRuntimeMeta(tt.meta)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetOwnerDatasetUIDFromRuntimeMeta() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if gotUID != tt.wantUID {
				t.Errorf("GetOwnerDatasetUIDFromRuntimeMeta() uid = %v, want %v", gotUID, tt.wantUID)
			}
		})
	}
}
