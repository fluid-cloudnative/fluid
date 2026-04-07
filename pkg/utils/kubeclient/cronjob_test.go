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

package kubeclient

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("GetCronJobStatus", func() {
	var (
		nowTime           time.Time
		testDate          metav1.Time
		namespace         string
		testCronJobInputs []*batchv1.CronJob
		testCronJobs      []runtime.Object
		client            client.Client
		patch             *gomonkey.Patches
	)

	BeforeEach(func() {
		nowTime = time.Now()
		testDate = metav1.NewTime(time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), 0, 0, 0, nowTime.Location()))
		namespace = "default"

		testCronJobInputs = []*batchv1.CronJob{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: namespace,
				},
				Status: batchv1.CronJobStatus{
					LastScheduleTime: &testDate,
				},
			},
		}

		testCronJobs = []runtime.Object{}
		for _, cj := range testCronJobInputs {
			testCronJobs = append(testCronJobs, cj.DeepCopy())
		}

		_ = batchv1beta1.AddToScheme(testScheme)
		client = fake.NewFakeClientWithScheme(testScheme, testCronJobs...)

		// Apply gomonkey patch
		patch = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
			return true
		})
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Context("when CronJob exists", func() {
		It("should return the CronJob status successfully", func() {
			key := types.NamespacedName{
				Namespace: namespace,
				Name:      "test1",
			}

			got, err := GetCronJobStatus(client, key)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.LastScheduleTime).To(Equal(&testDate))
		})
	})

	Context("when CronJob does not exist", func() {
		It("should return an error", func() {
			key := types.NamespacedName{
				Namespace: namespace,
				Name:      "test-notexist",
			}

			got, err := GetCronJobStatus(client, key)

			Expect(err).To(HaveOccurred())
			Expect(got).To(BeNil())
		})
	})

	Context("when batchv1 CronJob is not supported and batchv1beta1 CronJob exists", func() {
		It("should return converted status from batchv1beta1", func() {
			key := types.NamespacedName{
				Namespace: namespace,
				Name:      "test-beta",
			}
			betaCronJob := &batchv1beta1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-beta",
					Namespace: namespace,
				},
				Status: batchv1beta1.CronJobStatus{
					Active: []corev1.ObjectReference{
						{Name: "pod-0"},
					},
					LastScheduleTime:   &testDate,
					LastSuccessfulTime: &testDate,
				},
			}
			betaClient := fake.NewFakeClientWithScheme(testScheme, betaCronJob.DeepCopy())

			patch.Reset()
			patch = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
				return false
			})

			got, err := GetCronJobStatus(betaClient, key)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.LastScheduleTime).To(Equal(&testDate))
			Expect(got.LastSuccessfulTime).To(Equal(&testDate))
			Expect(got.Active).To(HaveLen(1))
			Expect(got.Active[0].Name).To(Equal("pod-0"))
		})
	})

	Context("when batchv1 CronJob is not supported and batchv1beta1 CronJob does not exist", func() {
		It("should return an error", func() {
			key := types.NamespacedName{
				Namespace: namespace,
				Name:      "test-beta-notexist",
			}
			emptyClient := fake.NewFakeClientWithScheme(testScheme)

			patch.Reset()
			patch = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
				return false
			})

			got, err := GetCronJobStatus(emptyClient, key)

			Expect(err).To(HaveOccurred())
			Expect(got).To(BeNil())
		})
	})
})
