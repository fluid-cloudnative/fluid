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
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func Test_podEventHandler_onDeleteFunc(t *testing.T) {
	// 1. the Object is not pod
	delPodEvent := event.DeleteEvent{
		Object: &appsv1.DaemonSet{},
	}
	podEventHandler := &podEventHandler{}

	f := podEventHandler.onDeleteFunc(&FakePodReconciler{})
	predicate := f(delPodEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", delPodEvent)
	}

	// 2. the Object is pod
	delPodEvent.Object = &corev1.Pod{}
	predicate = f(delPodEvent)
	if predicate {
		t.Errorf("The event %v should not be reconciled, but pass.", delPodEvent)
	}
}

func Test_podEventHandler_onCreateFunc(t *testing.T) {
	// 1. the Object is not pod
	delPodEvent := event.CreateEvent{
		Object: &appsv1.DaemonSet{},
	}
	podEventHandler := &podEventHandler{}

	f := podEventHandler.onCreateFunc(&FakePodReconciler{})
	predicate := f(delPodEvent)

	if predicate {
		t.Errorf("The event %v should not be reconciled, but skip.", delPodEvent)
	}

	// 2. the Object is pod
	delPodEvent.Object = &corev1.Pod{}
	predicate = f(delPodEvent)
	if predicate {
		t.Errorf("The event %v should not be reconciled, but pass.", delPodEvent)
	}
}

func Test_podEventHandler_onUpdateFunc(t *testing.T) {
	updatePodEvent := event.UpdateEvent{
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
	podEventHandler := &podEventHandler{}

	f := podEventHandler.onUpdateFunc(&FakePodReconciler{})
	predicate := f(updatePodEvent)

	// 1. expect the updateEvent is validated
	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", updatePodEvent)
	}

	// 2. expect the updateEvent is not validated due to the resource version is equal
	updatePodEvent.ObjectOld.SetResourceVersion("456")
	predicate = f(updatePodEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updatePodEvent)
	}

	// 3. expect the updateEvent is not validated due to the object is not kind of runtimeInterface
	updatePodEvent.ObjectOld = &corev1.Pod{}
	updatePodEvent.ObjectNew = &corev1.Pod{}
	predicate = f(updatePodEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updatePodEvent)
	}

	// 4. expect the updateEvent is not validate due the old Object  is not kind of the runtimeInterface
	updatePodEvent.ObjectNew = &appsv1.DaemonSet{}
	predicate = f(updatePodEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updatePodEvent)
	}
}

func Test_shouldRequeue(t *testing.T) {
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no-fuse-label",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
				},
			},
			want: false,
		},
		{
			name: "restartAlways",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Labels: map[string]string{
							common.InjectServerless:  common.True,
							common.InjectSidecarDone: common.True,
						},
					},
					Spec: corev1.PodSpec{RestartPolicy: corev1.RestartPolicyAlways},
				},
			},
			want: false,
		},
		{
			name: "no-fuse",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
						Labels: map[string]string{
							common.InjectServerless:  common.True,
							common.InjectSidecarDone: common.True,
						},
					},
					Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
				},
			},
			want: false,
		},
		{
			name: "app-cn-not-exit",
			args: args{
				pod: &corev1.Pod{
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
			},
			want: false,
		},
		{
			name: "fuse-cn-exit",
			args: args{
				pod: &corev1.Pod{
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
			},
			want: false,
		},
		{
			name: "fuse-cn-no-exit",
			args: args{
				pod: &corev1.Pod{
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
			},
			want: true,
		},
		{
			name: "multi-cn-exit",
			args: args{
				pod: &corev1.Pod{
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
			},
			want: true,
		},
		{
			name: "multi-cn-not-exit",
			args: args{
				pod: &corev1.Pod{
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
			},
			want: false,
		},
		{
			name: "pod-pending",
			args: args{
				pod: &corev1.Pod{
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
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldInQueue(tt.args.pod); got != tt.want {
				t.Errorf("shouldReconcile() = %v, want %v", got, tt.want)
			}
		})
	}
}
