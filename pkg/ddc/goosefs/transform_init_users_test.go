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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("TransformInitUsers", func() {
	Describe("without RunAs", func() {
		It("should disable init users when RunAs is not specified", func() {
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{},
			}
			goosefsValue := &GooseFS{}

			engine := &GooseFSEngine{Log: fake.NullLogger()}
			engine.transformInitUsers(runtime, goosefsValue)

			Expect(goosefsValue.InitUsers.Enabled).To(BeFalse())
		})
	})

	Describe("with RunAs", func() {
		It("should enable init users and set default image", func() {
			value := int64(1000)
			runtime := &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					RunAs: &datav1alpha1.User{
						UID:       &value,
						GID:       &value,
						UserName:  "user1",
						GroupName: "group1",
					},
				},
			}
			goosefsValue := &GooseFS{}

			engine := &GooseFSEngine{
				Log:       fake.NullLogger(),
				initImage: common.DefaultInitImage,
			}
			engine.transformInitUsers(runtime, goosefsValue)

			Expect(goosefsValue.InitUsers.Enabled).To(BeTrue())

			imageInfo := strings.Split(common.DefaultInitImage, ":")
			Expect(goosefsValue.InitUsers.Image).To(Equal(imageInfo[0]))
			Expect(goosefsValue.InitUsers.ImageTag).To(Equal(imageInfo[1]))
		})
	})

	Describe("with image overwrite", func() {
		It("should use custom image when specified in runtime", func() {
			value := int64(1000)
			image := "some-registry.some-repository"
			imageTag := "v1.0.0-abcdefg"

			runtime := &datav1alpha1.GooseFSRuntime{
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
			}
			goosefsValue := &GooseFS{}

			engine := &GooseFSEngine{
				Log:       fake.NullLogger(),
				initImage: common.DefaultInitImage,
			}
			engine.transformInitUsers(runtime, goosefsValue)

			Expect(goosefsValue.InitUsers.Enabled).To(BeTrue())
			Expect(goosefsValue.InitUsers.Image).To(Equal(image))
			Expect(goosefsValue.InitUsers.ImageTag).To(Equal(imageTag))
		})
	})
})
