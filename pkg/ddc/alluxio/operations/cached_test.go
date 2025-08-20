/*
Copyright 2021 The Fluid Authors.

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

package operations

import (
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)


var _ = Describe("AlluxioFileUtils.CachedState", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		It("should return an error", func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)

			_, err := a.CachedState()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully", func() {
		It("should not return an error", func() {
			ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alluxio cluster summary: \n    Master Address: 192.168.0.193:20009  \n Used Capacity: 0B\n", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommon)

			cached, err := a.CachedState()
			Expect(err).NotTo(HaveOccurred())
			Expect(cached).To(Equal(int64(0)))
		})
	})
})

var _ = Describe("AlluxioFileUtils.CleanCache", func() {
	var (
		a     *AlluxioFileUtils
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		a = &AlluxioFileUtils{log: fake.NullLogger()}
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when exec returns an error", func() {
		It("should return an error", func() {
			ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecErr)

			err := a.CleanCache("/", 30)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when exec executes successfully on Ubuntu", func() {
		It("should not return an error", func() {
			ExecCommonUbuntu := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Ubuntu", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommonUbuntu)

			err := a.CleanCache("/", 30)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when exec executes successfully on Alpine", func() {
		It("should not return an error", func() {
			ExecCommonAlpine := func(command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Alpine", "", nil
			}
			patch = gomonkey.ApplyPrivateMethod(a, "exec", ExecCommonAlpine)

			err := a.CleanCache("/", 30)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})