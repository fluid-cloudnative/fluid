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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FindExtraArgsFromMetadata", func() {
	DescribeTable("should extract extra args from annotations",
		func(metaObj metav1.ObjectMeta, platform string, wantExtraArgs map[string]string) {
			gotExtraArgs := FindExtraArgsFromMetadata(metaObj, platform)
			Expect(gotExtraArgs).To(Equal(wantExtraArgs))
		},
		Entry("empty_annotations",
			metav1.ObjectMeta{Annotations: nil},
			"myplatform",
			map[string]string{},
		),
		Entry("without_extra_args",
			metav1.ObjectMeta{Annotations: map[string]string{"foo": "bar"}},
			"myplatform",
			map[string]string{},
		),
		Entry("with_extra_args",
			metav1.ObjectMeta{Annotations: map[string]string{"foo": "bar", "myplatform.fluid.io/key1": "value1", "myplatform.fluid.io/key2": "value2"}},
			"myplatform",
			map[string]string{"key1": "value1", "key2": "value2"},
		),
	)
})
