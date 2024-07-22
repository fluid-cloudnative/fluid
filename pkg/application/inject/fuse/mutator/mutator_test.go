/*
Copyright 2023 The Fluid Authors.

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

package mutator

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindExtraArgsFromMetadata(t *testing.T) {
	type args struct {
		metaObj  metav1.ObjectMeta
		platform string
	}
	tests := []struct {
		name          string
		args          args
		wantExtraArgs map[string]string
	}{
		{
			name: "empty_annotations",
			args: args{
				metaObj: metav1.ObjectMeta{
					Annotations: nil,
				},
				platform: "myplatform",
			},
			wantExtraArgs: nil,
		},
		{
			name: "without_extra_args",
			args: args{
				metaObj: metav1.ObjectMeta{
					Annotations: map[string]string{"foo": "bar"},
				},
				platform: "myplatform",
			},
			wantExtraArgs: nil,
		},
		{
			name: "with_extra_args",
			args: args{
				metaObj: metav1.ObjectMeta{
					Annotations: map[string]string{"foo": "bar", "myplatform.fluid.io/key1": "value1", "myplatform.fluid.io/key2": "value2"},
				},
				platform: "myplatform",
			},
			wantExtraArgs: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotExtraArgs := FindExtraArgsFromMetadata(tt.args.metaObj, tt.args.platform); !reflect.DeepEqual(gotExtraArgs, tt.wantExtraArgs) {
				t.Errorf("FindExtraArgsFromMetadata() = %v, want %v", gotExtraArgs, tt.wantExtraArgs)
			}
		})
	}
}
