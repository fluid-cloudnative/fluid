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
package docker

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var testScheme *runtime.Scheme

var _ = BeforeSuite(func() {
	testScheme = runtime.NewScheme()
	Expect(v1.AddToScheme(testScheme)).To(Succeed())
})

var _ = Describe("Docker", func() {
	Describe("ParseDockerImage", func() {
		DescribeTable("parsing docker images",
			func(input, expectedImage, expectedTag string) {
				image, tag := ParseDockerImage(input)
				Expect(image).To(Equal(expectedImage))
				Expect(tag).To(Equal(expectedTag))
			},
			Entry("with explicit tag", "test:abc", "test", "abc"),
			Entry("without tag defaults to latest", "test", "test", "latest"),
			Entry("with registry port", "test:35000/test:abc", "test:35000/test", "abc"),
			Entry("with multiple colons in image name", "registry:5000/namespace:subnamespace/image:v1.0", "registry:5000/namespace:subnamespace/image", "v1.0"),
			Entry("empty string", "", "", "latest"),
		)
	})

	Describe("GetImageRepoFromEnv", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
			GinkgoT().Setenv("ALLUXIO_IMAGE_ENV", "alluxio")
		})

		DescribeTable("getting image repo from environment",
			func(envName, expected string) {
				result := GetImageRepoFromEnv(envName)
				Expect(result).To(Equal(expected))
			},
			Entry("valid env with tag", "FLUID_IMAGE_ENV", "fluid"),
			Entry("non-existent env", "NOT EXIST", ""),
			Entry("env without tag", "ALLUXIO_IMAGE_ENV", ""),
		)
	})

	Describe("GetImageTagFromEnv", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
			GinkgoT().Setenv("ALLUXIO_IMAGE_ENV", "alluxio")
		})

		DescribeTable("getting image tag from environment",
			func(envName, expected string) {
				result := GetImageTagFromEnv(envName)
				Expect(result).To(Equal(expected))
			},
			Entry("valid env with tag", "FLUID_IMAGE_ENV", "0.6.0"),
			Entry("non-existent env", "NOT EXIST", ""),
			Entry("env without tag", "ALLUXIO_IMAGE_ENV", ""),
		)
	})

	Describe("GetImagePullSecretsFromEnv", func() {
		DescribeTable("getting image pull secrets from environment",
			func(envValue string, expected []v1.LocalObjectReference) {
				GinkgoT().Setenv(common.EnvImagePullSecretsKey, envValue)
				result := GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
				Expect(result).To(Equal(expected))
			},
			Entry("multiple secrets",
				"test1,test2",
				[]v1.LocalObjectReference{
					{Name: "test1"},
					{Name: "test2"},
				}),
			Entry("empty value",
				"",
				[]v1.LocalObjectReference{}),
			Entry("single secret",
				"str1",
				[]v1.LocalObjectReference{{Name: "str1"}}),
			Entry("trailing comma",
				"str1,",
				[]v1.LocalObjectReference{{Name: "str1"}}),
			Entry("leading commas and trailing comma",
				",,,str1,",
				[]v1.LocalObjectReference{{Name: "str1"}}),
			Entry("multiple secrets with extra commas",
				",,,str1,,,str2,,",
				[]v1.LocalObjectReference{{Name: "str1"}, {Name: "str2"}}),
		)

		It("should return empty slice when environment variable does not exist", func() {
			result := GetImagePullSecretsFromEnv("NON_EXISTENT_ENV_VAR")
			Expect(result).To(BeEmpty())
		})
	})

	Describe("ParseInitImage", func() {
		Context("with valid environment variable", func() {
			BeforeEach(func() {
				GinkgoT().Setenv("FLUID_IMAGE_ENV", "fluid:0.6.0")
			})

			DescribeTable("parsing init image parameters",
				func(image, tag, imagePullPolicy, envName, wantImage, wantTag, wantImagePullPolicy string) {
					resultImage, resultTag, resultImagePullPolicy := ParseInitImage(image, tag, imagePullPolicy, envName)
					Expect(resultImage).To(Equal(wantImage))
					Expect(resultTag).To(Equal(wantTag))
					Expect(resultImagePullPolicy).To(Equal(wantImagePullPolicy))
				},
				Entry("all parameters provided with default policy",
					"fluid", "0.6.0", "", "FLUID_IMAGE_ENV",
					"fluid", "0.6.0", common.DefaultImagePullPolicy),
				Entry("empty image with Always policy",
					"", "0.6.0", "Always", "FLUID_IMAGE_ENV",
					"fluid", "0.6.0", "Always"),
				Entry("empty tag with Always policy",
					"fluid", "", "Always", "FLUID_IMAGE_ENV",
					"fluid", "0.6.0", "Always"),
				Entry("all parameters provided with Always policy",
					"fluid", "0.6.0", "Always", "FLUID_IMAGE_ENV",
					"fluid", "0.6.0", "Always"),
			)
		})

		Context("without environment variable - fallback to default", func() {
			It("should use default init image when env is not set and params are empty", func() {
				// Parse the default init image to get expected values
				defaultImageParts := strings.Split(common.DefaultInitImage, ":")
				expectedImage := defaultImageParts[0]
				expectedTag := ""
				if len(defaultImageParts) >= 2 {
					expectedTag = defaultImageParts[1]
				}

				resultImage, resultTag, resultImagePullPolicy := ParseInitImage("", "", "", "NON_EXISTENT_ENV")
				Expect(resultImage).To(Equal(expectedImage))
				Expect(resultTag).To(Equal(expectedTag))
				Expect(resultImagePullPolicy).To(Equal(common.DefaultImagePullPolicy))
			})

			It("should use default tag when env is not set and tag is empty", func() {
				defaultImageParts := strings.Split(common.DefaultInitImage, ":")
				expectedTag := ""
				if len(defaultImageParts) >= 2 {
					expectedTag = defaultImageParts[1]
				}

				resultImage, resultTag, resultImagePullPolicy := ParseInitImage("custom-image", "", "IfNotPresent", "NON_EXISTENT_ENV")
				Expect(resultImage).To(Equal("custom-image"))
				Expect(resultTag).To(Equal(expectedTag))
				Expect(resultImagePullPolicy).To(Equal("IfNotPresent"))
			})
		})
	})

	Describe("GetWorkerImage", func() {
		var (
			testClient      client.Client
			configMapInputs []*v1.ConfigMap
		)

		BeforeEach(func() {
			configMapInputs = []*v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "hbase-alluxio-values", Namespace: "default"},
					Data: map[string]string{
						"data": "image: fluid\nimageTag: 0.6.0",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "spark-alluxio-values", Namespace: "default"},
					Data: map[string]string{
						"test-data": "image: fluid\n imageTag: 0.6.0",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "hadoop-alluxio-values", Namespace: "default"},
					Data: map[string]string{
						"data": "image: hadoop-image\nimageTag: 2.0.0\nextra: line",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "kafka-alluxio-values", Namespace: "default"},
					Data: map[string]string{
						"data": "imageTag: 1.0.0",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "redis-alluxio-values", Namespace: "default"},
					Data: map[string]string{
						"data": "image: redis-image",
					},
				},
			}

			testConfigMaps := []runtime.Object{}
			for _, cm := range configMapInputs {
				testConfigMaps = append(testConfigMaps, cm.DeepCopy())
			}

			testClient = fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)
		})

		DescribeTable("getting worker image from ConfigMap",
			func(datasetName, runtimeType, namespace, wantImageName, wantImageTag string) {
				resultImageName, resultImageTag := GetWorkerImage(testClient, datasetName, runtimeType, namespace)
				Expect(resultImageName).To(Equal(wantImageName))
				Expect(resultImageTag).To(Equal(wantImageTag))
			},
			Entry("jindoruntime in different namespace",
				"hbase", "jindoruntime", "fluid",
				"", ""),
			Entry("alluxio with valid ConfigMap",
				"hbase", "alluxio", "default",
				"fluid", "0.6.0"),
			Entry("alluxio with invalid ConfigMap data key",
				"spark", "alluxio", "default",
				"", ""),
			Entry("alluxio with both image and imageTag",
				"hadoop", "alluxio", "default",
				"hadoop-image", "2.0.0"),
			Entry("alluxio with only imageTag",
				"kafka", "alluxio", "default",
				"", "1.0.0"),
			Entry("alluxio with only image",
				"redis", "alluxio", "default",
				"redis-image", ""),
			Entry("non-existent ConfigMap",
				"nonexistent", "alluxio", "default",
				"", ""),
		)
	})
})
