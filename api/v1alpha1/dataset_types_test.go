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
		{
			name: "test3",
			fields: fields{
				Status: DatasetStatus{},
			},
			args: args{
				operationType: "DataLoad",
				name:          "test1",
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
		{
			name: "test3",
			fields: fields{
				Status: DatasetStatus{
					OperationRef: map[string]string{
						"DataLoad": "test1",
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
						"DataLoad": "test",
					},
				},
			},
			args: args{
				operationType: "DataLoad",
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
