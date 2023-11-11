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

package efc

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestGetTiredStoreLevel0(t *testing.T) {
	testCases := map[string]struct {
		name            string
		namespace       string
		efc             *EFC
		wantPath        string
		wantType        string
		wantQuotaString string
		wantMediumType  string
	}{
		"test getTiredStoreLevel0Path case 1": {
			name:      "efc-01",
			namespace: "default",
			efc: &EFC{
				Worker: Worker{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								Level:      0,
								Path:       "/mnt/demo/data",
								Type:       string(common.VolumeTypeEmptyDir),
								MediumType: string(common.Memory),
								Quota:      "1GB",
							},
						},
					},
				},
			},
			wantPath:        "/mnt/demo/data",
			wantType:        string(common.VolumeTypeEmptyDir),
			wantQuotaString: "1GB",
			wantMediumType:  string(common.Memory),
		},
	}

	for k, item := range testCases {
		got := item.efc.getTiredStoreLevel0Path()
		if got != item.wantPath {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantPath, got)
		}

		gott := item.efc.getTiredStoreLevel0Type()
		if gott != item.wantType {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantType, gott)
		}

		gottt := item.efc.getTiredStoreLevel0Quota()
		if gottt != item.wantQuotaString {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantQuotaString, gottt)
		}

		gotttt := item.efc.getTiredStoreLevel0MediumType()
		if gotttt != item.wantMediumType {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantMediumType, gotttt)
		}
	}
}
