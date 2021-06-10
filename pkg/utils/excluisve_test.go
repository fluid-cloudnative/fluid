package utils

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"testing"
)

func TestGetExclusiveKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test for GetExclusiveKey",
			want: common.FluidExclusiveKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExclusiveKey(); got != tt.want {
				t.Errorf("GetExclusiveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExclusiveValue(t *testing.T) {
	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default test-dataset-1",
			args: args{
				name:      "test-dataset-1",
				namespace: "default",
			},
			want: "default_test-dataset-1",
		},
		{
			name: "otherns test-dataset-2",
			args: args{
				name:      "test-dataset-2",
				namespace: "otherns",
			},
			want: "otherns_test-dataset-2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExclusiveValue(tt.args.namespace, tt.args.name); got != tt.want {
				t.Errorf("GetExclusiveValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
