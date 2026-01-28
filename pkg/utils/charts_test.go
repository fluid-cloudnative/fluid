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
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	Describe("PathExists", func() {
		It("should return true for existing path", func() {
			path := os.TempDir()
			Expect(PathExists(path)).To(BeTrue())
		})

		It("should return false for non-existing path", func() {
			path := os.TempDir() + "test/"
			Expect(PathExists(path)).To(BeFalse())
		})
	})

	Describe("GetChartsDirectory", func() {
		It("should return /charts when HOME/charts does not exist", func() {
			f, err := os.CreateTemp("", "test")
			Expect(err).NotTo(HaveOccurred())
			testDir := f.Name()

			GinkgoT().Setenv("HOME", testDir)
			Expect(GetChartsDirectory()).To(Equal("/charts"))
		})

		It("should return /charts when HOME/charts exists", func() {
			tempDir, err := os.MkdirTemp("", "test")
			Expect(err).NotTo(HaveOccurred())

			GinkgoT().Setenv("HOME", tempDir)
			homeChartsFolder := os.Getenv("HOME") + "/charts"
			err = os.Mkdir(homeChartsFolder, 0600)
			Expect(err).NotTo(HaveOccurred())
			Expect(GetChartsDirectory()).To(Equal("/charts"))
		})
	})
})
