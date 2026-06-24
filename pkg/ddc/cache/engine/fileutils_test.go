/*
Copyright 2026 The Fluid Authors.

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

package engine

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("CacheFileUtils Tests", Label("pkg.ddc.cache.engine.fileutils_test.go"), func() {
	var (
		fileUtil CacheFileUtil
		log      logr.Logger
		patches  *gomonkey.Patches
	)

	BeforeEach(func() {
		log = GinkgoLogr
		fileUtil = NewCacheFileUtil("test-pod", "test-container", "default", log)
	})

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	Describe("Mount Tests", func() {
		Context("when command executes successfully", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "mount successful", "", nil
					})
			})

			It("should return stdout without error", func() {
				stdout, err := fileUtil.Execute([]string{"/mount.sh"}, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("mount successful"))
			})
		})

		Context("when command returns empty output", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "", nil
					})
			})

			It("should return empty stdout without error", func() {
				stdout, err := fileUtil.Execute([]string{"/mount.sh"}, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})
		})

		Context("when command execution fails", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "error output", errors.New("command failed")
					})
			})

			It("should return error", func() {
				stdout, err := fileUtil.Execute([]string{"/mount.sh"}, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("command failed"))
				Expect(stdout).To(BeEmpty())
			})
		})

		Context("when command has sensitive information", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						// Verify that the command was called (sensitive info filtering happens in exec)
						return "success", "", nil
					})
			})

			It("should execute command successfully", func() {
				command := []string{"/mount.sh", "--password=secret123"}
				stdout, err := fileUtil.Execute(command, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("success"))
			})
		})

		Context("when timeout is very short", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "completed", "", nil
					})
			})

			It("should still execute with short timeout", func() {
				stdout, err := fileUtil.Execute([]string{"/quick-mount.sh"}, 1*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("completed"))
			})
		})

		Context("when command contains multiple arguments", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						// Verify all arguments are passed
						Expect(command).To(HaveLen(4))
						return "multi-arg success", "", nil
					})
			})

			It("should pass all arguments correctly", func() {
				command := []string{"/mount.sh", "--source=/data", "--target=/mnt", "--options=rw"}
				stdout, err := fileUtil.Execute(command, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("multi-arg success"))
			})
		})

		Context("when stderr contains warnings but command succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "mount ok", "warning: deprecated option", nil
					})
			})

			It("should return success despite warnings in stderr", func() {
				stdout, err := fileUtil.Execute([]string{"/mount.sh"}, 30*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal("mount ok"))
			})
		})

		Context("when kubeclient returns wrapped error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithTimeout,
					func(podName, containerName, namespace string, command []string, timeout time.Duration) (stdout string, stderr string, err error) {
						return "", "", errors.New("connection refused")
					})
			})

			It("should wrap error with command context", func() {
				stdout, err := fileUtil.Execute([]string{"/mount.sh"}, 30*time.Second)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error when executing command"))
				Expect(err.Error()).To(ContainSubstring("connection refused"))
				Expect(stdout).To(BeEmpty())
			})
		})
	})

	Describe("newCacheFileUtils Tests", func() {
		Context("when creating new CacheFileUtils instance", func() {
			It("should return non-nil interface", func() {
				utils := NewCacheFileUtil("pod1", "container1", "ns1", log)
				Expect(utils).NotTo(BeNil())
			})
		})

		Context("when creating with different parameters", func() {
			It("should create separate instances", func() {
				utils1 := NewCacheFileUtil("pod1", "container1", "ns1", log)
				utils2 := NewCacheFileUtil("pod2", "container2", "ns2", log)
				Expect(utils1).NotTo(Equal(utils2))
			})
		})
	})
})
