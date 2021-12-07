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

package util

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetRuntimeNameFromFusePod(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name            string
		args            args
		wantRuntimeName string
		wantErr         bool
	}{
		{
			name: "test-right",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test-fuse-123"},
				},
			},
			wantRuntimeName: "test",
			wantErr:         false,
		},
		{
			name: "test-error",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
				},
			},
			wantRuntimeName: "",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRuntimeName, err := GetRuntimeNameFromFusePod(tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeNameFromFusePod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRuntimeName != tt.wantRuntimeName {
				t.Errorf("GetRuntimeNameFromFusePod() gotRuntimeName = %v, want %v", gotRuntimeName, tt.wantRuntimeName)
			}
		})
	}
}
