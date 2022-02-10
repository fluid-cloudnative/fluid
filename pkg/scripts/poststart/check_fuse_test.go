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

package poststart

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"testing"
)

func TestScriptGeneratorForFuse_getConfigmapName(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		mountPath string
		mountType string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test-alluxio",
			fields: fields{
				name:      "test",
				namespace: "default",
				mountPath: "/dev",
				mountType: common.ALLUXIO_MOUNT_TYPE,
			},
			want: "test-" + common.ALLUXIO_MOUNT_TYPE + "-" + configMapName,
		},
		{
			name: "test-jindo",
			fields: fields{
				name:      "test",
				namespace: "default",
				mountPath: "/dev",
				mountType: common.JindoMountType,
			},
			want: "test-" + common.JindoMountType + "-" + configMapName,
		},
		{
			name: "test-juicefs",
			fields: fields{
				name:      "test",
				namespace: "default",
				mountPath: "/dev",
				mountType: common.JuiceFSMountType,
			},
			want: "test-juicefs-" + configMapName,
		},
		{
			name: "test-goosefs",
			fields: fields{
				name:      "test",
				namespace: "default",
				mountPath: "/dev",
				mountType: common.GooseFSMountType,
			},
			want: "test-" + common.GooseFSMountType + "-" + configMapName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &ScriptGeneratorForFuse{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				mountPath: tt.fields.mountPath,
				mountType: tt.fields.mountType,
			}
			if got := f.getConfigmapName(); got != tt.want {
				t.Errorf("getConfigmapName() = %v, want %v", got, tt.want)
			}
		})
	}
}
