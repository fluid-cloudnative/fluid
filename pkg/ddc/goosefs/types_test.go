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

package goosefs

import "testing"

func TestGetTiredStoreLevel0Path(t *testing.T) {
	testCases := map[string]struct {
		name      string
		namespace string
		goosefs   *GooseFS
		wantPath  string
	}{
		"test getTiredStoreLevel0Path case 1": {
			name:      "goosefs-01",
			namespace: "default",
			goosefs: &GooseFS{
				Tieredstore: Tieredstore{
					Levels: []Level{
						{
							Level: 0,
							Path:  "/mnt/demo/data",
						},
					},
				},
			},
			wantPath: "/mnt/demo/data",
		},
		"test getTiredStoreLevel0Path case 2": {
			name:      "goosefs-01",
			namespace: "default",
			goosefs: &GooseFS{
				Tieredstore: Tieredstore{
					Levels: []Level{
						{
							Level: 1,
							Path:  "/mnt/demo/data",
						},
					},
				},
			},
			wantPath: "/dev/shm/default/goosefs-01",
		},
	}

	for k, item := range testCases {
		got := item.goosefs.getTiredStoreLevel0Path(item.name, item.namespace)
		if got != item.wantPath {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantPath, got)
		}
	}
}
