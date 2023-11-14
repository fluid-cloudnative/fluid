/*
Copyright 2023 The Fluid Authors.

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

package utils

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetEnvByKey(t *testing.T) {
	testCases := map[string]struct {
		key       string
		envKey    string
		value     string
		wantValue string
	}{
		"test get env value by key case 1": {
			key:       "MY_POD_NAMESPACE",
			envKey:    "MY_POD_NAMESPACE",
			value:     common.NamespaceFluidSystem,
			wantValue: common.NamespaceFluidSystem,
		},
		"test get env value by key case 2": {
			key:       "MY_POD_NAMESPACE",
			envKey:    "MY_POD_NAMESPACES",
			value:     common.NamespaceFluidSystem,
			wantValue: "",
		},
	}

	for k, item := range testCases {
		// prepare env value
		t.Setenv(item.key, item.value)
		gotValue, _ := GetEnvByKey(item.envKey)
		if gotValue != item.wantValue {
			t.Errorf("%s check failure, want:%v,got:%v", k, item.wantValue, gotValue)
		}
	}
}

func TestIsSubPath(t *testing.T) {
	testCases := map[string]struct {
		path      string
		subPath   string
		isSubPath bool
	}{
		"test is sub path case 1": {
			path:      "/mnt/spark",
			subPath:   "/mnt/spark/data/part0",
			isSubPath: true,
		},
		"test is sub path case 2": {
			path:      "/mnt/spark",
			subPath:   "/mnt/sparks/data/part0",
			isSubPath: false,
		},
		"test is sub path case 3": {
			path:      "/mnt/spark",
			subPath:   "/mnt/spar/kdata/part0",
			isSubPath: false,
		},
		"test is sub path case 4": {
			path:      "/mnt/spark",
			subPath:   "/m/ntsparkdata/part0",
			isSubPath: false,
		},
		"test is sub path case 5": {
			path:      "/mnt/spark",
			subPath:   "/mnts",
			isSubPath: false,
		},
		"test is sub path case 6": {
			path:      "/mnt/spark",
			subPath:   "/mnt/spark",
			isSubPath: true,
		},
		"test is sub path case 7": {
			path:      "/mnt/spark",
			subPath:   "/mnt/spark/data",
			isSubPath: true,
		},
	}

	for k, item := range testCases {
		got := IsSubPath(item.path, item.subPath)
		if got != item.isSubPath {
			t.Errorf("%s check failure,want:%t,got:%t", k, item.isSubPath, got)
		}
	}
}
