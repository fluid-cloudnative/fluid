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

import (
	"errors"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a dataset.
   If a pod is labeled for cache, it should require nodes with the mounted dataset on them.
   else, it should prefer nodes with the mounted dataset on them.
*/

const Name = "NodeAffinityWithCache"
const NodeLocalityLabel = "fluid.io/node"

var (
	log logr.Logger
)

type NodeAffinityWithCache struct {
	client         client.Client
	name           string
	tieredLocality *TieredLocality
}

func NewPlugin(c client.Client, args string) api.MutatingHandler {
	var tieredLocality TieredLocality
	err := yaml.Unmarshal([]byte(args), &tieredLocality)
	if err != nil {
		log.Error(err, "the args type is the TieredLocality format", "args", args)
	}

	return &NodeAffinityWithCache{
		client:         c,
		name:           Name,
		tieredLocality: &tieredLocality,
	}
}

func (p *NodeAffinityWithCache) GetTieredLocality() *TieredLocality {
	return p.tieredLocality
}

func (p *NodeAffinityWithCache) GetName() string {
	return p.name
}

func (p *NodeAffinityWithCache) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}

	tieredLocality := p.tieredLocality
	if tieredLocality == nil {
		log.Error(errors.New("the plugin tiered locality config may has wrong format"), "skip mutating for plugin", "name", Name)
		return
	}

	// app pod contains locality affinity, not inject.
	if tieredLocality.hasRepeatedLocality(pod) {
		log.Info("Warning: pdd already has locality affinity, skip injecting cache affinity.")
		return
	}

	// get required and preferred affinity for runtime
	requireRuntimes, preferredRuntimes := getRequiredAndPreferredSchedulingRuntimes(pod, runtimeInfos)

	// inject required
	nodeSelectorTerms, err := p.getTieredLocalityNodeSelectorTerms(requireRuntimes, tieredLocality.Required)
	if err != nil {
		return shouldStop, err
	}
	utils.InjectNodeSelectorTerms(nodeSelectorTerms, pod)

	// inject preferred
	preferredSchedulingTerms, err := p.getTieredLocalityPreferredSchedulingTerms(preferredRuntimes, tieredLocality.getPreferredAsMap())
	if err != nil {
		return shouldStop, err
	}
	utils.InjectPreferredSchedulingTerms(preferredSchedulingTerms, pod)

	return
}

func (p *NodeAffinityWithCache) getTieredLocalityPreferredSchedulingTerms(preferredRuntimes map[string]base.RuntimeInfoInterface,
	preferredLocality map[string]int32) (preferredSchedulingTerms []corev1.PreferredSchedulingTerm, err error) {
	for name, runtimeInfo := range preferredRuntimes {
		if runtimeInfo == nil {
			log.Info("Warning: pvc has no runtime, skip inject affinity", "Name", name)
			continue
		}

		// get runtime worker node affinity
		status, err := base.GetRuntimeStatus(p.client, runtimeInfo.GetRuntimeType(), runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			return preferredSchedulingTerms, err
		}

		// fluid.io/node locality
		nodeLocalityWeight, existed := preferredLocality[NodeLocalityLabel]
		if existed {
			nodePreferredSchedulingTerm := getPreferredSchedulingTerm(runtimeInfo, nodeLocalityWeight)
			if nodePreferredSchedulingTerm != nil {
				preferredSchedulingTerms = append(preferredSchedulingTerms, *nodePreferredSchedulingTerm)
			}
		}
		// customized locality
		affinity := status.CacheAffinity
		if affinity != nil && affinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			terms := getPreferredSchedulingTermsFromRequired(affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, preferredLocality)
			preferredSchedulingTerms = append(preferredSchedulingTerms, terms...)
		}
	}
	return
}

