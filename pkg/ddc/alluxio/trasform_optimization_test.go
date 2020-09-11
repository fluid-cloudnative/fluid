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

package alluxio

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestOptimizeDefaultProperties(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &Alluxio{}, "alluxio.fuse.jnifuse.enabled", "true"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.alluxioValue)
		if test.alluxioValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.alluxioValue.Properties[test.key], test.key)
		}
	}
}

func TestOptimizeDefaultPropertiesWithSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{
					"alluxio.fuse.jnifuse.enabled": "false",
				},
			},
		}, &Alluxio{}, "alluxio.fuse.jnifuse.enabled", "false"},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		engine.optimizeDefaultProperties(test.runtime, test.alluxioValue)
		if test.alluxioValue.Properties[test.key] != test.expect {
			t.Errorf("expected %s, got %v for key %s", test.expect, test.alluxioValue.Properties[test.key], test.key)
		}
	}
}

func TestSetDefaultPropertiesWithoutSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		value        string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, "alluxio.fuse.jnifuse.enabled", "true", "true"},
	}
	for _, test := range tests {
		setDefaultProperties(test.runtime, test.alluxioValue, test.key, test.value)
		if test.value != test.expect {
			t.Errorf("expected %v, got %v for key %s", test.expect, test.value, test.key)
		}
	}
}

func TestSetDefaultPropertiesWithSet(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.AlluxioRuntime
		alluxioValue *Alluxio
		key          string
		value        string
		expect       string
	}{
		{&datav1alpha1.AlluxioRuntime{
			Spec: datav1alpha1.AlluxioRuntimeSpec{
				Properties: map[string]string{
					"alluxio.fuse.jnifuse.enabled": "false",
				},
			},
		}, &Alluxio{
			Properties: map[string]string{},
		}, "alluxio.fuse.jnifuse.enabled", "true", "false"},
	}
	for _, test := range tests {
		setDefaultProperties(test.runtime, test.alluxioValue, test.key, test.value)
		if test.value == test.expect {
			t.Errorf("expected %v, got %v for key %s", test.expect, test.value, test.key)
		}
	}
}
