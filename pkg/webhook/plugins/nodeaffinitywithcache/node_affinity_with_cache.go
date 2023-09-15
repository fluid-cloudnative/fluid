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
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
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
const NodeLocalityLabel = "fluid.io/node"

var (
	log logr.Logger
)

func init() {
	log = ctrl.Log.WithName(Name)
}

type NodeAffinityWithCache struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *NodeAffinityWithCache {
	return &NodeAffinityWithCache{
		client: c,
		name:   Name,
	}
}

func (p *NodeAffinityWithCache) GetName() string {
	return p.name
}

func (p *NodeAffinityWithCache) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}

	// get required and preferred affinity for runtime
	requireRuntimes, preferredRuntimes := getRequiredAndPreferredSchedulingRuntimes(pod, runtimeInfos)

	// inject required
	err = injectRequiredSchedulingTerms(pod, requireRuntimes)
	if err != nil {
		return shouldStop, err
	}

	// inject preferred
	tieredLocality, err := p.getTieredLocality()
	if err != nil {
		log.Info("get tiered locality config error, skip inject related affinity", "error", err)
		return
	}

	preferredLocality := tieredLocality.getPreferredAsMap()
	for _, runtimeInfo := range preferredRuntimes {
		// get runtime worker node affinity
		status, err := base.GetRuntimeStatus(p.client, runtimeInfo.GetRuntimeType(), runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			return shouldStop, err
		}
		preferredSchedulingTerms, err := p.getTiredLocalityPreferredSchedulingTerm(runtimeInfo, status.WorkerNodeAffinity, preferredLocality)
		if err != nil {
			return shouldStop, err
		}
		if preferredSchedulingTerms != nil {
			utils.InjectPreferredSchedulingTerms(preferredSchedulingTerms, pod)
		}
	}
	return
}

func (p *NodeAffinityWithCache) getTiredLocalityPreferredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface,
	affinity *corev1.NodeAffinity, preferredLocality map[string]int32) (preferredSchedulingTerms []corev1.PreferredSchedulingTerm, err error) {
	// fluid.io/node locality
	nodeLocalityWeight := preferredLocality[NodeLocalityLabel]
	nodePreferredSchedulingTerm, err := getPreferredSchedulingTerm(runtimeInfo, nodeLocalityWeight)
	if err != nil {
		return nil, err
	}
	preferredSchedulingTerms = append(preferredSchedulingTerms, *nodePreferredSchedulingTerm)

	// customized locality
	if affinity == nil {
		return
	}
	if affinity.RequiredDuringSchedulingIgnoredDuringExecution == nil && affinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
		return
	}

	termsFromRequired := getPreferredSchedulingTermsFromRequired(affinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, preferredLocality)
	preferredSchedulingTerms = append(preferredSchedulingTerms, termsFromRequired...)

	termsFromPreferred := getPreferredSchedulingTermsFromPreferred(affinity.PreferredDuringSchedulingIgnoredDuringExecution, preferredLocality)
	preferredSchedulingTerms = append(preferredSchedulingTerms, termsFromPreferred...)

	return
}

func (p *NodeAffinityWithCache) getTieredLocality() (*TieredLocality, error) {
	cm, err := kubeclient.GetConfigmapByName(p.client, "configmap-name", "configmap-namespace")
	if err != nil {
		return nil, err
	}
	tieredLocality := TieredLocality{}
	err = yaml.Unmarshal([]byte(cm.Data["tieredLocality"]), &tieredLocality)
	if err != nil {
		return nil, errors.Wrap(err, "tiered locality content in configmap is not yaml format.")
	}
	return &tieredLocality, nil

}

func getPreferredSchedulingTermsFromPreferred(preferredTerms []corev1.PreferredSchedulingTerm, tiredLabels map[string]int32) (resultSchedulingTerms []corev1.PreferredSchedulingTerm) {
	if preferredTerms == nil {
		return
	}

	for _, term := range preferredTerms {
		for _, matchExpression := range term.Preference.MatchExpressions {
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
				resultSchedulingTerms = append(resultSchedulingTerms, localityTerm)
			}
		}
	}
	return resultSchedulingTerms
}

func getPreferredSchedulingTermsFromRequired(workerRequiredTerms []corev1.NodeSelectorTerm, tiredLabels map[string]int32) (preferredSchedulingTerms []corev1.PreferredSchedulingTerm) {
	if workerRequiredTerms == nil {
		return
	}
	for _, term := range workerRequiredTerms {
		for _, matchExpression := range term.MatchExpressions {
			// label represents tired locality
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

func injectRequiredSchedulingTerms(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) error {
	requiredSchedulingTerms := []corev1.NodeSelectorTerm{}
	for _, runtimeInfo := range runtimeInfos {
		requiredSchedulingTerm, err := getRequiredSchedulingTerms(runtimeInfo)
		if err != nil {
			return err
		}
		if requiredSchedulingTerms != nil {
			requiredSchedulingTerms = append(requiredSchedulingTerms, *requiredSchedulingTerm)
		}
	}
	utils.InjectNodeSelectorTerms(requiredSchedulingTerms, pod)

	return nil
}

func getRequiredSchedulingTerms(runtimeInfo base.RuntimeInfoInterface) (requiredSchedulingTerm *corev1.NodeSelectorTerm, err error) {
	requiredSchedulingTerm = nil
	if runtimeInfo == nil {
		err = fmt.Errorf("RuntimeInfo is nil")
		return
	}

	requiredSchedulingTerm = &corev1.NodeSelectorTerm{
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

func getPreferredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface, weight int32) (preferredSchedulingTerm *corev1.PreferredSchedulingTerm, err error) {
	preferredSchedulingTerm = nil

	if runtimeInfo == nil {
		err = fmt.Errorf("RuntimeInfo is nil")
		return
	}

	isGlobalMode, _ := runtimeInfo.GetFuseDeployMode()
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
