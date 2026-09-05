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

package v1alpha1

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testDataLoad = "DataLoad"
	testLoad1    = "load-1"
	test1Name    = "test1"
)

func TestDatasetRemoveDataOperationInProgress(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       DatasetSpec
		Status     DatasetStatus
	}
	type args struct {
		operationType string
		name          string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: test1Name,
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: test1Name,
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				name:          test1Name,
			},
			want: "",
		},
		{
			name: "test2",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: "test1,test2",
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				name:          test1Name,
			},
			want: "test2",
		},
		{
			name: "test3",
			fields: fields{
				Status: DatasetStatus{},
			},
			args: args{
				operationType: testDataLoad,
				name:          test1Name,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := &Dataset{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := dataset.RemoveDataOperationInProgress(tt.args.operationType, tt.args.name); got != tt.want {
				t.Errorf("RemoveDataOperationInProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatasetSetDataOperationInProgress(t *testing.T) {
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       DatasetSpec
		Status     DatasetStatus
	}
	type args struct {
		operationType string
		name          string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: test1Name,
			fields: fields{
				Status: DatasetStatus{},
			},
			args: args{
				operationType: testDataLoad,
				name:          test1Name,
			},
			want: test1Name,
		},
		{
			name: "test2",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: test1Name,
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				name:          "test2",
			},
			want: "test1,test2",
		},
		{
			name: "test3",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: test1Name,
					},
				},
			},
			args: args{
				operationType: "DataMigrate",
				name:          "test",
			},
			want: "test",
		},
		{
			name: "test4",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: "test",
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				name:          "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := &Dataset{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			dataset.SetDataOperationInProgress(tt.args.operationType, tt.args.name)
			if got := dataset.GetDataOperationInProgress(tt.args.operationType); got != tt.want {
				t.Errorf("SetDataOperationInProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatasetCanStartDataOperation(t *testing.T) {
	type fields struct {
		Status DatasetStatus
	}
	type args struct {
		operationType string
		maxParallel   int32
		name          string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "empty_status",
			fields: fields{
				Status: DatasetStatus{},
			},
			args: args{
				operationType: testDataLoad,
				maxParallel:   1,
				name:          testLoad1,
			},
			want: true,
		},
		{
			name: "already_in_progress_reentrant",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: testLoad1,
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				maxParallel:   1,
				name:          testLoad1,
			},
			want: true,
		},
		{
			name: "blocked_by_max_parallel_1",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: testLoad1,
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				maxParallel:   1,
				name:          "load-2",
			},
			want: false,
		},
		{
			name: "allowed_by_max_parallel_2",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: testLoad1,
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				maxParallel:   2,
				name:          "load-2",
			},
			want: true,
		},
		{
			name: "blocked_by_max_parallel_2",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: "load-1,load-2",
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				maxParallel:   2,
				name:          "load-3",
			},
			want: false,
		},
		{
			name: "zero_max_parallel_treated_as_unlimited",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						testDataLoad: testLoad1,
					},
				},
			},
			args: args{
				operationType: testDataLoad,
				maxParallel:   0,
				name:          "load-2",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := &Dataset{
				Status: tt.fields.Status,
			}
			if got := dataset.CanStartDataOperation(tt.args.operationType, tt.args.maxParallel, tt.args.name); got != tt.want {
				t.Errorf("CanStartDataOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}
