/*
Copyright 2022 The Fluid Authors.

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
