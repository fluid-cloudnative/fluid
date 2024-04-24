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

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// NodeSelectorRequirement is a selector that contains values, a key, and an operator
// that relates the key and values.
type NodeSelectorRequirement struct {
	// The label key that the selector applies to.
	Key string `json:"key,omitempty"`
	// Represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
	Operator string `json:"operator,omitempty"`
	// +optional
	Values []string `json:"values,omitempty"`
}

// NodeSelectorTerm represents expressions and fields required to select nodes.
// A null or empty node selector term matches no objects. The requirements of
// them are ANDed.
// The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
type NodeSelectorTerm struct {
	// A list of node selector requirements by node's labels.
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions"`
}

// NodeSelector represents the union of the results of one or more label queries
// over a set of nodes; that is, it represents the OR of the selectors represented
// by the node selector terms.
type NodeSelector struct {
	//Required. A list of node selector terms.
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms"`
}

type NodeAffinity struct {
	// NOT YET IMPLEMENTED. TODO: Uncomment field once it is implemented.
	// If the affinity requirements specified by this field are not met at
	// scheduling time, the pod will not be scheduled onto the node.
	// If the affinity requirements specified by this field cease to be met
	// at some point during pod execution (e.g. due to an update), the system
	// will try to eventually evict the pod from its node.
	// +optional
	// RequiredDuringSchedulingRequiredDuringExecution *NodeSelector

	// If the affinity requirements specified by this field are not met at
	// scheduling time, the pod will not be scheduled onto the node.
	// If the affinity requirements specified by this field cease to be met
	// at some point during pod execution (e.g. due to an update), the system
	// may or may not try to eventually evict the pod from its node.
	// +optional
	RequiredDuringSchedulingIgnoredDuringExecution *NodeSelector `json:"requiredDuringSchedulingIgnoredDuringExecution"`
}

func translateCacheToNodeAffinity(dataAffinity *datav1alpha1.CacheableNodeAffinity) (nodeAffinity *NodeAffinity) {
	nodeAffinity = nil
	if dataAffinity == nil || dataAffinity.Required == nil {
		return
	}

	nodeAffinity = &NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &NodeSelector{
			NodeSelectorTerms: []NodeSelectorTerm{},
		},
	}

	for _, srcTerm := range dataAffinity.Required.NodeSelectorTerms {
		dstTerm := NodeSelectorTerm{
			MatchExpressions: []NodeSelectorRequirement{},
		}

		for _, srcMatch := range srcTerm.MatchExpressions {

			dstMatch := NodeSelectorRequirement{
				Key:      srcMatch.Key,
				Operator: string(srcMatch.Operator),
				Values:   srcMatch.Values,
			}

			dstTerm.MatchExpressions = append(dstTerm.MatchExpressions, dstMatch)
		}
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			dstTerm)

	}

	return

}
