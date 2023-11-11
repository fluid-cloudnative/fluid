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

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetExclusiveKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test for GetExclusiveKey",
			want: common.FluidExclusiveKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExclusiveKey(); got != tt.want {
				t.Errorf("GetExclusiveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExclusiveValue(t *testing.T) {
	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default test-dataset-1",
			args: args{
				name:      "test-dataset-1",
				namespace: "default",
			},
			want: "default_test-dataset-1",
		},
		{
			name: "otherns test-dataset-2",
			args: args{
				name:      "test-dataset-2",
				namespace: "otherns",
			},
			want: "otherns_test-dataset-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExclusiveValue(tt.args.namespace, tt.args.name); got != tt.want {
				t.Errorf("GetExclusiveValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
