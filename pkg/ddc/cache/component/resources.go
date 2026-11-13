/*
  Copyright 2026 The Fluid Authors.

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

package component

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

// ResourcesEqual reports whether two container resource requirements ask for the same thing.
//
// Requests and limits compare semantically: resource.Quantity caches the string it was parsed
// from and a quantity computed by arithmetic has none, so a structural comparison would report a
// change on every reconcile. A nil list equals an empty one.
//
// Claims compare as a set keyed by name. The API server rejects duplicate names, and the order of
// the list means nothing to the Pod, but a reordering alone would otherwise roll the workload,
// which for an AdvancedStatefulSet recreates its Pods.
func ResourcesEqual(a, b corev1.ResourceRequirements) bool {
	if !equality.Semantic.DeepEqual(a.Requests, b.Requests) || !equality.Semantic.DeepEqual(a.Limits, b.Limits) {
		return false
	}
	if len(a.Claims) != len(b.Claims) {
		return false
	}
	byName := make(map[string]corev1.ResourceClaim, len(a.Claims))
	for _, claim := range a.Claims {
		byName[claim.Name] = claim
	}
	for _, claim := range b.Claims {
		if other, ok := byName[claim.Name]; !ok || other != claim {
			return false
		}
	}
	return true
}
