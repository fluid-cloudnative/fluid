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

package v1alpha1

import "testing"

func TestIsHostNetwork(t *testing.T) {
	testCases := map[string]struct {
		n    NetworkMode
		want bool
	}{
		"test host network case 1": {
			n:    HostNetworkMode,
			want: true,
		},
		"test host network case 2": {
			n:    "",
			want: true,
		},
		"test container network case 1": {
			n:    ContainerNetworkMode,
			want: false,
		},
	}

	for k, v := range testCases {
		got := IsHostNetwork(v.n)
		if v.want != got {
			t.Errorf("check %s failure, got:%t,want:%t", k, got, v.want)
		}
	}
}
