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
	"errors"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("InitClient", func() {
	var (
		pathExistPatch            *gomonkey.Patches
		buildConfigFromFlagsPatch *gomonkey.Patches
		newForConfigPatch         *gomonkey.Patches
	)

	BeforeEach(func() {
		// Set environment variable
		GinkgoT().Setenv(common.RecommendedKubeConfigPathEnv, "Path for test")

		// Reset global variables
		restConfig = nil
		clientset = nil
	})

	AfterEach(func() {
		// Reset all patches
		if pathExistPatch != nil {
			pathExistPatch.Reset()
			pathExistPatch = nil
		}
		if buildConfigFromFlagsPatch != nil {
			buildConfigFromFlagsPatch.Reset()
			buildConfigFromFlagsPatch = nil
		}
		if newForConfigPatch != nil {
			newForConfigPatch.Reset()
			newForConfigPatch = nil
		}
	})

	Context("when kubeconfig path exists", func() {
		BeforeEach(func() {
			pathExistPatch = gomonkey.ApplyFunc(utils.PathExists, func(path string) bool {
				return true
			})
		})

		Context("when BuildConfigFromFlags fails", func() {
			BeforeEach(func() {
				buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
					return nil, errors.New("fail to run the function")
				})
			})

			It("should return an error", func() {
				err := initClient()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when BuildConfigFromFlags succeeds but NewForConfig fails", func() {
			BeforeEach(func() {
				buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
					return nil, nil
				})
				newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
					return nil, errors.New("fail to run the function")
				})
			})

			It("should return an error", func() {
				err := initClient()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when both BuildConfigFromFlags and NewForConfig succeed", func() {
			BeforeEach(func() {
				buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
					return nil, nil
				})
				newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
					return nil, nil
				})
			})

			It("should succeed without error", func() {
				err := initClient()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when kubeconfig path does not exist", func() {
		BeforeEach(func() {
			pathExistPatch = gomonkey.ApplyFunc(utils.PathExists, func(path string) bool {
				return false
			})
		})

		Context("when BuildConfigFromFlags fails", func() {
			BeforeEach(func() {
				buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
					return nil, errors.New("fail to run the function")
				})
			})

			It("should return an error", func() {
				err := initClient()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when BuildConfigFromFlags succeeds but NewForConfig fails", func() {
			BeforeEach(func() {
				buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
					return nil, nil
				})
				newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
					return nil, errors.New("fail to run the function")
				})
			})

			It("should return an error", func() {
				err := initClient()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when both BuildConfigFromFlags and NewForConfig succeed", func() {
			BeforeEach(func() {
				buildConfigFromFlagsPatch = gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
					return nil, nil
				})
				newForConfigPatch = gomonkey.ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
					return nil, nil
				})
			})

			It("should succeed without error", func() {
				err := initClient()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
