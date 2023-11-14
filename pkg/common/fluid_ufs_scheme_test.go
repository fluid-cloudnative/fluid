/*

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
