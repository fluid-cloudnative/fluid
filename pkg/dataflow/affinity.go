/*
Copyright 2024 The Fluid Authors.

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

package dataflow

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InjectAffinityByRunAfterOp inject the affinity based on preceding operation
func InjectAffinityByRunAfterOp(c client.Client, runAfter *datav1alpha1.OperationRef, opNamespace string, currentAffinity *v1.Affinity) (*v1.Affinity, error) {
	// no previous operation or use default affinity strategy, no need to generate node affinity
	if runAfter == nil || runAfter.AffinityStrategy.Policy == datav1alpha1.DefaultAffinityStrategy {
		return currentAffinity, nil
	}
	precedingOpNamespace := opNamespace
	if len(runAfter.Namespace) != 0 {
		precedingOpNamespace = runAfter.Namespace
	}

	precedingOpStatus, err := utils.GetPrecedingOperationStatus(c, runAfter, precedingOpNamespace)
	if err != nil {
		return currentAffinity, err
	}

	// require policy
	if runAfter.AffinityStrategy.Policy == datav1alpha1.RequireAffinityStrategy {
		return injectRequiredAffinity(runAfter, precedingOpStatus.NodeAffinity, currentAffinity)
	}

	// prefer policy
	if runAfter.AffinityStrategy.Policy == datav1alpha1.PreferAffinityStrategy {
		return injectPreferredAffinity(runAfter, precedingOpStatus.NodeAffinity, currentAffinity)
	}

	return currentAffinity, fmt.Errorf("unknown policy for affinity strategy: %s", runAfter.AffinityStrategy.Policy)
}

func injectPreferredAffinity(runAfter *datav1alpha1.OperationRef, nodeLabels map[string]string, currentAffinity *v1.Affinity) (*v1.Affinity, error) {
	var preferTerms []v1.PreferredSchedulingTerm
	prefer := runAfter.AffinityStrategy.Prefer
	if prefer == nil {
		prefer = []datav1alpha1.Prefer{
			{
				Name:   common.K8sNodeNameLabelKey,
				Weight: 100,
			},
		}
	}
	for _, item := range prefer {
		if value, ok := nodeLabels[item.Name]; ok {
			preferTerms = append(preferTerms, v1.PreferredSchedulingTerm{
				Weight: item.Weight,
				Preference: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      item.Name,
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{value},
						},
					},
				},
			})
		}
	}
	return utils.InjectPreferredSchedulingTermsToAffinity(preferTerms, currentAffinity), nil
}

func injectRequiredAffinity(runAfter *datav1alpha1.OperationRef, prevOpNodeAffinity *v1.NodeAffinity, currentAffinity *v1.Affinity) (*v1.Affinity, error) {
	if prevOpNodeAffinity == nil || prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		return currentAffinity, nil
	}

	var terms []v1.NodeSelectorTerm
	keys := runAfter.AffinityStrategy.Require
	if keys == nil {
		keys = []string{common.K8sNodeNameLabelKey}
	}
	for _, key := range keys {

		for match := range prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {

		}

		if value, ok := nodeLabels[key]; ok {
			terms = append(terms, v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      key,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{value},
					},
				},
			})
		}
	}
	return utils.InjectNodeSelectorTermsToAffinity(terms, currentAffinity), nil
}
