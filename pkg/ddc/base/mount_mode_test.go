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
)

func TestParseMountModeSelectorFromStr(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantModes   []MountMode
		wantErr     bool
	}{
		{
			name:      "empty string returns empty selector",
			input:     "",
			wantModes: []MountMode{},
			wantErr:   false,
		},
		{
			name:      "All selects all modes",
			input:     "All",
			wantModes: SupportedMountModes,
			wantErr:   false,
		},
		{
			name:      "None returns empty selector",
			input:     "None",
			wantModes: []MountMode{},
			wantErr:   false,
		},
		{
			name:      "MountPod selects only MountPod",
			input:     "MountPod",
			wantModes: []MountMode{MountPodMountMode},
			wantErr:   false,
		},
		{
			name:      "Sidecar selects only Sidecar",
			input:     "Sidecar",
			wantModes: []MountMode{SidecarMountMode},
			wantErr:   false,
		},
		{
			name:      "comma separated selects multiple",
			input:     "MountPod,Sidecar",
			wantModes: []MountMode{MountPodMountMode, SidecarMountMode},
			wantErr:   false,
		},
		{
			name:      "unsupported mode returns error",
			input:     "InvalidMode",
			wantModes: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMountModeSelectorFromStr(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMountModeSelectorFromStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for _, mode := range tt.wantModes {
				if !got.Selected(mode) {
					t.Errorf("ParseMountModeSelectorFromStr() missing mode %v", mode)
				}
			}

			if len(got) != len(tt.wantModes) {
				t.Errorf("ParseMountModeSelectorFromStr() got %d modes, want %d", len(got), len(tt.wantModes))
			}
		})
	}
}

func TestMountModeSelectorSelected(t *testing.T) {
	selector := mountModeSelector{MountPodMountMode: true}

	if !selector.Selected(MountPodMountMode) {
		t.Error("Selected() should return true for existing mode")
	}

	if selector.Selected(SidecarMountMode) {
		t.Error("Selected() should return false for non-existing mode")
	}
}
