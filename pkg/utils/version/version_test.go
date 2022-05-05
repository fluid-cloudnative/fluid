package version

import "testing"

func TestParse(t *testing.T) {
	validVersions := []string{
		"2.7.2-SNAPSHOT-3714f2b",
		"release-2.7.2-SNAPSHOT-3714f2b",
		"2.8.0",
	}

	for _, s := range validVersions {
		t.Run(s, func(t *testing.T) {
			ver, err := RuntimeVersion(s)
			t.Log("Valid: ", s, ver, err)
			if err != nil {
				t.Errorf("RuntimeVersion unexpected error for version %q: %v", s, err)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	v1 := "release-2.7.2-SNAPSHOT-3714f2b"
	v2 := "2.8.0"
	expect := -1
	got, err := Compare(v1, v2)
	if err != nil {
		t.Errorf("Compare unexpected error for version %q: %v", v1, err)
	}

	if got != -1 {
		t.Errorf("Expect %q, but got %q", expect, got)
	}

}
