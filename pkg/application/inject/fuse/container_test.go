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

package fuse

import (
	"k8s.io/utils/ptr"
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	corev1 "k8s.io/api/core/v1"
)

func Test_findInjectedSidecars(t *testing.T) {

	pod1 := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test",
				},
				{
					Name: "test2",
				},
			},
		},
	}
	podObjs1, err := pod.NewApplication(pod1).GetPodSpecs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pod2 := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "fluid-fuse-0",
				},
				{
					Name: "test",
				},
			},
		},
	}
	podObjs2, err := pod.NewApplication(pod2).GetPodSpecs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type args struct {
		pod common.FluidObject
	}
	tests := []struct {
		name                 string
		args                 args
		wantInjectedSidecars []corev1.Container
		wantErr              bool
	}{
		{
			name: "no_injected_sidecars",
			args: args{
				pod: podObjs1[0],
			},
			wantInjectedSidecars: []corev1.Container{},
			wantErr:              false,
		},
		{
			name: "one_injected_sidecar",
			args: args{
				pod: podObjs2[0],
			},
			wantInjectedSidecars: []corev1.Container{
				{
					Name: "fluid-fuse-0",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInjectedSidecars, err := findInjectedSidecars(tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("findInjectedSidecars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInjectedSidecars, tt.wantInjectedSidecars) {
				t.Errorf("findInjectedSidecars() = %v, want %v", gotInjectedSidecars, tt.wantInjectedSidecars)
			}
		})
	}
}
