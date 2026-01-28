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

package jindo

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestJindo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jindo Suite")
}

var _ = Describe("Jindo", func() {
	Describe("GetDefaultEngineImpl", func() {
		var originalEnv string

		BeforeEach(func() {
			// Save original environment variable
			originalEnv = os.Getenv(engineTypeFromEnv)
		})

		AfterEach(func() {
			// Restore original environment variable
			if originalEnv != "" {
				os.Setenv(engineTypeFromEnv, originalEnv)
			} else {
				os.Unsetenv(engineTypeFromEnv)
			}
		})

		Context("when environment variable is not set", func() {
			It("should return JindoCacheEngineImpl as default", func() {
				os.Unsetenv(engineTypeFromEnv)
				engine := GetDefaultEngineImpl()
				Expect(engine).To(Equal(common.JindoCacheEngineImpl))
			})
		})

		Context("when environment variable is set to JindoFSEngineImpl", func() {
			It("should return JindoFSEngineImpl", func() {
				os.Setenv(engineTypeFromEnv, common.JindoFSEngineImpl)
				engine := GetDefaultEngineImpl()
				Expect(engine).To(Equal(common.JindoFSEngineImpl))
			})
		})

		Context("when environment variable is set to JindoFSxEngineImpl", func() {
			It("should return JindoFSxEngineImpl", func() {
				os.Setenv(engineTypeFromEnv, common.JindoFSxEngineImpl)
				engine := GetDefaultEngineImpl()
				Expect(engine).To(Equal(common.JindoFSxEngineImpl))
			})
		})

		Context("when environment variable is set to an unknown value", func() {
			It("should return JindoCacheEngineImpl as default", func() {
				os.Setenv(engineTypeFromEnv, "unknown-engine")
				engine := GetDefaultEngineImpl()
				Expect(engine).To(Equal(common.JindoCacheEngineImpl))
			})
		})
	})

	Describe("GetRuntimeImage", func() {
		var originalEnv string

		BeforeEach(func() {
			// Save original environment variable
			originalEnv = os.Getenv(engineTypeFromEnv)
		})

		AfterEach(func() {
			// Restore original environment variable
			if originalEnv != "" {
				os.Setenv(engineTypeFromEnv, originalEnv)
			} else {
				os.Unsetenv(engineTypeFromEnv)
			}
		})

		Context("when engine is JindoFSxEngineImpl", func() {
			It("should return defaultJindofsxRuntimeImage", func() {
				os.Setenv(engineTypeFromEnv, common.JindoFSxEngineImpl)
				image := GetRuntimeImage()
				Expect(image).To(Equal(defaultJindofsxRuntimeImage))
				Expect(image).To(Equal("registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:4.6.8"))
			})
		})

		Context("when engine is JindoFSEngineImpl", func() {
			It("should return defaultJindofsRuntimeImage", func() {
				os.Setenv(engineTypeFromEnv, common.JindoFSEngineImpl)
				image := GetRuntimeImage()
				Expect(image).To(Equal(defaultJindofsRuntimeImage))
				Expect(image).To(Equal("registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.8.0"))
			})
		})

		Context("when engine is JindoCacheEngineImpl", func() {
			It("should return defaultJindoCacheRuntimeImage", func() {
				os.Setenv(engineTypeFromEnv, common.JindoCacheEngineImpl)
				image := GetRuntimeImage()
				Expect(image).To(Equal(defaultJindoCacheRuntimeImage))
				Expect(image).To(Equal("registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:6.2.0"))
			})
		})

		Context("when no environment variable is set", func() {
			It("should return defaultJindoCacheRuntimeImage as default", func() {
				os.Unsetenv(engineTypeFromEnv)
				image := GetRuntimeImage()
				Expect(image).To(Equal(defaultJindoCacheRuntimeImage))
				Expect(image).To(Equal("registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:6.2.0"))
			})
		})
	})

	Describe("Integration tests", func() {
		var originalEnv string

		BeforeEach(func() {
			originalEnv = os.Getenv(engineTypeFromEnv)
		})

		AfterEach(func() {
			if originalEnv != "" {
				os.Setenv(engineTypeFromEnv, originalEnv)
			} else {
				os.Unsetenv(engineTypeFromEnv)
			}
		})

		It("should return consistent engine type and image", func() {
			testCases := []struct {
				envValue      string
				expectedImage string
			}{
				{common.JindoFSEngineImpl, defaultJindofsRuntimeImage},
				{common.JindoFSxEngineImpl, defaultJindofsxRuntimeImage},
				{common.JindoCacheEngineImpl, defaultJindoCacheRuntimeImage},
				{"", defaultJindoCacheRuntimeImage},
			}

			for _, tc := range testCases {
				if tc.envValue != "" {
					os.Setenv(engineTypeFromEnv, tc.envValue)
				} else {
					os.Unsetenv(engineTypeFromEnv)
				}

				image := GetRuntimeImage()
				Expect(image).To(Equal(tc.expectedImage))
			}
		})
	})
})
