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

package goosefs

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// NodeSelectorRequirement is a selector that contains values, a key, and an operator
// that relates the key and values.
type NodeSelectorRequirement struct {
	// The label key that the selector applies to.
	Key string `yaml:"key,omitempty"`
	// Represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
	Operator string `yaml:"operator,omitempty"`
	// +optional
	Values []string `yaml:"values,omitempty"`
}

// NodeSelectorTerm represents expressions and fields required to select nodes.
// A null or empty node selector term matches no objects. The requirements of
// them are ANDed.
// The TopologySelectorTerm type implements a subset of the NodeSelectorTerm.
type NodeSelectorTerm struct {
	// A list of node selector requirements by node's labels.
	MatchExpressions []NodeSelectorRequirement `yaml:"matchExpressions"`
}

// NodeSelector represents the union of the results of one or more label queries
// over a set of nodes; that is, it represents the OR of the selectors represented
// by the node selector terms.
type NodeSelector struct {
	//Required. A list of node selector terms.
	NodeSelectorTerms []NodeSelectorTerm `yaml:"nodeSelectorTerms"`
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
	RequiredDuringSchedulingIgnoredDuringExecution *NodeSelector `yaml:"requiredDuringSchedulingIgnoredDuringExecution"`
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
