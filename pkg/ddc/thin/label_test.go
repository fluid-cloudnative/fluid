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
)

func TestThinEngine_getFuseLabelName(t1 *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "fuse",
			fields: fields{
				name:      "fuse1",
				namespace: "fluid",
			},
			want: "fluid.io/f-fluid-fuse1",
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			out := t.getFuseLabelName()
			if out != tt.want {
				t1.Errorf("in: %s-%s, expect: %s, got: %s", t.namespace, t.name, tt.want, out)
			}
		})
	}
}
