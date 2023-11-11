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
	"fmt"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

var testUid int64 = 1000
var testGid int64 = 1000
var testUserName = "test-user-1"
var testGroupName = "test-group-1"
var exampleUser = &datav1alpha1.User{
	UID:       &testUid,
	GID:       &testGid,
	UserName:  testUserName,
	GroupName: testGroupName,
}

func TestGetInitUserEnv(t *testing.T) {
	var expectedUserEnv = fmt.Sprintf("%d:%s:%d,%d:%s", testUid, testUserName, testGid, testGid, testGroupName)

	tests := []struct {
		name string
		user *datav1alpha1.User
		want string
	}{
		{
			name: "test for get init user env",
			user: exampleUser,
			want: expectedUserEnv,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInitUserEnv(tt.user); got != tt.want {
				t.Errorf("GetInitUserEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInitUsersArgs(t *testing.T) {
	var (
		expectedUserStr  = fmt.Sprintf("%d:%s:%d", testUid, testUserName, testGid)
		expectedGroupStr = fmt.Sprintf("%d:%s", testGid, testGroupName)
	)

	tests := []struct {
		name string
		user *datav1alpha1.User
		want []string
	}{
		{
			name: "test for init user args",
			user: exampleUser,
			want: []string{expectedUserStr, expectedGroupStr},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetInitUsersArgs(tt.user); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInitUsersArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
