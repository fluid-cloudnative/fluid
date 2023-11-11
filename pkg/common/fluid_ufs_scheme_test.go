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

package common

import "testing"

func TestIsFluidNativeScheme(t *testing.T) {
	testCases := map[string]struct {
		endpoint string
		want     bool
	}{
		"test fluid native scheme case 1": {
			endpoint: "pvc://mnt/fluid/data",
			want:     true,
		},
		"test fluid native scheme case 2": {
			endpoint: "local://mnt/fluid/data",
			want:     true,
		},
		"test fluid native scheme case 3": {
			endpoint: "http://mnt/fluid/data",
			want:     false,
		},
		"test fluid native scheme case 4": {
			endpoint: "https://mnt/fluid/data",
			want:     false,
		},
	}

	for k, item := range testCases {
		got := IsFluidNativeScheme(item.endpoint)
		if got != item.want {
			t.Errorf("%s check failure, got:%t,want:%t", k, got, item.want)
		}
	}
}

func TestIsFluidWebScheme(t *testing.T) {
	testCases := map[string]struct {
		endpoint string
		want     bool
	}{
		"test fluid native scheme case 1": {
			endpoint: "pvc://mnt/fluid/data",
			want:     false,
		},
		"test fluid native scheme case 2": {
			endpoint: "local://mnt/fluid/data",
			want:     false,
		},
		"test fluid native scheme case 3": {
			endpoint: "http://mnt/fluid/data",
			want:     true,
		},
		"test fluid native scheme case 4": {
			endpoint: "https://mnt/fluid/data",
			want:     true,
		},
	}

	for k, item := range testCases {
		got := IsFluidWebScheme(item.endpoint)
		if got != item.want {
			t.Errorf("%s check failure, got:%t,want:%t", k, got, item.want)
		}
	}
}

func TestIsFluidRefScheme(t *testing.T) {
	testCases := map[string]struct {
		endpoint string
		want     bool
	}{
		"test fluid native scheme case 1": {
			endpoint: "dataset://mnt/fluid/data",
			want:     true,
		},
		"test fluid native scheme case 2": {
			endpoint: "local://mnt/fluid/data",
			want:     false,
		},
	}

	for k, item := range testCases {
		got := IsFluidRefSchema(item.endpoint)
		if got != item.want {
			t.Errorf("%s check failure, got:%t,want:%t", k, got, item.want)
		}
	}
}
