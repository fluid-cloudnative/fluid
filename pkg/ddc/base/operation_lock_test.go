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

package base

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestGetDataBackupRef(t *testing.T) {
	type args struct {
		object *v1alpha1.DataBackup
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				object: &v1alpha1.DataBackup{
					ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
					Spec:       v1alpha1.DataBackupSpec{},
				},
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataOperationKey(tt.args.object); got != tt.want {
				t.Errorf("GetDataBackupRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
