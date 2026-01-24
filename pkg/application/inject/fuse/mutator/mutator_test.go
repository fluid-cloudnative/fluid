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

package mutator

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FindExtraArgsFromMetadata", func() {
	DescribeTable("when annotations do not contain extra args for the platform",
		func(annotations map[string]string) {
			metaObj := metav1.ObjectMeta{
				Annotations: annotations,
			}
			result := FindExtraArgsFromMetadata(metaObj, "myplatform")
			Expect(result).To(Equal(map[string]string{}))
		},
		Entry("with nil annotations", nil),
		Entry("with annotations but no matching ones", map[string]string{"foo": "bar"}),
	)
	Context("when annotations exist without extra args", func() {
		It("should return empty map", func() {
			metaObj := metav1.ObjectMeta{
				Annotations: map[string]string{"foo": "bar"},
			}
			result := FindExtraArgsFromMetadata(metaObj, "myplatform")
			Expect(result).To(Equal(map[string]string{}))
		})
	})

	Context("when annotations contain extra args", func() {
		It("should extract matching platform args", func() {
			metaObj := metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo":                      "bar",
					"myplatform.fluid.io/key1": "value1",
					"myplatform.fluid.io/key2": "value2",
				},
			}
			result := FindExtraArgsFromMetadata(metaObj, "myplatform")
			Expect(result).To(Equal(map[string]string{
				"key1": "value1",
				"key2": "value2",
			}))
		})
	})

	Context("when platform is empty", func() {
		It("should return empty map", func() {
			metaObj := metav1.ObjectMeta{
				Annotations: map[string]string{"myplatform.fluid.io/key1": "value1"},
			}
			result := FindExtraArgsFromMetadata(metaObj, "")
			Expect(result).To(Equal(map[string]string{}))
		})
	})

	Context("when multiple platforms are mixed", func() {
		It("should extract only matching platform args", func() {
			metaObj := metav1.ObjectMeta{
				Annotations: map[string]string{
					"platform1.fluid.io/key1":  "value1",
					"platform2.fluid.io/key2":  "value2",
					"myplatform.fluid.io/key3": "value3",
				},
			}
			result := FindExtraArgsFromMetadata(metaObj, "myplatform")
			Expect(result).To(Equal(map[string]string{
				"key3": "value3",
			}))
		})
	})

	Context("when partial matches exist", func() {
		It("should not match partial platform names", func() {
			metaObj := metav1.ObjectMeta{
				Annotations: map[string]string{
					"myplatform.other.io/key1":        "value1",
					"prefix-myplatform.fluid.io/key2": "value2",
				},
			}
			result := FindExtraArgsFromMetadata(metaObj, "myplatform")
			Expect(result).To(Equal(map[string]string{}))
		})
	})

	Context("when key has slashes in suffix", func() {
		It("should preserve slashes in the key", func() {
			metaObj := metav1.ObjectMeta{
				Annotations: map[string]string{
					"myplatform.fluid.io/key/with/slashes": "value",
				},
			}
			result := FindExtraArgsFromMetadata(metaObj, "myplatform")
			Expect(result).To(Equal(map[string]string{
				"key/with/slashes": "value",
			}))
		})
	})
})

var _ = Describe("BuildMutator", func() {
	var buildArgs MutatorBuildArgs

	BeforeEach(func() {
		buildArgs = MutatorBuildArgs{
			Client:    nil,
			Log:       logr.Discard(),
			Specs:     &MutatingPodSpecs{},
			Options:   common.FuseSidecarInjectOption{},
			ExtraArgs: nil,
		}
	})

	Context("when platform is default", func() {
		It("should build mutator successfully", func() {
			result, err := BuildMutator(buildArgs, utils.ServerlessPlatformDefault)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})

	Context("when platform is unprivileged", func() {
		It("should build mutator successfully", func() {
			result, err := BuildMutator(buildArgs, utils.ServerlessPlatformUnprivileged)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})

	Context("when platform is unknown", func() {
		It("should return an error", func() {
			result, err := BuildMutator(buildArgs, "unknown-platform")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("when platform is empty", func() {
		It("should return an error", func() {
			result, err := BuildMutator(buildArgs, "")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})
	DescribeTable("when platform is invalid",
		func(platform string) {
			result, err := BuildMutator(buildArgs, platform)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		},
		Entry("should return an error if platform is unknown", "unknown-platform"),
		Entry("should return an error if platform is empty", ""),
	)
})
