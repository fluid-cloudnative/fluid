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
