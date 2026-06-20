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

package databackup

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("OnceHandler", func() {
	var (
		testScheme     *runtime.Scheme
		mockDataBackup *v1alpha1.DataBackup
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		_ = v1alpha1.AddToScheme(testScheme)
		_ = corev1.AddToScheme(testScheme)

		mockDataBackup = &v1alpha1.DataBackup{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: v1alpha1.DataBackupSpec{},
		}
	})

	Describe("GetOperationStatus", func() {
		It("should return PhaseComplete when backup pod succeeded", func() {
			mockPod := corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, mockDataBackup, &mockPod)
			onceStatusHandler := &OnceHandler{dataBackup: mockDataBackup}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Client: client,
				Log:    fake.NullLogger(),
			}

			opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataBackup.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
		})

		It("should return PhaseFailed when backup pod failed", func() {
			mockFailedPod := corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, mockDataBackup, &mockFailedPod)
			onceStatusHandler := &OnceHandler{dataBackup: mockDataBackup}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Client: client,
				Log:    fake.NullLogger(),
			}

			opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataBackup.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseFailed))
		})

		It("should return unchanged status when backup pod is still running", func() {
			runningPod := corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, mockDataBackup, &runningPod)
			onceStatusHandler := &OnceHandler{dataBackup: mockDataBackup}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Client: client,
				Log:    fake.NullLogger(),
			}

			opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataBackup.Status)
			Expect(err).NotTo(HaveOccurred())
			// Pod is still running, so status should not change phase
			Expect(opStatus.Phase).To(Equal(mockDataBackup.Status.Phase))
		})

		It("should return unchanged status when backup pod does not exist", func() {
			// No pod in the fake client: GetPodByName returns (nil, nil) for not-found
			// IsFinishedPod(nil) returns false, so status is returned unchanged
			client := fake.NewFakeClientWithScheme(testScheme, mockDataBackup)
			onceStatusHandler := &OnceHandler{dataBackup: mockDataBackup}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Client: client,
				Log:    fake.NullLogger(),
			}

			opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataBackup.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(mockDataBackup.Status.Phase))
		})

		It("should use conditions LastTransitionTime when conditions are set on succeeded pod", func() {
			conditionTime := v1.Now()
			mockPodWithConditions := corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
					Conditions: []corev1.PodCondition{
						{
							LastTransitionTime: conditionTime,
						},
					},
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, mockDataBackup, &mockPodWithConditions)
			onceStatusHandler := &OnceHandler{dataBackup: mockDataBackup}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Client: client,
				Log:    fake.NullLogger(),
			}

			opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataBackup.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
			Expect(opStatus.Conditions).To(HaveLen(1))
			Expect(opStatus.Conditions[0].LastTransitionTime.Time).To(BeTemporally("~", conditionTime.Time, time.Second))
		})
	})
})
