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

package goosefs

import (
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformInitUsersWithoutRunAs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{},
		}, &GooseFS{}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{Log: fake.NullLogger()}
		engine.transformInitUsers(test.runtime, test.goosefsValue)
		if test.goosefsValue.InitUsers.Enabled {
			t.Errorf("expected init users are disabled, but got %v", test.goosefsValue.InitUsers.Enabled)
		}
	}
}

func TestTransformInitUsersWithRunAs(t *testing.T) {

	value := int64(1000)
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				RunAs: &datav1alpha1.User{
					UID:       &value,
					GID:       &value,
					UserName:  "user1",
					GroupName: "group1",
				},
			},
		}, &GooseFS{}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{
			Log:       fake.NullLogger(),
			initImage: common.DefaultInitImage,
		}
		engine.transformInitUsers(test.runtime, test.goosefsValue)
		if !test.goosefsValue.InitUsers.Enabled {
			t.Errorf("expected init users are enabled, but got %v", test.goosefsValue.InitUsers.Enabled)
		}

		imageInfo := strings.Split(common.DefaultInitImage, ":")
		if test.goosefsValue.InitUsers.Image != imageInfo[0] || test.goosefsValue.InitUsers.ImageTag != imageInfo[1] {
			t.Errorf("expected image info are set properly, but got image: %v, imageTag: %v", test.goosefsValue.InitUsers.Image, test.goosefsValue.InitUsers.ImageTag)
		}
	}
}

func TestTransformInitUsersImageOverwrite(t *testing.T) {
	value := int64(1000)
	image := "some-registry.some-repository"
	imageTag := "v1.0.0-abcdefg"
	var tests = []struct {
		runtime      *datav1alpha1.GooseFSRuntime
		goosefsValue *GooseFS
	}{
		{&datav1alpha1.GooseFSRuntime{
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				RunAs: &datav1alpha1.User{
					UID:       &value,
					GID:       &value,
					UserName:  "user1",
					GroupName: "group1",
				},
				InitUsers: datav1alpha1.InitUsersSpec{
					Image:    image,
					ImageTag: imageTag,
				},
			},
		}, &GooseFS{}},
	}
	for _, test := range tests {
		engine := &GooseFSEngine{
			Log:       fake.NullLogger(),
			initImage: common.DefaultInitImage,
		}
		engine.transformInitUsers(test.runtime, test.goosefsValue)
		if !test.goosefsValue.InitUsers.Enabled {
			t.Errorf("expected init users are enabled, but got %v", test.goosefsValue.InitUsers.Enabled)
		}

		if test.goosefsValue.InitUsers.Image != image || test.goosefsValue.InitUsers.ImageTag != imageTag {
			t.Errorf("expected image info should be overwrite, but got image: %v, imageTag: %v", test.goosefsValue.InitUsers.Image, test.goosefsValue.InitUsers.ImageTag)
		}
	}
}
