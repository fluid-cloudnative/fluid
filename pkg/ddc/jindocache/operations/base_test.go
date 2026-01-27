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
	Describe("NewJindoFileUtils", func() {

		It("should handle parameter order correctly", func() {
			result := NewJindoFileUtils("test-pod", "test-container", "test-namespace", fake.NullLogger())
			Expect(result.podName).To(Equal("test-pod"))
			Expect(result.namespace).To(Equal("test-namespace"))
			Expect(result.container).To(Equal("test-container"))
		})
	})

	Describe("exec", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecWithoutTimeoutErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecWithoutTimeoutErr)

				_, _, err := a.exec([]string{"jindo", "fs", "-report"}, false)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when exec succeeds", func() {
			It("should return stdout without error", func() {
				ExecWithoutTimeoutCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "Test stdout", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecWithoutTimeoutCommon)

				stdout, _, err := a.exec([]string{"jindo", "fs", "-report"}, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("Test stdout"))
			})
		})
	})

	Describe("ReportSummary", func() {
		var (
			a       JindoFileUtils
			patches *gomonkey.Patches
		)

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecErr)

				_, err := a.ReportSummary()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when exec succeeds", func() {
			It("should return result without error", func() {
				ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "Test stdout", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecCommon)

				_, err := a.ReportSummary()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GetUfsTotalSize", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecWithoutTimeoutErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecWithoutTimeoutErr)

				_, err := a.GetUfsTotalSize("/tmpDictionary")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when exec succeeds", func() {
			It("should return size without error", func() {
				ExecWithoutTimeoutCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "2      1    108 testUrl", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecWithoutTimeoutCommon)

				_, err := a.GetUfsTotalSize("/tmpDictionary")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when output is invalid", func() {
			It("should return a parse error", func() {
				ExecInvalidOutput := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "field1 field2", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecInvalidOutput)

				_, err := a.GetUfsTotalSize("/tmpDictionary")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Ready", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return false", func() {
				ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecErr)

				ready := a.Ready()
				Expect(ready).To(BeFalse())
			})
		})

		Context("when exec succeeds", func() {
			It("should return true", func() {
				ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "Test stdout ", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecCommon)

				ready := a.Ready()
				Expect(ready).To(BeTrue())
			})
		})
	})

	Describe("IsMounted", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecErr)

				_, err := a.IsMounted("/test/mount")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when mount point exists", func() {
			It("should return true", func() {
				ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "field1 field2 /test/mount field3\nfield1 field2 /other/mount field3", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecCommon)

				mounted, err := a.IsMounted("/test/mount")
				Expect(err).NotTo(HaveOccurred())
				Expect(mounted).To(BeTrue())
			})
		})

		Context("when mount point does not exist", func() {
			It("should return false", func() {
				ExecNotMounted := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "field1 field2 /other/mount field3", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecNotMounted)

				mounted, err := a.IsMounted("/test/mount")
				Expect(err).NotTo(HaveOccurred())
				Expect(mounted).To(BeFalse())
			})
		})

		Context("when output has short fields", func() {
			It("should return false", func() {
				ExecShortFields := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "field1 field2\nfield1", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecShortFields)

				mounted, err := a.IsMounted("/test/mount")
				Expect(err).NotTo(HaveOccurred())
				Expect(mounted).To(BeFalse())
			})
		})
	})

	Describe("Mount", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecErr)

				err := a.Mount("/test/mount", "oss://bucket/path")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when mount point already exists", func() {
			It("should ignore the error and return nil", func() {
				ExecAlreadyExists := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "Mount point already exists", "", errors.New("mount point exists")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecAlreadyExists)

				err := a.Mount("/test/mount", "oss://bucket/path")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when mount succeeds", func() {
			It("should return nil", func() {
				ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "mount success", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecCommon)

				err := a.Mount("/test/mount", "oss://bucket/path")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when exec returns other error", func() {
			It("should return an error", func() {
				ExecOtherErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "Some other error", "", errors.New("different error")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecOtherErr)

				err := a.Mount("/test/mount", "oss://bucket/path")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("IsRefreshed", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecErr)

				_, err := a.IsRefreshed()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when cache strategy is found", func() {
			It("should return true", func() {
				ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "line1\ncacheStrategy: some strategy\nline3", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecCommon)

				refreshed, err := a.IsRefreshed()
				Expect(err).NotTo(HaveOccurred())
				Expect(refreshed).To(BeTrue())
			})
		})

		Context("when cache strategy is not found", func() {
			It("should return false", func() {
				ExecNotRefreshed := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "line1\nline2\nline3", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecNotRefreshed)

				refreshed, err := a.IsRefreshed()
				Expect(err).NotTo(HaveOccurred())
				Expect(refreshed).To(BeFalse())
			})
		})
	})

	Describe("RefreshCacheSet", func() {
		var (
			a       *JindoFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = &JindoFileUtils{log: fake.NullLogger()}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return an error", func() {
				ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "", "", errors.New("fail to run the command")
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecErr)

				err := a.RefreshCacheSet()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when exec succeeds", func() {
			It("should return nil", func() {
				ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
					return "refresh cache set success", "", nil
				}
				patches = gomonkey.ApplyPrivateMethod(JindoFileUtils{}, "exec", ExecCommon)

				err := a.RefreshCacheSet()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
