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

package thin

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func TestThinEngine_SyncRuntime(t1 *testing.T) {
	type fields struct{}
	type args struct {
		ctx runtime.ReconcileRequestContext
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantChanged bool
		wantErr     bool
	}{
		{
			name:        "default",
			wantChanged: false,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := ThinEngine{}
			gotChanged, err := t.SyncRuntime(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t1.Errorf("SyncRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t1.Errorf("SyncRuntime() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
		})
	}
}
