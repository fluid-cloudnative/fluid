/*
  Copyright 2022 The Fluid Authors.

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

package watch

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

var _ = Describe("podEventHandler", func() {
	Describe("onDeleteFunc", func() {
		It("should return false when Object is not a pod", func() {
			delPodEvent := event.DeleteEvent{
				Object: &appsv1.DaemonSet{},
			}
			handler := &podEventHandler{}
			f := handler.onDeleteFunc(&FakePodReconciler{})
			predicate := f(delPodEvent)
			Expect(predicate).To(BeFalse())
		})

		It("should return false when Object is an empty pod", func() {
			delPodEvent := event.DeleteEvent{
				Object: &corev1.Pod{},
			}
			handler := &podEventHandler{}
			f := handler.onDeleteFunc(&FakePodReconciler{})
			predicate := f(delPodEvent)
			Expect(predicate).To(BeFalse())
		})
	})

	Describe("onCreateFunc", func() {
		It("should return false when Object is not a pod", func() {
			createPodEvent := event.CreateEvent{
				Object: &appsv1.DaemonSet{},
			}
			handler := &podEventHandler{}
			f := handler.onCreateFunc(&FakePodReconciler{})
			predicate := f(createPodEvent)
			Expect(predicate).To(BeFalse())
		})

		It("should return false when Object is an empty pod", func() {
			createPodEvent := event.CreateEvent{
				Object: &corev1.Pod{},
			}
			handler := &podEventHandler{}
			f := handler.onCreateFunc(&FakePodReconciler{})
			predicate := f(createPodEvent)
			Expect(predicate).To(BeFalse())
		})
	})

	Describe("onUpdateFunc", func() {
		var updatePodEvent event.UpdateEvent

		BeforeEach(func() {
			updatePodEvent = event.UpdateEvent{
				ObjectOld: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "123",
						Name:            "test",
						Labels: map[string]string{
							common.InjectServerless:  common.True,
							common.InjectSidecarDone: common.True,
						},
					},
					Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: common.FuseContainerName + "-0"}}},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "app",
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{
										StartedAt: metav1.Time{Time: time.Now()},
									},
								},
							},
							{
								Name: common.FuseContainerName + "-0",
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{
										StartedAt: metav1.Time{Time: time.Now()},
									},
								},
							},
						}},
				},
				ObjectNew: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "456",
						Name:            "test",
						Labels: map[string]string{
							common.InjectServerless:  common.True,
							common.InjectSidecarDone: common.True,
						},
					},
					Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: common.FuseContainerName + "-0"}}},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "app",
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										StartedAt: metav1.Time{Time: time.Now()},
										ExitCode:  0,
									},
								},
							},
							{
								Name: common.FuseContainerName + "-0",
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{
										StartedAt: metav1.Time{Time: time.Now()},
									},
								},
							},
						}},
				},
			}
		})

		It("should return true when updateEvent is validated", func() {
			handler := &podEventHandler{}
			f := handler.onUpdateFunc(&FakePodReconciler{})
			predicate := f(updatePodEvent)
			Expect(predicate).To(BeTrue())
		})

		It("should return false when resource version is equal", func() {
			updatePodEvent.ObjectOld.SetResourceVersion("456")
			handler := &podEventHandler{}
			f := handler.onUpdateFunc(&FakePodReconciler{})
			predicate := f(updatePodEvent)
			Expect(predicate).To(BeFalse())
		})

		It("should return false when object is not kind of runtimeInterface", func() {
			updatePodEvent.ObjectOld = &corev1.Pod{}
			updatePodEvent.ObjectNew = &corev1.Pod{}
			handler := &podEventHandler{}
			f := handler.onUpdateFunc(&FakePodReconciler{})
			predicate := f(updatePodEvent)
			Expect(predicate).To(BeFalse())
		})

		It("should return false when new Object is not kind of runtimeInterface", func() {
			updatePodEvent.ObjectNew = &appsv1.DaemonSet{}
			handler := &podEventHandler{}
			f := handler.onUpdateFunc(&FakePodReconciler{})
			predicate := f(updatePodEvent)
			Expect(predicate).To(BeFalse())
		})
	})
})

var _ = Describe("ShouldInQueue", func() {
	DescribeTable("should handle various pod scenarios",
		func(pod *corev1.Pod, want bool) {
			got := ShouldInQueue(pod)
			Expect(got).To(Equal(want))
		},
		Entry("no-fuse-label",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			false,
		),
		Entry("restartAlways",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{RestartPolicy: corev1.RestartPolicyAlways},
			},
			false,
		),
		Entry("no-fuse",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
			},
			false,
		),
		Entry("app-cn-not-exit",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: common.FuseContainerName + "-0"}}},
				Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: "app",
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{
								StartedAt: metav1.Time{Time: time.Now()},
							},
						},
					},
					{
						Name: common.FuseContainerName + "-0",
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{
								StartedAt: metav1.Time{Time: time.Now()},
							},
						},
					},
				}},
			},
			false,
		),
		Entry("fuse-cn-exit",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: common.FuseContainerName + "-0"}}},
				Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: "app",
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								StartedAt: metav1.Time{Time: time.Now()},
								ExitCode:  0,
							},
						},
					},
					{
						Name: common.FuseContainerName + "-0",
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								StartedAt: metav1.Time{Time: time.Now()},
								ExitCode:  0,
							},
						},
					},
				}},
			},
			false,
		),
		Entry("fuse-cn-no-exit",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: common.FuseContainerName + "-0"}}},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "app",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									StartedAt: metav1.Time{Time: time.Now()},
									ExitCode:  0,
								},
							},
						},
						{
							Name: common.FuseContainerName + "-0",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.Time{Time: time.Now()},
								},
							},
						},
					}},
			},
			true,
		),
		Entry("multi-cn-exit",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: "app2"}, {Name: common.FuseContainerName + "-0"}}},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "app",
							State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
								StartedAt: metav1.Time{Time: time.Now()},
								ExitCode:  0,
							}},
						},
						{
							Name: "app2",
							State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
								StartedAt: metav1.Time{Time: time.Now()},
								ExitCode:  0,
							}},
						},
						{
							Name: common.FuseContainerName + "-0",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.Time{Time: time.Now()},
								},
							},
						}}},
			},
			true,
		),
		Entry("multi-cn-not-exit",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: "app2"}, {Name: common.FuseContainerName + "-0"}}},
				Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
					{
						Name: "app",
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								StartedAt: metav1.Time{Time: time.Now()},
								ExitCode:  0,
							},
						},
					},
					{
						Name: "app2",
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{
								StartedAt: metav1.Time{Time: time.Now()},
							},
						},
					},
					{
						Name: common.FuseContainerName + "-0",
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{
								StartedAt: metav1.Time{Time: time.Now()},
							},
						},
					}}},
			},
			false,
		),
		Entry("pod-pending",
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: common.FuseContainerName + "-0"}}},
				Status: corev1.PodStatus{
					Phase:             corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{}},
			},
			false,
		),
	)
})
