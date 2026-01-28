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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JindoFileUtils", func() {
	Describe("CleanCache", func() {
		It("should return error when exec fails", func() {
			execErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
				return "", "", errors.New("fail to run the command")
			}

			patches := gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", execErr)
			defer patches.Reset()

			a := &JindoFileUtils{log: fake.NullLogger()}
			err := a.CleanCache()
			Expect(err).To(HaveOccurred())
		})

		It("should succeed when exec succeeds", func() {
			execCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
				return "Test stout", "", nil
			}

			patches := gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", execCommon)
			defer patches.Reset()

			a := &JindoFileUtils{log: fake.NullLogger()}
			err := a.CleanCache()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
