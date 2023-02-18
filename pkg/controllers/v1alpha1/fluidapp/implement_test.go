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

package fluidapp

import (
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func TestFluidAppReconcilerImplement_umountFuseSidecars(t *testing.T) {
	mockExec := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
		return "", "", nil
	}
	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	defer wrappedUnhook()

	type fields struct {
		Client   client.Client
		Log      logr.Logger
		Recorder record.EventRecorder
	}
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test-no-fuse",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "test"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test-no-mountpath",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: common.FuseContainerName + "-0"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test-prestop",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: common.FuseContainerName + "-0",
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{Command: []string{"umount"}},
								},
							},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test-mountpath",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: common.FuseContainerName + "-0",
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "juicefs-fuse-mount",
								MountPath: "/mnt/jfs",
							}},
						}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test-multi-sidecar",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: common.FuseContainerName + "-0",
								VolumeMounts: []corev1.VolumeMount{{
									Name:      "juicefs-fuse-mount",
									MountPath: "/mnt/jfs",
								}},
							},
							{
								Name: common.FuseContainerName + "-1",
								VolumeMounts: []corev1.VolumeMount{{
									Name:      "juicefs-fuse-mount",
									MountPath: "/mnt/jfs",
								}},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &FluidAppReconcilerImplement{
				Log: fake.NullLogger(),
			}
			if err := i.umountFuseSidecars(tt.args.pod); (err != nil) != tt.wantErr {
				t.Errorf("umountFuseSidecar() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
