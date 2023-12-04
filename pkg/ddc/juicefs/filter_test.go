package juicefs

import (
	"reflect"
	"testing"
)

// TestincludeEncryptEnvOptionsWithKeys runs multiple test cases to ensure includeEncryptEnvOptionsWithKeys function behaves as expected.
func TestIncludeEncryptEnvOptionsWithKeys(t *testing.T) {
	tests := []struct {
		name           string
		opts           []EncryptEnvOption
		keys           []string
		expectedResult []EncryptEnvOption
	}{
		{
			name:           "empty options",
			opts:           []EncryptEnvOption{},
			keys:           []string{"name1"},
			expectedResult: []EncryptEnvOption{},
		},
		{
			name:           "single option, no match",
			opts:           []EncryptEnvOption{{Name: "name1", EnvName: "envName1", SecretKeyRefName: "refName1", SecretKeyRefKey: "refKey1"}},
			keys:           []string{"non-existent"},
			expectedResult: []EncryptEnvOption{},
		},
		{
			name:           "single option, matches",
			opts:           []EncryptEnvOption{{Name: "name1", EnvName: "envName1", SecretKeyRefName: "refName1", SecretKeyRefKey: "refKey1"}},
			keys:           []string{"name1"},
			expectedResult: []EncryptEnvOption{{Name: "name1", EnvName: "envName1", SecretKeyRefName: "refName1", SecretKeyRefKey: "refKey1"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := includeEncryptEnvOptionsWithKeys(test.opts, test.keys)

			if len(result) == 0 && len(test.expectedResult) == 0 {
				return
			}

			if !reflect.DeepEqual(result, test.expectedResult) {
				t.Errorf("Expected %v, but got %v", test.expectedResult, result)
			}
		})
	}
}

func TestIncludeOptionsWithKeys(t *testing.T) {
	tests := []struct {
		name           string
		opts           map[string]string
		keys           []string
		expectedResult map[string]string
	}{
		{
			name:           "empty map",
			opts:           map[string]string{},
			keys:           []string{"name1"},
			expectedResult: map[string]string{},
		},
		{
			name:           "single option, no match",
			opts:           map[string]string{"name1": "value1"},
			keys:           []string{"non-existent"},
			expectedResult: map[string]string{},
		},
		{
			name:           "single option, matches",
			opts:           map[string]string{"name1": "value1"},
			keys:           []string{"name1"},
			expectedResult: map[string]string{"name1": "value1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := includeOptionsWithKeys(test.opts, test.keys)

			if len(result) == 0 && len(test.expectedResult) == 0 {
				return
			}

			if !reflect.DeepEqual(result, test.expectedResult) {
				t.Errorf("Expected %v, but got %v", test.expectedResult, result)
			}
		})
	}
}
