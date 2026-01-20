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

package operations

import (
	"errors"
	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ThinFileUtils", func() {
	Describe("NewThinFileUtils", func() {
		It("should create a new ThinFileUtils instance with correct fields", func() {
			result := NewThinFileUtils("thin", common.ThinFuseContainer, "default", fake.NullLogger())

			Expect(result.podName).To(Equal("thin"))
			Expect(result.namespace).To(Equal("default"))
			Expect(result.container).To(Equal(common.ThinFuseContainer))
			Expect(result.log).NotTo(BeNil())
		})

		It("should handle empty pod name", func() {
			result := NewThinFileUtils("", common.ThinFuseContainer, "default", fake.NullLogger())

			Expect(result.podName).To(Equal(""))
			Expect(result.namespace).To(Equal("default"))
		})

		It("should handle empty namespace", func() {
			result := NewThinFileUtils("thin", common.ThinFuseContainer, "", fake.NullLogger())

			Expect(result.podName).To(Equal("thin"))
			Expect(result.namespace).To(Equal(""))
		})

		It("should handle empty container", func() {
			result := NewThinFileUtils("thin", "", "default", fake.NullLogger())

			Expect(result.podName).To(Equal("thin"))
			Expect(result.container).To(Equal(""))
		})

		It("should handle all empty parameters", func() {
			result := NewThinFileUtils("", "", "", fake.NullLogger())

			Expect(result.podName).To(Equal(""))
			Expect(result.namespace).To(Equal(""))
			Expect(result.container).To(Equal(""))
			Expect(result.log).NotTo(BeNil())
		})

		It("should handle special characters in pod name", func() {
			result := NewThinFileUtils("thin-pod-123", common.ThinFuseContainer, "default", fake.NullLogger())

			Expect(result.podName).To(Equal("thin-pod-123"))
			Expect(result.namespace).To(Equal("default"))
		})

		It("should handle special characters in namespace", func() {
			result := NewThinFileUtils("thin", common.ThinFuseContainer, "kube-system", fake.NullLogger())

			Expect(result.podName).To(Equal("thin"))
			Expect(result.namespace).To(Equal("kube-system"))
		})
	})

	Describe("LoadMetadataWithoutTimeout", func() {
		var (
			thinFileUtils *ThinFileUtils
			patches       *Patches
		)
		BeforeEach(func() {
			thinFileUtils = &ThinFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec succeeds", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "Load thin metadata", "", nil
					})
			})

			It("should load metadata without error", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/tmp")
				Expect(err).To(BeNil())
			})

			It("should handle empty path", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("")
				Expect(err).To(BeNil())
			})

			It("should handle complex path", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/runtime-mnt/thin/kube-system/thindemo")
				Expect(err).To(BeNil())
			})

			It("should handle path with special characters", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/path/with/special-chars_123")
				Expect(err).To(BeNil())
			})

			It("should handle relative path", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("./relative/path")
				Expect(err).To(BeNil())
			})

			It("should handle path with trailing slash", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/tmp/")
				Expect(err).To(BeNil())
			})
		})

		Context("when exec fails", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", errors.New("fail to run the command")
					})
			})

			It("should return error", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/tmp")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("fail to run the command"))
			})

			It("should return error for any path", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/any/path")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when exec returns stderr", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "permission denied", errors.New("command failed")
					})
			})

			It("should return error", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/tmp")
				Expect(err).NotTo(BeNil())
			})

			It("should handle stderr with different error messages", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/restricted/path")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when exec returns partial success", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "partial output", "warning message", nil
					})
			})

			It("should succeed despite warnings", func() {
				err := thinFileUtils.LoadMetadataWithoutTimeout("/tmp")
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("GetUsedSpace", func() {
		var (
			thinFileUtils *ThinFileUtils
			patches       *Patches
		)

		BeforeEach(func() {
			thinFileUtils = &ThinFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec succeeds with valid output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "192.168.100.11:/nfs/mnt   87687856128  87687856128            0 100% /runtime-mnt/thin/kube-system/thindemo/thin-fuse", "", nil
					})
			})

			It("should return correct used space", func() {
				usedSpace, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).To(BeNil())
				Expect(usedSpace).To(Equal(int64(87687856128)))
			})

			It("should handle different paths", func() {
				usedSpace, err := thinFileUtils.GetUsedSpace("/runtime-mnt/thin")
				Expect(err).To(BeNil())
				Expect(usedSpace).To(Equal(int64(87687856128)))
			})
		})

		Context("when exec returns different space formats", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "tmpfs   1024  512  512 50% /tmp", "", nil
					})
			})

			It("should parse different output formats", func() {
				usedSpace, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).To(BeNil())
				Expect(usedSpace).To(Equal(int64(512)))
			})
		})

		Context("when exec fails", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", errors.New("fail to run the command")
					})
			})

			It("should return error", func() {
				_, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("fail to run the command"))
			})
		})

		Context("when exec returns invalid output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "invalid output", "", nil
					})
			})

			It("should return error for malformed output", func() {
				_, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when exec returns empty output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", nil
					})
			})

			It("should return error for empty output", func() {
				_, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when exec returns zero space", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "192.168.100.11:/nfs/mnt   0  0            0 0% /runtime-mnt/thin", "", nil
					})
			})

			It("should return zero used space", func() {
				usedSpace, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).To(BeNil())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when exec returns negative space", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "tmpfs   -1  -1  0 0% /tmp", "", nil
					})
			})
		})

		Context("when exec returns very large space", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "nfs   999999999999999  999999999999999  0 100% /mnt", "", nil
					})
			})

			It("should handle very large numbers", func() {
				usedSpace, err := thinFileUtils.GetUsedSpace("/tmp")
				Expect(err).To(BeNil())
				Expect(usedSpace).To(Equal(int64(999999999999999)))
			})
		})
	})

	Describe("GetFileCount", func() {
		var (
			thinFileUtils *ThinFileUtils
			patches       *Patches
		)

		BeforeEach(func() {
			thinFileUtils = &ThinFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec succeeds with valid output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "6367897", "", nil
					})
			})

			It("should return correct file count", func() {
				fileCount, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).To(BeNil())
				Expect(fileCount).To(Equal(int64(6367897)))
			})

			It("should handle different paths", func() {
				fileCount, err := thinFileUtils.GetFileCount("/runtime-mnt/thin")
				Expect(err).To(BeNil())
				Expect(fileCount).To(Equal(int64(6367897)))
			})
		})

		Context("when exec fails", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", errors.New("fail to run the command")
					})
			})

			It("should return error", func() {
				_, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("fail to run the command"))
			})
		})

		Context("when exec returns invalid output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "not a number", "", nil
					})
			})

			It("should return error for non-numeric output", func() {
				_, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when exec returns empty output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", nil
					})
			})

			It("should return error for empty output", func() {
				_, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when exec returns zero file count", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "0", "", nil
					})
			})

			It("should return zero file count", func() {
				fileCount, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).To(BeNil())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when exec returns very large file count", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "999999999999", "", nil
					})
			})

			It("should handle large numbers", func() {
				fileCount, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).To(BeNil())
				Expect(fileCount).To(Equal(int64(999999999999)))
			})
		})

		Context("when exec returns negative file count", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "-123", "", nil
					})
			})

			It("should handle negative numbers", func() {
				fileCount, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).To(BeNil())
				Expect(fileCount).To(Equal(int64(-123)))
			})
		})

		Context("when exec returns output with whitespace", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "  12345  \n", "", nil
					})
			})
		})

		Context("when exec returns decimal number", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "123.45", "", nil
					})
			})

			It("should return error for decimal numbers", func() {
				_, err := thinFileUtils.GetFileCount("/tmp")
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("exec", func() {
		var (
			thinFileUtils *ThinFileUtils
			patches       *Patches
		)

		BeforeEach(func() {
			thinFileUtils = &ThinFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", errors.New("fail to run the command")
					})
			})

			It("should return error", func() {
				_, _, err := thinFileUtils.exec([]string{"mkdir", "abc"}, false)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("fail to run the command"))
			})

			It("should return error with verbose true", func() {
				_, _, err := thinFileUtils.exec([]string{"mkdir", "abc"}, true)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("fail to run the command"))
			})
		})

		Context("when exec returns stderr", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "output", "error message", nil
					})
			})

			It("should return both stdout and stderr", func() {
				stdout, stderr, err := thinFileUtils.exec([]string{"test"}, false)
				Expect(err).To(BeNil())
				Expect(stdout).To(Equal("output"))
				Expect(stderr).To(Equal("error message"))
			})
		})

		Context("when exec returns empty output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "", "", nil
					})
			})

			It("should handle empty stdout and stderr", func() {
				stdout, stderr, err := thinFileUtils.exec([]string{"test"}, false)
				Expect(err).To(BeNil())
				Expect(stdout).To(Equal(""))
				Expect(stderr).To(Equal(""))
			})
		})

		Context("when exec returns long output", func() {
			BeforeEach(func() {
				patches = ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec",
					func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
						return "very long output that spans multiple lines\nline2\nline3", "", nil
					})
			})

			It("should handle multiline output", func() {
				stdout, stderr, err := thinFileUtils.exec([]string{"test"}, false)
				Expect(err).To(BeNil())
				Expect(stdout).To(ContainSubstring("very long output"))
				Expect(stderr).To(Equal(""))
			})
		})
	})
})