func (p *NodeAffinityWithCache) getTieredLocalityNodeSelectorTerms(runtimeInfos map[string]base.RuntimeInfoInterface,
	requireLocalityNames []string) (requiredSchedulingTerms []corev1.NodeSelectorTerm, err error) {

	for name, runtimeInfo := range runtimeInfos {
		if runtimeInfo == nil {
			log.Info("Warning: pvc has no runtime, skip inject affinity", "Name", name)
			continue
		}

		// get runtime worker node affinity
		status, err := base.GetRuntimeStatus(p.client, runtimeInfo.GetRuntimeType(), runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			return requiredSchedulingTerms, err
		}

		// fluid.io/node locality
		if utils.ContainsString(requireLocalityNames, NodeLocalityLabel) {
			requiredSchedulingTerms = append(requiredSchedulingTerms, getRequiredSchedulingTerms(runtimeInfo))
		}

		// customized locality
		affinity := status.CacheAffinity
		// only RequiredDuringSchedulingIgnoredDuringExecution, not considering PreferredDuringSchedulingIgnoredDuringExecution
		if affinity != nil && affinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			terms := getNodeSelectorTermsFromRequired(affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, requireLocalityNames)
			requiredSchedulingTerms = append(requiredSchedulingTerms, terms...)
		}
	}

	return
}

func getRequiredAndPreferredSchedulingRuntimes(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (
	requiredRuntimes map[string]base.RuntimeInfoInterface, preferredRuntimes map[string]base.RuntimeInfoInterface) {

	// get the pod specified bound dataset
	boundDataSetNames := map[string]bool{}
	for key, value := range pod.Labels {
		// if it matches, return array with length 2, and the second element is the match
		matchString := common.LabelAnnotationPodSchedRegex.FindStringSubmatch(key)
		if len(matchString) == 2 && value == "required" {
			boundDataSetNames[matchString[1]] = true
		}
	}

	// get the preferred and required runtime
	preferredRuntimes = map[string]base.RuntimeInfoInterface{}
	requiredRuntimes = map[string]base.RuntimeInfoInterface{}

	// copy instead of use origin runtimeInfos
	for name, runtimeInfo := range runtimeInfos {
		preferredRuntimes[name] = runtimeInfo
	}

	for name := range boundDataSetNames {
		// labeled dataset has no runtime(maybe wrong name), logging info instead of return error
		runtimeInfo, ok := runtimeInfos[name]
		if !ok {
			log.V(1).Info("labeled dataset has no runtime")
			continue
		}

		requiredRuntimes[name] = runtimeInfo
		delete(preferredRuntimes, name)
	}

	return
}

func getNodeSelectorTermsFromRequired(nodeSelectorTerms []corev1.NodeSelectorTerm, requireLocalityNames []string) (resultTerms []corev1.NodeSelectorTerm) {
	for _, term := range nodeSelectorTerms {
		for _, matchExpression := range term.MatchExpressions {
			// Key represents tiered locality name
			if utils.ContainsString(requireLocalityNames, matchExpression.Key) {
				localityTerm := corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      matchExpression.Key,
							Operator: matchExpression.Operator,
							Values:   matchExpression.Values,
						},
					},
				}
				resultTerms = append(resultTerms, localityTerm)
			}
		}
	}
	return
}

func getRequiredSchedulingTerms(runtimeInfo base.RuntimeInfoInterface) (requiredSchedulingTerm corev1.NodeSelectorTerm) {
	requiredSchedulingTerm = corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      runtimeInfo.GetCommonLabelName(),
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"true"},
			},
		},
	}
	return
}

func getPreferredSchedulingTermsFromRequired(nodeSelectorTerms []corev1.NodeSelectorTerm, tiredLabels map[string]int32) (preferredSchedulingTerms []corev1.PreferredSchedulingTerm) {
	for _, term := range nodeSelectorTerms {
		for _, matchExpression := range term.MatchExpressions {
			// Key represents tiered locality name
			if weight, ok := tiredLabels[matchExpression.Key]; ok {
				localityTerm := corev1.PreferredSchedulingTerm{
					Weight: weight,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      matchExpression.Key,
								Operator: matchExpression.Operator,
								Values:   matchExpression.Values,
							},
						},
					},
				}
				preferredSchedulingTerms = append(preferredSchedulingTerms, localityTerm)
			}
		}
	}
	return
}

func getPreferredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface, weight int32) (preferredSchedulingTerm *corev1.PreferredSchedulingTerm) {
	isGlobalMode, _ := runtimeInfo.GetFuseDeployMode()
	// since fluid 0.7, always true
	if isGlobalMode {
		preferredSchedulingTerm = &corev1.PreferredSchedulingTerm{
			Weight: weight,
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      runtimeInfo.GetCommonLabelName(),
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		}
	}

	return
}
