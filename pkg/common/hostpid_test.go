/*
Copyright 2024 The Fluid Authors.

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

package common

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("HostPIDEnabled", func() {
	ginkgo.DescribeTable("should correctly determine if HostPID is enabled",
		func(annotations map[string]string, expected bool) {
			gomega.Expect(HostPIDEnabled(annotations)).To(gomega.Equal(expected))
		},
		ginkgo.Entry("nil annotations return false", nil, false),
		ginkgo.Entry("empty annotations return false", map[string]string{}, false),
		ginkgo.Entry("wrong value returns false",
			map[string]string{RuntimeFuseHostPIDKey: "sss"}, false),
		ginkgo.Entry("'true' returns true",
			map[string]string{RuntimeFuseHostPIDKey: "true"}, true),
		ginkgo.Entry("'True' returns true",
			map[string]string{RuntimeFuseHostPIDKey: "True"}, true),
	)
})
