/*
Copyright 2024 The Fluid Authors.

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

package base

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestRuntimeInfoGetWorkerStatefulsetName(t *testing.T) {
	tests := []struct {
		name        string
		runtimeName string
		runtimeType string
		want        string
	}{
		{
			name:        "JindoRuntime uses jindofs suffix",
			runtimeName: "mydata",
			runtimeType: common.JindoRuntime,
			want:        "mydata-jindofs-worker",
		},
		{
			name:        "JindoCacheEngineImpl uses jindofs suffix",
			runtimeName: "cache",
			runtimeType: common.JindoCacheEngineImpl,
			want:        "cache-jindofs-worker",
		},
		{
			name:        "JindoFSxEngineImpl uses jindofs suffix",
			runtimeName: "fsx",
			runtimeType: common.JindoFSxEngineImpl,
			want:        "fsx-jindofs-worker",
		},
		{
			name:        "AlluxioRuntime uses default suffix",
			runtimeName: "alluxio-data",
			runtimeType: common.AlluxioRuntime,
			want:        "alluxio-data-worker",
		},
		{
			name:        "JuiceFSRuntime uses default suffix",
			runtimeName: "juice",
			runtimeType: common.JuiceFSRuntime,
			want:        "juice-worker",
		},
		{
			name:        "empty runtime type uses default suffix",
			runtimeName: "test",
			runtimeType: "",
			want:        "test-worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &RuntimeInfo{
				name:        tt.runtimeName,
				runtimeType: tt.runtimeType,
			}

			got := info.GetWorkerStatefulsetName()
			if got != tt.want {
				t.Errorf("GetWorkerStatefulsetName() = %v, want %v", got, tt.want)
			}
		})
	}
}
