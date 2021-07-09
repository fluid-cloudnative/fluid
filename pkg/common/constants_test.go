package common

import (
	"testing"
)

func TestGetDefaultTieredStoreOrder(t *testing.T) {
	testCases := map[string]struct {
		mediumType MediumType
		want       int
	}{
		"test case 1": {
			mediumType: Memory,
			want:       0,
		},
		"test case 2": {
			mediumType: SSD,
			want:       1,
		},
		"test case 3": {
			mediumType: HDD,
			want:       2,
		},
		"test case 4": {
			mediumType: "unknown",
			want:       0,
		},
	}
	for k, item := range testCases {
		result := GetDefaultTieredStoreOrder(item.mediumType)
		if item.want != result {
			t.Errorf("%s cannot paas, want %v, get %v", k, item.want, result)
		}
	}

}
