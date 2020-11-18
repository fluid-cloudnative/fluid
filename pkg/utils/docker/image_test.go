package docker

import (
	"testing"
)

func TestParseDockerImage(t *testing.T) {
	var testCases = []struct {
		input string
		image string
		tag   string
	}{
		{"test:abc", "test", "abc"},
		{"test", "test", "latest"},
	}
	for _, tc := range testCases {
		image, tag := ParseDockerImage(tc.input)
		if tc.image != image {
			t.Errorf("expected image %#v, got %#v",
				tc.image, image)
		}

		if tc.tag != tag {
			t.Errorf("expected tag %#v, got %#v",
				tc.tag, tag)
		}
	}
}
