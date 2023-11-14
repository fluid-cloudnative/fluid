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

package utils

import (
	"strconv"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func GetInitUsersArgs(user *datav1alpha1.User) []string {
	uid := strconv.FormatInt(*user.UID, 10)
	gid := strconv.FormatInt(*user.GID, 10)
	username := user.UserName
	args := []string{uid + ":" + username + ":" + gid,
		gid + ":" + user.GroupName}
	return args
}

func GetInitUserEnv(user *datav1alpha1.User) string {
	return strings.Join(GetInitUsersArgs(user), ",")
}
