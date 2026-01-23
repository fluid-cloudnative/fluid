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

package juicefs

import (
	"errors"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("UfsInternal", func() {
	var (
		engine  *JuiceFSEngine
		patches *gomonkey.Patches
	)

	BeforeEach(func() {
		engine = &JuiceFSEngine{
			name:      "test",
			namespace: "fluid",
			Log:       fake.NullLogger(),
		}
	})

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	Describe("totalStorageBytesInternal", func() {
		Context("when GetRunningPodsOfStatefulSet returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return nil, errors.New("failed to get pods")
					})
			})

			It("should return error and zero bytes", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).To(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})

		Context("when GetRunningPodsOfStatefulSet returns empty pods", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{}, nil
					})
			})

			It("should return zero bytes without error", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 0, errors.New("failed to get used space")
					})
			})

			It("should return error and zero bytes", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).To(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 1024, nil
					})
			})

			It("should return correct total bytes", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(total).To(Equal(int64(1024)))
			})
		})
	})

	Describe("totalFileNumsInternal", func() {
		Context("when GetRunningPodsOfStatefulSet returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return nil, errors.New("failed to get pods")
					})
			})

			It("should return error and zero file count", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).To(HaveOccurred())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when GetRunningPodsOfStatefulSet returns empty pods", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{}, nil
					})
			})

			It("should return zero file count without error", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when GetFileCount returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetFileCount",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 0, errors.New("failed to get file count")
					})
			})

			It("should return error and zero file count", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).To(HaveOccurred())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when GetFileCount succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetFileCount",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 100, nil
					})
			})

			It("should return correct file count", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(fileCount).To(Equal(int64(100)))
			})
		})
	})

	Describe("usedSpaceInternal", func() {
		Context("when GetRunningPodsOfStatefulSet returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return nil, errors.New("failed to get pods")
					})
			})

			It("should return error and zero used space", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).To(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when GetRunningPodsOfStatefulSet returns empty pods", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{}, nil
					})
			})

			It("should return zero used space without error", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 0, errors.New("failed to get used space")
					})
			})

			It("should return error and zero used space", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).To(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 2048, nil
					})
			})

			It("should return correct used space", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(2048)))
			})
		})
	})
})

var _ = Describe("UfsInternal", func() {
	var (
		engine  *JuiceFSEngine
		patches *gomonkey.Patches
	)

	BeforeEach(func() {
		engine = &JuiceFSEngine{
			name:      "test",
			namespace: "fluid",
			Log:       fake.NullLogger(),
		}
	})

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	Describe("totalStorageBytesInternal", func() {
		Context("when GetRunningPodsOfStatefulSet returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return nil, errors.New("failed to get pods")
					})
			})

			It("should return error and zero bytes", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).To(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})

		Context("when GetRunningPodsOfStatefulSet returns empty pods", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{}, nil
					})
			})

			It("should return zero bytes without error", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 0, errors.New("failed to get used space")
					})
			})

			It("should return error and zero bytes", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).To(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 1024, nil
					})
			})

			It("should return correct total bytes", func() {
				total, err := engine.totalStorageBytesInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(total).To(Equal(int64(1024)))
			})
		})
	})

	Describe("totalFileNumsInternal", func() {
		Context("when GetRunningPodsOfStatefulSet returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return nil, errors.New("failed to get pods")
					})
			})

			It("should return error and zero file count", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).To(HaveOccurred())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when GetRunningPodsOfStatefulSet returns empty pods", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{}, nil
					})
			})

			It("should return zero file count without error", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when GetFileCount returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetFileCount",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 0, errors.New("failed to get file count")
					})
			})

			It("should return error and zero file count", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).To(HaveOccurred())
				Expect(fileCount).To(Equal(int64(0)))
			})
		})

		Context("when GetFileCount succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetFileCount",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 100, nil
					})
			})

			It("should return correct file count", func() {
				fileCount, err := engine.totalFileNumsInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(fileCount).To(Equal(int64(100)))
			})
		})
	})

	Describe("usedSpaceInternal", func() {
		Context("when GetRunningPodsOfStatefulSet returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return nil, errors.New("failed to get pods")
					})
			})

			It("should return error and zero used space", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).To(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when GetRunningPodsOfStatefulSet returns empty pods", func() {
			BeforeEach(func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{}, nil
					})
			})

			It("should return zero used space without error", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace returns error", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 0, errors.New("failed to get used space")
					})
			})

			It("should return error and zero used space", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).To(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(0)))
			})
		})

		Context("when GetUsedSpace succeeds", func() {
			BeforeEach(func() {
				patches = gomonkey.NewPatches()
				patches.ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
					func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
						return []corev1.Pod{{
							ObjectMeta: metav1.ObjectMeta{Name: "test-worker-0"},
						}}, nil
					})

				var fileUtils operations.JuiceFileUtils
				patches.ApplyMethod(reflect.TypeOf(fileUtils), "GetUsedSpace",
					func(_ operations.JuiceFileUtils, path string) (int64, error) {
						return 2048, nil
					})
			})

			It("should return correct used space", func() {
				usedSpace, err := engine.usedSpaceInternal()
				Expect(err).NotTo(HaveOccurred())
				Expect(usedSpace).To(Equal(int64(2048)))
			})
		})
	})
})
