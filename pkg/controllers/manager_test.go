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

package controllers

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NewFluidControllerRateLimiter", func() {
	It("should create a rate limiter with valid parameters", func() {
		limiter := NewFluidControllerRateLimiter(
			5*time.Millisecond,
			1000*time.Second,
			10,
			100,
		)
		Expect(limiter).NotTo(BeNil())
	})

	It("should return increasing delays for repeated failures on the same item", func() {
		limiter := NewFluidControllerRateLimiter(
			5*time.Millisecond,
			1000*time.Second,
			10,
			100,
		)
		first := limiter.When("test-item")
		second := limiter.When("test-item")
		Expect(second).To(BeNumerically(">=", first))
	})

	It("should return different delays for different items", func() {
		limiter := NewFluidControllerRateLimiter(
			5*time.Millisecond,
			1000*time.Second,
			10,
			100,
		)
		// Advance item-a several times to increase its delay
		for i := 0; i < 5; i++ {
			limiter.When("item-a")
		}
		delayA := limiter.When("item-a")
		delayB := limiter.When("item-b")
		Expect(delayA).To(BeNumerically(">", delayB))
	})

	It("should respect the max delay parameter", func() {
		maxDelay := 100 * time.Millisecond
		limiter := NewFluidControllerRateLimiter(
			5*time.Millisecond,
			maxDelay,
			10,
			100,
		)
		// Push the item through many failures
		for i := 0; i < 100; i++ {
			limiter.When("test-item")
		}
		delay := limiter.When("test-item")
		Expect(delay).To(BeNumerically("<=", maxDelay))
	})

	It("should reset delay after forgetting an item", func() {
		limiter := NewFluidControllerRateLimiter(
			5*time.Millisecond,
			1000*time.Second,
			10,
			100,
		)
		for i := 0; i < 10; i++ {
			limiter.When("test-item")
		}
		limiter.Forget("test-item")
		delay := limiter.When("test-item")
		firstDelay := limiter.When("fresh-item")
		// After forget, the item's delay should be close to the initial value
		Expect(delay).To(BeNumerically("<=", firstDelay*2))
	})
})

var _ = Describe("manager client and config helpers", func() {
	Describe("NewFluidControllerClient", func() {
		var (
			originalVal string
			wasSet      bool
		)

		BeforeEach(func() {
			originalVal, wasSet = os.LookupEnv("HELM_DRIVER")
		})

		AfterEach(func() {
			if wasSet {
				Expect(os.Setenv("HELM_DRIVER", originalVal)).To(Succeed())
			} else {
				Expect(os.Unsetenv("HELM_DRIVER")).To(Succeed())
			}
		})

		It("returns error on secret driver path with nil rest config", func() {
			Expect(os.Setenv("HELM_DRIVER", "secret")).To(Succeed())

			c, err := NewFluidControllerClient(nil, client.Options{})
			Expect(err).To(HaveOccurred())
			Expect(c).To(BeNil())
		})

		It("returns error on cache-bypass path with nil rest config", func() {
			Expect(os.Setenv("HELM_DRIVER", "configmap")).To(Succeed())

			c, err := NewFluidControllerClient(nil, client.Options{Cache: &client.CacheOptions{}})
			Expect(err).To(HaveOccurred())
			Expect(c).To(BeNil())
		})
	})

	Describe("NewCacheClientBypassSecrets", func() {
		It("returns error with nil rest config", func() {
			c, err := NewCacheClientBypassSecrets(nil, client.Options{Cache: &client.CacheOptions{}})
			Expect(err).To(HaveOccurred())
			Expect(c).To(BeNil())
		})
	})

	Describe("GetConfigOrDieWithQPSAndBurst", func() {
		var (
			originalKubeconfig string
			wasSet             bool
		)

		BeforeEach(func() {
			originalKubeconfig, wasSet = os.LookupEnv("KUBECONFIG")
		})

		AfterEach(func() {
			if wasSet {
				Expect(os.Setenv("KUBECONFIG", originalKubeconfig)).To(Succeed())
			} else {
				Expect(os.Unsetenv("KUBECONFIG")).To(Succeed())
			}
		})

		It("sets qps and burst when both are positive", func() {
			if os.Getenv("FLUID_GET_CONFIG_SUBPROCESS") == "1" {
				cfg := GetConfigOrDieWithQPSAndBurst(123, 456)
				if cfg == nil || cfg.QPS != float32(123) || cfg.Burst != 456 {
					os.Exit(2)
				}
				os.Exit(0)
			}

			tmpDir := GinkgoT().TempDir()
			kubeconfig := filepath.Join(tmpDir, "config")
			content := `apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: https://127.0.0.1:65535
    insecure-skip-tls-verify: true
users:
- name: local-user
  user:
    token: fake-token
contexts:
- name: local
  context:
    cluster: local
    user: local-user
current-context: local
`
			Expect(os.WriteFile(kubeconfig, []byte(content), 0o600)).To(Succeed())
			Expect(os.Setenv("KUBECONFIG", kubeconfig)).To(Succeed())

			cmd := exec.Command(os.Args[0], "-test.run=TestControllers", "-ginkgo.focus=sets qps and burst when both are positive")
			cmd.Env = append(os.Environ(), "FLUID_GET_CONFIG_SUBPROCESS=1", "KUBECONFIG="+kubeconfig)
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		})
	})
})
