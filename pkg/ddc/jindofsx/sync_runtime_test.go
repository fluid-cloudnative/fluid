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

package jindofsx

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestTranformResources(t *testing.T) {
	type args struct {
		runtimeResources corev1.ResourceRequirements
		current          corev1.ResourceRequirements
	}
	tests := []struct {
		name string
		args args
		want corev1.ResourceRequirements
	}{
		// {
		// 	name: "no resource",
		// 	args: args{
		// 		runtimeResources:
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tranformResources(tt.args.runtimeResources, tt.args.current); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tranformResources() = %v, want %v", got, tt.want)
			}
		})
	}
}
