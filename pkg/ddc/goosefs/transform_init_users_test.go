/*
Copyright 2022 The Fluid Authors.

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
