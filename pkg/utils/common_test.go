/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
