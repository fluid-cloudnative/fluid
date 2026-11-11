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
// Requests and limits are compared semantically rather than structurally: resource.Quantity
// caches the string it was parsed from, and a quantity produced by arithmetic carries an empty
// cache, so a structural comparison would see a change on every reconcile even when the value
// is identical. A nil and an empty list are equal.
//
// Claims are compared as a set keyed by name. The API server rejects duplicate claim names and
// the order a claim is listed in changes nothing for the Pod, while a reordering alone would
// otherwise roll the workload, which for an AdvancedStatefulSet means recreating its Pods.
func ResourcesEqual(a, b corev1.ResourceRequirements) bool {
	if !equality.Semantic.DeepEqual(a.Limits, b.Limits) || !equality.Semantic.DeepEqual(a.Requests, b.Requests) {
		return false
	}
	if len(a.Claims) != len(b.Claims) {
		return false
	}
	claims := make(map[string]corev1.ResourceClaim, len(a.Claims))
	for _, claim := range a.Claims {
		claims[claim.Name] = claim
	}
	for _, claim := range b.Claims {
		if other, found := claims[claim.Name]; !found || other != claim {
			return false
		}
	}
	return true
}
