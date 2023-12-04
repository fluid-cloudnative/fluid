package juicefs

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

func TestBuildFormatCmdFilterForEnterpriseEditionAndFilterEncryptEnvOptions(t *testing.T) {
	var mockEncryptEnvOptions = []EncryptEnvOption{
		{Name: AccessKey2, EnvName: utils.ConvertDashToUnderscore(AccessKey2), SecretKeyRefName: "ref1", SecretKeyRefKey: "key1"},
		{Name: SecretKey2, EnvName: utils.ConvertDashToUnderscore(SecretKey2), SecretKeyRefName: "ref2", SecretKeyRefKey: "key2"},
		{Name: "DisallowedEnvOption", EnvName: "env3", SecretKeyRefName: "ref3", SecretKeyRefKey: "key3"},
	}

	testcases := []struct {
		name       string
		givenOpts  []EncryptEnvOption
		expectOpts []EncryptEnvOption
	}{
		{
			"Test with empty options",
			[]EncryptEnvOption{},
			[]EncryptEnvOption{},
		},
		{
			"Test with all allowed options",
			mockEncryptEnvOptions,
			mockEncryptEnvOptions[:2], // expect the first two options
		},
		{
			"Test with duplicate options",
			[]EncryptEnvOption{mockEncryptEnvOptions[0], mockEncryptEnvOptions[0]},
			[]EncryptEnvOption{mockEncryptEnvOptions[0]},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a filter using the buildFormatCmdFilterForEnterpriseEdition function
			filter := buildFormatCmdFilterForEnterpriseEdition()

			// Use the created filter to filter the encrypted environment options
			result := filter.filterEncryptEnvOptions(tc.givenOpts)

			if len(result) == 0 && len(tc.expectOpts) == 0 {
				return
			}

			// Check if the filter result is as expected
			if !reflect.DeepEqual(result, tc.expectOpts) {
				t.Errorf("filterEncryptEnvOptions() = %v, want %v", result, tc.expectOpts)
			}
		})
	}
}

func Test_buildFormatCmdFilterForCommunityEdition(t *testing.T) {
	var mockOptions = map[string]string{
		"JuiceBucket2":     "value",
		"DisallowedOption": "value",
	}

	// Define test cases
	testCases := []struct {
		name           string
		input          map[string]string
		expectedOutput map[string]string
	}{
		{
			name:           "Test with empty options",
			input:          map[string]string{},
			expectedOutput: map[string]string{},
		},
		{
			name:  "Test with all allowed options",
			input: mockOptions,
			expectedOutput: map[string]string{"JuiceBucket2": "value",
				"DisallowedOption": "value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create filter object
			filter := buildFormatCmdFilterForCommunityEdition()

			// Use filter to filter options
			output := filter.filterOption(tc.input)

			if len(output) == 0 && len(tc.expectedOutput) == 0 {
				return
			}

			// Check if the filter result is as expected
			if !reflect.DeepEqual(output, tc.expectedOutput) {
				t.Errorf("filterOption() = %v, want %v", output, tc.expectedOutput)
			}
		})
	}
}

func Test_buildFuseMountCmdFilter(t *testing.T) {
	var mockOptions = map[string]string{
		"JuiceBucket2":     "value",
		"DisallowedOption": "value",
	}

	var mockOptions2 = map[string]string{
		"JuiceBucket2":     "value",
		"DisallowedOption": "value",
		JuiceBucket2:       "bucket2",
	}

	// Define test cases
	testCases := []struct {
		name           string
		input          map[string]string
		expectedOutput map[string]string
	}{
		{
			name:           "Test with empty options",
			input:          map[string]string{},
			expectedOutput: map[string]string{},
		},
		{
			name:  "Test with all allowed options",
			input: mockOptions,
			expectedOutput: map[string]string{"JuiceBucket2": "value",
				"DisallowedOption": "value"},
		}, {
			name:  "Test with disallowed options",
			input: mockOptions2,
			expectedOutput: map[string]string{"JuiceBucket2": "value",
				"DisallowedOption": "value"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Create filter object
			filter := buildFuseMountCmdFilter()

			// Use filter to filter options
			output := filter.filterOption(tc.input)

			if len(output) == 0 && len(tc.expectedOutput) == 0 {
				return
			}

			// Check if the filter result is as expected
			if !reflect.DeepEqual(output, tc.expectedOutput) {
				t.Errorf("filterOption() = %v, want %v", output, tc.expectedOutput)
			}
		})
	}
}
