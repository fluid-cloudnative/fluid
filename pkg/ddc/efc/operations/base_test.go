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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NotExist = "not-exist"
	OtherErr = "other-err"
	FINE     = "fine"
)

var _ = Describe("EFCFileUtils", func() {
	Describe("NewEFCFileUtils", func() {
		It("should create a new EFCFileUtils instance with correct fields", func() {
			expectedResult := EFCFileUtils{
				podName:   "efcdemo",
				namespace: "default",
				container: "efc-master",
				log:       fake.NullLogger(),
			}
			result := NewEFCFileUtils("efcdemo", "efc-master", "default", fake.NullLogger())
			Expect(result).To(Equal(expectedResult))
		})
	})

	Describe("exec", func() {
		var (
			a       EFCFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = EFCFileUtils{
				podName:   "test-pod",
				namespace: "test-ns",
				container: "test-container",
				log:       fake.NullLogger(),
			}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when ExecCommandInContainerWithTimeout fails", func() {
			It("should return wrapped error", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "", "", errors.New("execution failed")
					})

				stdout, stderr, err := a.exec([]string{"test", "command"}, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error when executing command"))
				Expect(stdout).To(Equal(""))
				Expect(stderr).To(Equal(""))
			})

			It("should return wrapped error with stdout and stderr", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "some output", "some error", errors.New("execution failed")
					})

				stdout, stderr, err := a.exec([]string{"test", "command"}, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error when executing command"))
				Expect(stdout).To(Equal("some output"))
				Expect(stderr).To(Equal("some error"))
			})
		})

		Context("when ExecCommandInContainerWithTimeout succeeds", func() {
			It("should return stdout and stderr without logging when verbose is false", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "success output", "", nil
					})

				stdout, stderr, err := a.exec([]string{"ls", "-la"}, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("success output"))
				Expect(stderr).To(Equal(""))
			})

			It("should return stdout and stderr with logging when verbose is true", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "verbose output", "verbose stderr", nil
					})

				stdout, stderr, err := a.exec([]string{"cat", "file.txt"}, true)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("verbose output"))
				Expect(stderr).To(Equal("verbose stderr"))
			})
		})
	})

	Describe("DeleteDir", func() {
		var (
			a       EFCFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = EFCFileUtils{
				podName:   "test-pod",
				namespace: "test-ns",
				container: "test-container",
				log:       fake.NullLogger(),
			}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return error and log with stdout and stderr", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "delete failed output", "delete stderr", errors.New("delete failed")
					})

				err := a.DeleteDir("/test/path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error when executing command"))
			})
		})

		Context("when exec succeeds", func() {
			It("should delete directory without error", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "", "", nil
					})

				err := a.DeleteDir("/test/directory")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle deletion with output", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "deleted successfully", "", nil
					})

				err := a.DeleteDir("/another/path")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Ready", func() {
		var (
			a       EFCFileUtils
			patches *gomonkey.Patches
		)

		BeforeEach(func() {
			a = EFCFileUtils{
				podName:   "test-pod",
				namespace: "test-ns",
				container: "test-container",
				log:       fake.NullLogger(),
			}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when exec fails", func() {
			It("should return false and log error with stdout and stderr", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "mount output", "mount stderr", errors.New("mount check failed")
					})

				ready := a.Ready()
				Expect(ready).To(BeFalse())
			})
		})

		Context("when exec succeeds", func() {
			It("should return true", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "efc mount found", "", nil
					})

				ready := a.Ready()
				Expect(ready).To(BeTrue())
			})

			It("should return true with mount details", func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, cmd []string, timeout time.Duration) (string, string, error) {
						return "/dev/efc on /mnt type efc", "warning", nil
					})

				ready := a.Ready()
				Expect(ready).To(BeTrue())
			})
		})
	})
})
