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

package nodeaffinitywithcache

type Preferred struct {
	name   string `yaml:"name"`
	weight int32  `yaml:"weight"`
}

type TieredLocality struct {
	preferred []Preferred `yaml:"preferred,omitempty"`
}

func (t *TieredLocality) getPreferredAsMap() map[string]int32 {
	localityMap := map[string]int32{}
	for _, preferred := range t.preferred {
		localityMap[preferred.name] = preferred.weight
	}
	return localityMap
}
