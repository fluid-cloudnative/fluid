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

func TestDataset_RemoveDataOperationInProgress(t *testing.T) {
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
			name: "test1",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						"DataLoad": "test1",
					},
				},
			},
			args: args{
				operationType: "DataLoad",
				name:          "test1",
			},
			want: "",
		},
		{
			name: "test2",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						"DataLoad": "test1,test2",
					},
				},
			},
			args: args{
				operationType: "DataLoad",
				name:          "test1",
			},
			want: "test2",
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

func TestDataset_SetDataOperationInProgress(t *testing.T) {
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
			name: "test1",
			fields: fields{
				Status: DatasetStatus{},
			},
			args: args{
				operationType: "DataLoad",
				name:          "test1",
			},
			want: "test1",
		},
		{
			name: "test2",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						"DataLoad": "test1",
					},
				},
			},
			args: args{
				operationType: "DataLoad",
				name:          "test2",
			},
			want: "test1,test2",
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
