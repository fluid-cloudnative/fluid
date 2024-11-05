/*
Copyright 2021 The Fluid Authors.

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

package alluxio

import "testing"

func TestGetTiredStoreLevel0Path(t *testing.T) {
	testCases := map[string]struct {
		name      string
		namespace string
		alluxio   *Alluxio
		wantPath  string
	}{
		"test getTiredStoreLevel0Path case 1": {
			name:      "alluxio-01",
			namespace: "default",
			alluxio: &Alluxio{
				TieredStore: TieredStore{
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
			name:      "alluxio-01",
			namespace: "default",
			alluxio: &Alluxio{
				TieredStore: TieredStore{
					Levels: []Level{
						{
							Level: 1,
							Path:  "/mnt/demo/data",
						},
					},
				},
			},
			wantPath: "/dev/shm/default/alluxio-01",
		},
	}

	for k, item := range testCases {
		got := item.alluxio.getTiredStoreLevel0Path(item.name, item.namespace)
		if got != item.wantPath {
			t.Errorf("%s check failure, want:%s,got:%s", k, item.wantPath, got)
		}
	}
}
