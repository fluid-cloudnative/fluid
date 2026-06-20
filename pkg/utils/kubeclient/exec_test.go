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
)

const errFailToRun = "fail to run the function"

var _ = Describe("Exec Tests", func() {
	Context("InitClient", func() {
		var (
			pathExistPatch       *gomonkey.Patches
			originalBuildConfig  func(string, string) (*rest.Config, error)
			originalNewForConfig func(*rest.Config) (*kubernetes.Clientset, error)
		)

		BeforeEach(func() {
			t := GinkgoT()
			t.Setenv(common.RecommendedKubeConfigPathEnv, "Path for test")

			pathExistPatch = gomonkey.ApplyFunc(utils.PathExists, func(path string) bool {
				return true
			})

			restConfig = nil
			clientset = nil
			originalBuildConfig = buildConfigFromFlags
			originalNewForConfig = newClientsetForConfig
		})

		AfterEach(func() {
			if pathExistPatch != nil {
				pathExistPatch.Reset()
			}
			buildConfigFromFlags = originalBuildConfig
			newClientsetForConfig = originalNewForConfig
		})

		It("should fail when BuildConfigFromFlags fails", func() {
			buildConfigFromFlags = func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, errors.New(errFailToRun)
			}

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should fail when NewForConfig fails", func() {
			buildConfigFromFlags = func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			}
			newClientsetForConfig = func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, errors.New(errFailToRun)
			}

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should succeed when everything is correct", func() {
			buildConfigFromFlags = func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			}
			newClientsetForConfig = func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, nil
			}

			err := initClient()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail when kubeconfig path does not exist", func() {
			pathExistPatch.Reset()
			pathExistPatch = gomonkey.ApplyFunc(utils.PathExists, func(path string) bool {
				return false
			})

			buildConfigFromFlags = func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, errors.New(errFailToRun)
			}

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should fail when NewForConfig fails (second path)", func() {
			buildConfigFromFlags = func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			}
			newClientsetForConfig = func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, errors.New(errFailToRun)
			}

			err := initClient()
			Expect(err).To(HaveOccurred())
		})

		It("should succeed when everything is correct (second path)", func() {
			buildConfigFromFlags = func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, nil
			}
			newClientsetForConfig = func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, nil
			}

			err := initClient()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("ExecCommandInContainerWithTimeout", func() {
		var originalExecWithFullOutput func(context.Context, string, string, string, []string) (string, string, error)

		BeforeEach(func() {
			originalExecWithFullOutput = execInContainerWithOutput
		})

		AfterEach(func() {
			execInContainerWithOutput = originalExecWithFullOutput
		})

		It("should return correctly when command completes before timeout", func() {
			expectedStdout := "test output"
			expectedStderr := "test error output"

			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				return expectedStdout, expectedStderr, nil
			}

			stdout, stderr, err := ExecCommandInContainerWithTimeoutContext(context.Background(),
				"test-pod", "test-container", "test-namespace",
				[]string{"echo", "hello"}, 5*time.Second)

			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(Equal(expectedStdout))
			Expect(stderr).To(Equal(expectedStderr))
		})

		It("should return timeout error when command takes longer than timeout", func() {
			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				<-ctx.Done()
				return "should not see this", "should not see this either", ctx.Err()
			}

			timeout := 100 * time.Millisecond
			start := time.Now()

			stdout, stderr, err := ExecCommandInContainerWithTimeoutContext(context.Background(),
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

			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				return "", "command not found", expectedErr
			}

			_, stderr, err := ExecCommandInContainerWithTimeoutContext(context.Background(),
				"test-pod", "test-container", "test-namespace",
				[]string{"nonexistent-command"}, 5*time.Second)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expectedErr.Error()))
			Expect(stderr).To(Equal("command not found"))
		})

		It("should return context cancellation when parent context is canceled", func() {
			parentCtx, cancel := context.WithCancel(context.Background())
			cancel()

			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				return "", "", ctx.Err()
			}

			stdout, stderr, err := ExecCommandInContainerWithTimeoutContext(
				parentCtx,
				"test-pod", "test-container", "test-namespace",
				[]string{"echo", "hello"}, 5*time.Second)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("context canceled"))
			Expect(stdout).To(BeEmpty())
			Expect(stderr).To(BeEmpty())
		})

		It("should not have data races", func() {
			var callCount int32

			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				count := atomic.AddInt32(&callCount, 1)
				delay := time.Duration(count%3) * 10 * time.Millisecond
				select {
				case <-time.After(delay):
					return "stdout-" + podName, "stderr-" + podName, nil
				case <-ctx.Done():
					return "", "", ctx.Err()
				}
			}

			const numGoroutines = 50
			var wg sync.WaitGroup
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()
					podName := "pod-" + string(rune('a'+id%26))
					stdout, stderr, err := ExecCommandInContainerWithTimeoutContext(context.Background(),
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

			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				atomic.AddInt32(&activeGoroutines, 1)
				defer atomic.AddInt32(&activeGoroutines, -1)

				select {
				case <-time.After(1 * time.Second):
					return "completed", "", nil
				case <-ctx.Done():
					return "", "", ctx.Err()
				}
			}

			const numCalls = 10
			var wg sync.WaitGroup
			wg.Add(numCalls)

			for i := 0; i < numCalls; i++ {
				go func() {
					defer wg.Done()
					_, _, _ = ExecCommandInContainerWithTimeoutContext(context.Background(),
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

			execInContainerWithOutput = func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				<-ctx.Done()
				mu.Lock()
				ctxCancelled = true
				mu.Unlock()
				return "", "", ctx.Err()
			}

			_, _, err := ExecCommandInContainerWithTimeoutContext(context.Background(),
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

})
