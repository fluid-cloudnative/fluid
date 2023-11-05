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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestGetDataBackupRef(t *testing.T) {
	type args struct {
		object *v1alpha1.DataBackup
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				object: &v1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec:       v1alpha1.DataBackupSpec{},
				},
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataOperationKey(tt.args.object); got != tt.want {
				t.Errorf("GetDataBackupRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
