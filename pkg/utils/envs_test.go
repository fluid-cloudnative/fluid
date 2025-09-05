package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGeEnvsDifference(t *testing.T) {
	env1 := corev1.EnvVar{Name: "env1"}
	env2 := corev1.EnvVar{Name: "env2"}
	env3 := corev1.EnvVar{Name: "env3"}
	env4 := corev1.EnvVar{Name: "env4"}

	tests := []struct {
		name     string
		base     []corev1.EnvVar
		filter   []corev1.EnvVar
		expected []corev1.EnvVar
	}{
		{
			name:     "nil_envs",
			base:     []corev1.EnvVar{},
			filter:   []corev1.EnvVar{},
			expected: []corev1.EnvVar{},
		},
		{
			name:     "test base envs are nil",
			base:     []corev1.EnvVar{},
			filter:   []corev1.EnvVar{env1, env2},
			expected: []corev1.EnvVar{},
		},
		{
			name:     "test exclude envs are nil",
			base:     []corev1.EnvVar{env1, env2},
			filter:   []corev1.EnvVar{},
			expected: []corev1.EnvVar{env1, env2},
		},
		{
			name:     "test base envs are same with exclude envs",
			base:     []corev1.EnvVar{env1, env2},
			filter:   []corev1.EnvVar{env1, env2},
			expected: []corev1.EnvVar{},
		},
		{
			name:     "test base envs includes all exclude envs",
			base:     []corev1.EnvVar{env1, env2, env3, env4},
			filter:   []corev1.EnvVar{env2, env4},
			expected: []corev1.EnvVar{env1, env3},
		},
		{
			name:     "test base envs do not include with exclude envs",
			base:     []corev1.EnvVar{env1, env2},
			filter:   []corev1.EnvVar{env3, env4},
			expected: []corev1.EnvVar{env1, env2},
		},
		{
			name:     "test base envs includes partial exclude envs",
			base:     []corev1.EnvVar{env1, env2, env3},
			filter:   []corev1.EnvVar{env2, env4},
			expected: []corev1.EnvVar{env1, env3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEnvsDifference(tt.base, tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("expected slice length %d, actual slice length %d", len(tt.expected), len(result))
				return
			}
			expectedMap := make(map[string]bool)
			for _, v := range tt.expected {
				expectedMap[v.Name] = true
			}

			resultMap := make(map[string]bool)
			for _, v := range result {
				resultMap[v.Name] = true
			}
			for name, expectedEnv := range expectedMap {
				resultEnv, exist := resultMap[name]
				if !exist {
					t.Errorf("expected env %s, but not exist in return", name)
				}

				if !reflect.DeepEqual(resultEnv, expectedEnv) {
					t.Errorf("expected env %v, but got %v", expectedEnv, resultEnv)
				}
			}
		})
	}
}
