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
	"github.com/pkg/errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InjectAffinityAnnotation inject the affinity annotation for the pod.
func InjectAffinityAnnotation(opAnnotation map[string]string, podAnnotation map[string]string) map[string]string {
	value, exist := opAnnotation[common.AnnotationDataFlowAffinityLabelsName]
	if !exist {
		return podAnnotation
	}
	// return a copy not the origin map
	copiedMap := make(map[string]string)
	for k, v := range podAnnotation {
		copiedMap[k] = v
	}

	copiedMap[common.AnnotationDataFlowAffinityLabelsName] = value
	return copiedMap
}

// InjectAffinityByRunAfterOp inject the affinity based on preceding operation
func InjectAffinityByRunAfterOp(c client.Client, runAfter *datav1alpha1.OperationRef, opNamespace string, currentAffinity *v1.Affinity) (*v1.Affinity, error) {
	// no previous operation or use default affinity strategy, no need to generate node affinity
	if runAfter == nil || runAfter.AffinityStrategy.Policy == datav1alpha1.DefaultAffinityStrategy {
		return currentAffinity, nil
	}

	// if not specified, use the runAfter ObjectRef as dependent affinity operation
	dependOnOp := runAfter.AffinityStrategy.DependOn
	if dependOnOp == nil {
		dependOnOp = &runAfter.ObjectRef
	}

	precedingOpNamespace := opNamespace
	if len(dependOnOp.Namespace) != 0 {
		precedingOpNamespace = dependOnOp.Namespace
	}

	precedingOpStatus, err := utils.GetPrecedingOperationStatus(c, dependOnOp, precedingOpNamespace)
	if err != nil {
		return currentAffinity, err
	}

	// ensure the dependent operation was completed, the outer caller will record the event.
	if precedingOpStatus.Phase != common.PhaseComplete {
		return nil, errors.New(fmt.Sprintf("dependOn operation %s status is %s, not completed.", dependOnOp.Name, precedingOpStatus.Phase))
	}

	if precedingOpStatus.NodeAffinity == nil {
		return currentAffinity, nil
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

func injectPreferredAffinity(runAfter *datav1alpha1.OperationRef, prevOpNodeAffinity *v1.NodeAffinity, currentAffinity *v1.Affinity) (*v1.Affinity, error) {
	var preferTerms []v1.PreferredSchedulingTerm
	prefers := runAfter.AffinityStrategy.Prefers
	if len(prefers) == 0 {
		prefers = []datav1alpha1.Prefer{
			{
				Name:   common.K8sNodeNameLabelKey,
				Weight: 100,
			},
		}
	}

	if len(prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		return currentAffinity, fmt.Errorf("no node selector terms in the preceding operation")
	}

	// currently, only has one element.
	podNodeSelectorTerm := prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0]

	for _, item := range prefers {
		key := item.Name
		for _, expression := range podNodeSelectorTerm.MatchExpressions {
			if expression.Key == key {
				preferTerms = append(preferTerms, v1.PreferredSchedulingTerm{
					Weight: item.Weight,
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      expression.Key,
								Operator: expression.Operator,
								Values:   expression.Values,
							},
						},
					},
				})
			}
		}
	}
	return utils.InjectPreferredSchedulingTermsToAffinity(preferTerms, currentAffinity), nil
}

func injectRequiredAffinity(runAfter *datav1alpha1.OperationRef, prevOpNodeAffinity *v1.NodeAffinity, currentAffinity *v1.Affinity) (*v1.Affinity, error) {
	if prevOpNodeAffinity == nil || prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil ||
		prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		return currentAffinity, nil
	}

	var matchExpressions []v1.NodeSelectorRequirement
	requires := runAfter.AffinityStrategy.Requires
	if len(requires) == 0 {
		requires = []datav1alpha1.Require{
			{
				Name: common.K8sNodeNameLabelKey,
			},
		}
	}

	if len(prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		return currentAffinity, fmt.Errorf("no node selector terms in the preceding operation")
	}

	// currently, only has one element.
	podNodeSelectorTerm := prevOpNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0]

	for _, item := range requires {
		for _, expression := range podNodeSelectorTerm.MatchExpressions {
			if expression.Key == item.Name {
				matchExpressions = append(matchExpressions,
					v1.NodeSelectorRequirement{
						Key:      expression.Key,
						Operator: expression.Operator,
						Values:   expression.Values,
					},
				)
			}
		}
	}
	return utils.InjectNodeSelectorRequirements(matchExpressions, currentAffinity), nil
}
