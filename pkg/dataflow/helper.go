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
	"errors"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func GenerateNodeAffinity(job *batchv1.Job) (*corev1.NodeAffinity, error) {
	if job == nil {
		return nil, nil
	}
	// not inject, i.e. feature gate not enabled or job is a parallel job.
	if v := job.Annotations[common.AnnotationDataFlowAffinityInject]; v != "true" {
		return nil, nil
	}

	annotations := job.Annotations

	nodeAffinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: nil,
				},
			},
		},
	}

	// affinity labels with specific prefix.
	hasInjectedLabels := false
	for key, value := range annotations {
		if strings.HasPrefix(key, common.AnnotationDataFlowCustomizedAffinityPrefix) {
			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
				append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions,
					corev1.NodeSelectorRequirement{
						Key:      strings.TrimPrefix(key, common.AnnotationDataFlowCustomizedAffinityPrefix),
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{value},
					})
			hasInjectedLabels = true
		}
	}
	if !hasInjectedLabels {
		return nil, errors.New("the affinity label is not set, wait for next reconcile")
	}

	return nodeAffinity, nil
}
