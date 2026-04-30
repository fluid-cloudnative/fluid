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

package v1alpha1

import "testing"

func TestAlluxioRuntime_Replicas(t *testing.T) {
	testCases := map[string]struct {
		runtime *AlluxioRuntime
		want    int32
	}{
		"nil runtime defaults to zero": {
			runtime: nil,
			want:    0,
		},
		"returns configured replicas": {
			runtime: &AlluxioRuntime{
				Spec: AlluxioRuntimeSpec{
					Replicas: 3,
				},
			},
			want: 3,
		},
		"returns zero replicas when not configured": {
			runtime: &AlluxioRuntime{},
			want:    0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.runtime.Replicas()
			if got != tc.want {
				t.Fatalf("Replicas() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestAlluxioRuntime_GetStatus(t *testing.T) {
	t.Run("nil runtime returns empty status pointer", func(t *testing.T) {
		var runtime *AlluxioRuntime
		got := runtime.GetStatus()
		if got == nil {
			t.Fatalf("GetStatus() returned nil")
		}
		if got.MasterPhase != RuntimePhaseNone {
			t.Fatalf("GetStatus().MasterPhase = %q, want %q", got.MasterPhase, RuntimePhaseNone)
		}
	})

	t.Run("returns pointer to runtime status", func(t *testing.T) {
		runtime := &AlluxioRuntime{
			Status: RuntimeStatus{
				MasterPhase: RuntimePhaseReady,
				WorkerPhase: RuntimePhasePartialReady,
			},
		}

		got := runtime.GetStatus()
		if got == nil {
			t.Fatalf("GetStatus() returned nil")
		}
		if got.MasterPhase != RuntimePhaseReady {
			t.Fatalf("GetStatus().MasterPhase = %q, want %q", got.MasterPhase, RuntimePhaseReady)
		}

		got.FusePhase = RuntimePhaseNotReady
		if runtime.Status.FusePhase != RuntimePhaseNotReady {
			t.Fatalf("GetStatus() should return underlying status pointer")
		}
	})
}
