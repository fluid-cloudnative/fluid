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
