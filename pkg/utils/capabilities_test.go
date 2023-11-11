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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestTrimCapabilities(t *testing.T) {
	type args struct {
		inputs       []corev1.Capability
		excludeNames []string
	}
	tests := []struct {
		name        string
		args        args
		wantOutputs []corev1.Capability
	}{
		{
			name: "SYS_ADMIN_only",
			args: args{
				inputs:       []corev1.Capability{"SYS_ADMIN"},
				excludeNames: []string{"SYS_ADMIN"},
			},
			wantOutputs: []corev1.Capability{},
		},
		{
			name: "with_other_capabilities",
			args: args{
				inputs:       []corev1.Capability{"SYS_ADMIN", "CHOWN"},
				excludeNames: []string{"SYS_ADMIN"},
			},
			wantOutputs: []corev1.Capability{"CHOWN"},
		},
		{
			name: "exclude_multiple_capabilities",
			args: args{
				inputs:       []corev1.Capability{"SYS_ADMIN", "CHOWN", "SETPCAP"},
				excludeNames: []string{"SYS_ADMIN", "SETPCAP"},
			},
			wantOutputs: []corev1.Capability{"CHOWN"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutputs := TrimCapabilities(tt.args.inputs, tt.args.excludeNames); !reflect.DeepEqual(gotOutputs, tt.wantOutputs) {
				t.Errorf("TrimCapabilities() = %v, want %v", gotOutputs, tt.wantOutputs)
			}
		})
	}
}
