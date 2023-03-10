package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestTrimCapabilities(t *testing.T) {
	type args struct {
		inputs       []corev1.Capability
		excludeNames []string
	}
	tests := []struct {
		name        string
		args        args
		wantOutputs []corev1.Capability
	}{
		{
			name: "SYS_ADMIN_only",
			args: args{
				inputs:       []corev1.Capability{"SYS_ADMIN"},
				excludeNames: []string{"SYS_ADMIN"},
			},
			wantOutputs: []corev1.Capability{},
		},
		{
			name: "with_other_capabilities",
			args: args{
				inputs:       []corev1.Capability{"SYS_ADMIN", "CHOWN"},
				excludeNames: []string{"SYS_ADMIN"},
			},
			wantOutputs: []corev1.Capability{"CHOWN"},
		},
		{
			name: "exclude_multiple_capabilities",
			args: args{
				inputs:       []corev1.Capability{"SYS_ADMIN", "CHOWN", "SETPCAP"},
				excludeNames: []string{"SYS_ADMIN", "SETPCAP"},
			},
			wantOutputs: []corev1.Capability{"CHOWN"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOutputs := TrimCapabilities(tt.args.inputs, tt.args.excludeNames); !reflect.DeepEqual(gotOutputs, tt.wantOutputs) {
				t.Errorf("TrimCapabilities() = %v, want %v", gotOutputs, tt.wantOutputs)
			}
		})
	}
}
