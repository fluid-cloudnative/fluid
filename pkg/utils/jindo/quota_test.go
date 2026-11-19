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

package jindo

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSumQuotas(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"gi list", "10Gi,5Gi", "15Gi", false},
		{"single mi", "512Mi", "512Mi", false},
		// Jindo rewrites Mi to m; m must be read as Mi, not milli-bytes.
		{"jindo m suffix", "512m,512m", "1Gi", false},
		{"mixed g and m", "1Gi,512m", "1536Mi", false},
		{"invalid", "10Gi,abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SumQuotas(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			want := resource.MustParse(tt.want)
			if got.Cmp(want) != 0 {
				t.Errorf("SumQuotas(%q) = %s, want %s", tt.in, got.String(), want.String())
			}
		})
	}
}
