/*
Copyright 2020 The Fluid Author.

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

package alluxio

import (
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformInitUsersWithoutRunAs(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{},
		}, &Alluxio{}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{Log: fake.NullLogger()}
		engine.transformInitUsers(test.runtime, test.alluxioValue)
		if test.alluxioValue.InitUsers.Enabled {
			t.Errorf("expected init users are disabled, but got %v", test.alluxioValue.InitUsers.Enabled)
		}
	}
}

func TestTransformInitUsersWithRunAs(t *testing.T) {

	value := int64(1000)
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				RunAs: &datav1alpha1.User{
					UID:       &value,
					GID:       &value,
					UserName:  "user1",
					GroupName: "group1",
				},
			},
		}, &Alluxio{}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			initImage: common.DefaultInitImage,
		}
		engine.transformInitUsers(test.runtime, test.alluxioValue)
		if !test.alluxioValue.InitUsers.Enabled {
			t.Errorf("expected init users are enabled, but got %v", test.alluxioValue.InitUsers.Enabled)
		}

		imageInfo := strings.Split(common.DefaultInitImage, ":")
		if test.alluxioValue.InitUsers.Image != imageInfo[0] || test.alluxioValue.InitUsers.ImageTag != imageInfo[1] {
			t.Errorf("expected image info are set properly, but got image: %v, imageTag: %v", test.alluxioValue.InitUsers.Image, test.alluxioValue.InitUsers.ImageTag)
		}
	}
}

func TestTransformInitUsersImageOverwrite(t *testing.T) {
	value := int64(1000)
	image := "some-registry.some-repository"
	imageTag := "v1.0.0-abcdefg"
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
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
		}, &Alluxio{}},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{
			Log:       fake.NullLogger(),
			initImage: common.DefaultInitImage,
		}
		engine.transformInitUsers(test.runtime, test.alluxioValue)
		if !test.alluxioValue.InitUsers.Enabled {
			t.Errorf("expected init users are enabled, but got %v", test.alluxioValue.InitUsers.Enabled)
		}

		if test.alluxioValue.InitUsers.Image != image || test.alluxioValue.InitUsers.ImageTag != imageTag {
			t.Errorf("expected image info should be overwrite, but got image: %v, imageTag: %v", test.alluxioValue.InitUsers.Image, test.alluxioValue.InitUsers.ImageTag)
		}
	}
}
