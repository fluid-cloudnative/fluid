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
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a dataset.
   If a pod is labeled for cache, it should require nodes with the mounted dataset on them.
   else, it should prefer nodes with the mounted dataset on them.
*/

const Name = "NodeAffinityWithCache"

// fluid build affinity struct
type builtInAffinity struct {
	// affinity name, i.e. fluid.io/node
	name string
	// the label extractor based on RuntimeInfoInterface
	labelExtractor func(runtimeInterface base.RuntimeInfoInterface) string
}

// use struct not map to avoid the disorder of map traversal, resulting in test case occasional failure.
var fluidBuiltInAffinities = []builtInAffinity{
	{
		// NodeLocalityLabel prefer to schedule pods to nodes with runtime worker pods.
		name: "fluid.io/node",
		labelExtractor: func(runtimeInterface base.RuntimeInfoInterface) string {
			return runtimeInterface.GetCommonLabelName()
		},
	},
	{
		// FuseLocalityLabel prefer to schedule pods to nodes with runtime fuse pods.
		name: "fluid.io/fuse",
		labelExtractor: func(runtimeInterface base.RuntimeInfoInterface) string {
			return runtimeInterface.GetFuseLabelName()
		},
	},
}

var (
	log = ctrl.Log.WithName(Name)
)

type NodeAffinityWithCache struct {
	client         client.Client
	name           string
	tieredLocality *TieredLocality
}

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	var tieredLocality = &TieredLocality{}
	err := yaml.Unmarshal([]byte(args), tieredLocality)
	if err != nil {
		log.Error(err, "the args type is not the TieredLocality format", "args", args)
		return nil, err
	}

	return &NodeAffinityWithCache{
		client:         c,
		name:           Name,
		tieredLocality: tieredLocality,
	}, nil
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

		// fluid builtin locality
		for _, affinity := range fluidBuiltInAffinities {
			weight, existed := preferredLocality[affinity.name]
			if existed {
				preferredSchedulingTerms = append(preferredSchedulingTerms, getPreferredSchedulingTerm(weight, affinity.labelExtractor(runtimeInfo)))
			}
		}

		// customized locality
		statusAffinity := status.CacheAffinity
		if statusAffinity != nil && statusAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			terms := getPreferredSchedulingTermsFromRequired(statusAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, preferredLocality)
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

		// fluid builtin locality
		for _, affinity := range fluidBuiltInAffinities {
			if utils.ContainsString(requireLocalityNames, affinity.name) {
				requiredSchedulingTerms = append(requiredSchedulingTerms, getRequiredSchedulingTerms(affinity.labelExtractor(runtimeInfo)))
			}
		}

		// customized locality
		cacheAffinity := status.CacheAffinity
		// only RequiredDuringSchedulingIgnoredDuringExecution, not considering PreferredDuringSchedulingIgnoredDuringExecution
		if cacheAffinity != nil && cacheAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			terms := getNodeSelectorTermsFromRequired(cacheAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, requireLocalityNames)
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

func getRequiredSchedulingTerms(key string) (requiredSchedulingTerm corev1.NodeSelectorTerm) {
	requiredSchedulingTerm = corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      key,
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

func getPreferredSchedulingTerm(weight int32, key string) (preferredSchedulingTerm corev1.PreferredSchedulingTerm) {
	preferredSchedulingTerm = corev1.PreferredSchedulingTerm{
		Weight: weight,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      key,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	}

	return
}
