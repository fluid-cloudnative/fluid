/*
Copyright 2023 The Fluid Author.

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

package kubeclient

import (
	"context"
	"errors"

	"sync"
	"sync/atomic"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const errFailToRun = "fail to run the function"

var _ = Describe("Exec Tests", func() {
	Context("InitClient", func() {
		var (
			pathExistPatch            *gomonkey.Patches
			buildConfigFromFlagsPatch *gomonkey.Patches
			newForConfigPatch         *gomonkey.Patches
		)

		BeforeEach(func() {
			t := GinkgoT()
			t.Setenv(common.RecommendedKubeConfigPathEnv, "Path for test")

			pathExistPatch = gomonkey.ApplyFunc(utils.PathExists, func(path string) bool {
				return true
			})

			restConfig = nil
			clientset = nil
		})

		AfterEach(func() {
			if pathExistPatch != nil {
				pathExistPatch.Reset()
			}
			if buildConfigFromFlagsPatch != nil {
				buildConfigFromFlagsPatch.Reset()
			}
			if newForConfigPatch != nil {
				newForConfigPatch.Reset()
			}
		})

		It("should fail when BuildConfigFromFlags fails", func() {
			buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, errors.New(errFailToRun)
			})

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should fail when NewForConfig fails", func() {
			buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			})
			newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, errors.New(errFailToRun)
			})

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should succeed when everything is correct", func() {
			buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			})
			newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, nil
			})

			err := initClient()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail when kubeconfig path does not exist", func() {
			pathExistPatch.Reset()
			pathExistPatch = gomonkey.ApplyFunc(utils.PathExists, func(path string) bool {
				return false
			})

			buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, errors.New(errFailToRun)
			})

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should fail when NewForConfig fails (second path)", func() {
			buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			})
			newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, errors.New(errFailToRun)
			})

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should succeed when everything is correct (second path)", func() {
			buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			})
			newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, nil
			})

			err := initClient()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("ExecCommandInContainerWithTimeout", func() {
		It("should return correctly when command completes before timeout", func() {
			expectedStdout := "test output"
			expectedStderr := "test error output"

			patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
				func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
					return expectedStdout, expectedStderr, nil
				})
			defer patch.Reset()

			stdout, stderr, err := ExecCommandInContainerWithTimeout(
				"test-pod", "test-container", "test-namespace",
				[]string{"echo", "hello"}, 5*time.Second)

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(expectedStdout))
			Expect(stderr).To(Equal(expectedStderr))
		})

		It("should return timeout error when command takes longer than timeout", func() {
			patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
				func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
					// Wait for context cancellation
					<-ctx.Done()
					return "should not see this", "should not see this either", ctx.Err()
				})
			defer patch.Reset()

			timeout := 100 * time.Millisecond
			start := time.Now()

			stdout, stderr, err := ExecCommandInContainerWithTimeout(
				"test-pod", "test-container", "test-namespace",
				[]string{"sleep", "10"}, timeout)

			elapsed := time.Since(start)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timed out"))
			Expect(stdout).To(BeEmpty())
			Expect(stderr).To(BeEmpty())
			Expect(elapsed).To(BeNumerically("<=", 2*timeout))
		})

		It("should propagate errors from underlying exec function", func() {
			expectedErr := errors.New("exec failed: command not found")

			patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
				func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
					return "", "command not found", expectedErr
				})
			defer patch.Reset()

			_, stderr, err := ExecCommandInContainerWithTimeout(
				"test-pod", "test-container", "test-namespace",
				[]string{"nonexistent-command"}, 5*time.Second)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedErr.Error()))
			Expect(stderr).To(Equal("command not found"))
		})

		It("should not have data races", func() {
			var callCount int32

			patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
				func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
					count := atomic.AddInt32(&callCount, 1)
					// Vary the execution time to increase chance of race detection
					delay := time.Duration(count%3) * 10 * time.Millisecond
					select {
					case <-time.After(delay):
						return "stdout-" + podName, "stderr-" + podName, nil
					case <-ctx.Done():
						return "", "", ctx.Err()
					}
				})
			defer patch.Reset()

			const numGoroutines = 50
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()
					podName := "pod-" + string(rune('a'+id%26))
					stdout, stderr, err := ExecCommandInContainerWithTimeout(
						podName, "container", "namespace",
						[]string{"test"}, 500*time.Millisecond)

					if err == nil {
						Expect(stdout).To(HavePrefix("stdout-"))
						Expect(stderr).To(HavePrefix("stderr-"))
					}
				}(i)
			}

			wg.Wait()
		})

		It("should not leak goroutines on timeout", func() {
			var activeGoroutines int32

			patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
				func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
					atomic.AddInt32(&activeGoroutines, 1)
					defer atomic.AddInt32(&activeGoroutines, -1)

					// Simulate a slow operation
					select {
					case <-time.After(1 * time.Second):
						return "completed", "", nil
					case <-ctx.Done():
						return "", "", ctx.Err()
					}
				})
			defer patch.Reset()

			const numCalls = 10
			var wg sync.WaitGroup
			wg.Add(numCalls)

			for i := 0; i < numCalls; i++ {
				go func() {
					defer wg.Done()
					_, _, _ = ExecCommandInContainerWithTimeout(
						"pod", "container", "namespace",
						[]string{"slow-command"}, 50*time.Millisecond)
				}()
			}

			wg.Wait()

			Eventually(func() int32 {
				return atomic.LoadInt32(&activeGoroutines)
			}).WithTimeout(500 * time.Millisecond).Should(Equal(int32(0)))
		})

		It("should propagate context cancellation", func() {
			var ctxCancelled bool
			var mu sync.Mutex

			patch := gomonkey.ApplyFunc(ExecCommandInContainerWithFullOutput,
				func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
					<-ctx.Done()
					mu.Lock()
					ctxCancelled = true
					mu.Unlock()
					return "", "", ctx.Err()
				})
			defer patch.Reset()

			_, _, err := ExecCommandInContainerWithTimeout(
				"pod", "container", "namespace",
				[]string{"command"}, 10*time.Millisecond)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("timed out"))

			Eventually(func() bool {
				mu.Lock()
				defer mu.Unlock()
				return ctxCancelled
			}).WithTimeout(100 * time.Millisecond).Should(BeTrue())
		})
	})

	Context("ExecResult", func() {
		It("should store values correctly", func() {
			result := execResult{
				stdout: "test stdout",
				stderr: "test stderr",
				err:    errors.New("test error"),
			}

			Expect(result.stdout).To(Equal("test stdout"))
			Expect(result.stderr).To(Equal("test stderr"))
			Expect(result.err).To(HaveOccurred())
			Expect(result.err.Error()).To(Equal("test error"))
		})
	})
})
