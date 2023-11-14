/*
Copyright 2023 The Fluid Authors.

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

package mountpropagationinjector

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

const Name = "MountPropagationInjector"

var (
	log = ctrl.Log.WithName(Name)
)

type MountPropagationInjector struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client, args interface{}) api.MutatingHandler {
	return &MountPropagationInjector{
		client: c,
		name:   Name,
	}
}

func (p *MountPropagationInjector) GetName() string {
	return p.name
}

func (p *MountPropagationInjector) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}
	datasetNames := make([]string, len(runtimeInfos))
	for name, runtimeInfo := range runtimeInfos {
		if runtimeInfo == nil {
			err = fmt.Errorf("RuntimeInfo is nil")
			shouldStop = true
			return
		}
		// do not use the runtime name, as the pvc may be the dataset mounting another dataset
		datasetNames = append(datasetNames, name)
	}
	log.V(1).Info("InjectMountPropagation", "datasetNames", datasetNames)
	utils.InjectMountPropagation(datasetNames, pod)

	return
}
