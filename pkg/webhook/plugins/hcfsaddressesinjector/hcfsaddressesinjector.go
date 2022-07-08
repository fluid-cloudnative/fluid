/*
 Copyright 2022 The Fluid Authors.

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

package hcfsaddressesinjector

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const NAME = "HCFSAddressesInjector"

var (
	log = ctrl.Log.WithName(NAME)
)

type HCFSAddressesInjector struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *HCFSAddressesInjector {
	return &HCFSAddressesInjector{
		client: c,
		name:   NAME,
	}
}

func (p *HCFSAddressesInjector) GetName() string {
	return p.name
}

func (p *HCFSAddressesInjector) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod doesn't visit datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}

	// if the pod uses datasets only as PVC, should exit and call other plugins
	datasetsUsedAsHCFS, find := pod.Annotations[common.DatasetUseAsHCFS]
	if !find {
		return
	}

	datasetNames := strings.Split(datasetsUsedAsHCFS, ",")
	var runtimeNames []string

	for _, runtimeInfo := range runtimeInfos {
		if runtimeInfo == nil {
			err = fmt.Errorf("RuntimeInfo is nil")
			shouldStop = true
			return
		}
		find = false
		for _, datasetName := range datasetNames {
			if datasetName == runtimeInfo.GetName() {
				find = true
			}
		}
		if find {
			runtimeNames = append(runtimeNames, runtimeInfo.GetName())
		}
	}
	log.V(1).Info("InjectHCFSAddresses", "runtimeNames", runtimeNames)

	err = utils.InjectHCFSAddresses(p.client, runtimeNames, pod)

	return
}
