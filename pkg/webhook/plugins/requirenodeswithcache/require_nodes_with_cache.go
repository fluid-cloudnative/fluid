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

package requirenodeswithcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a  dataset.
   They should require nods with fuse.
*/

const NAME = "RequireNodesWithCache"

var (
	log logr.Logger
)

func init() {
	log = ctrl.Log.WithName(NAME)
}

type RequireNodesWithCache struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *RequireNodesWithCache {
	return &RequireNodesWithCache{
		client: c,
		name:   NAME,
	}
}

func (p *RequireNodesWithCache) GetName() string {
	return p.name
}

// Mutate mutates the pod based on runtimeInfo, this action shouldn't stop other handler
func (p *RequireNodesWithCache) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}

	requiredSchedulingTerms := []corev1.NodeSelectorTerm{}

	// get the pod specified bound dataset
	boundDataSetNames := map[string]bool{}
	for key, value := range pod.Labels {
		// if it matches, return array with length 2, and the second element is the match
		matchString := common.LabelAnnotationPodSchedRegex.FindStringSubmatch(key)
		if len(matchString) == 2 && value == "required" {
			boundDataSetNames[matchString[1]] = true
		}
	}

	// no dataset labeled for required scheduling
	if len(boundDataSetNames) == 0 {
		return
	}

	// append the node selector terms to pod
	for name := range boundDataSetNames {
		runtime := runtimeInfos[name]
		// dataset has no runtime, logging info
		if runtime == nil {
			log.V(1).Info("do not inject required cache affinity as the labeled dataset is not exist", "dataset", name)
			continue
		}

		term := corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      runtime.GetCommonLabelName(),
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		}

		requiredSchedulingTerms = append(requiredSchedulingTerms, term)
	}

	if len(requiredSchedulingTerms) > 0 {
		utils.InjectNodeSelectorTerms(requiredSchedulingTerms, pod)
	}

	return
}
