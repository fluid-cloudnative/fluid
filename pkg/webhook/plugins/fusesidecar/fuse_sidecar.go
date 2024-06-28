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

package fusesidecar

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	webhookutils "github.com/fluid-cloudnative/fluid/pkg/webhook/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject"
	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods without a mounted dataset.
   They should prefer nods without cache workers on them.

*/

const Name string = "FuseSidecar"

type FuseSidecar struct {
	client client.Client
	name   string
	log    logr.Logger
}

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	return &FuseSidecar{
		client: c,
		name:   Name,
		log:    ctrl.Log.WithName("FuseSidecar"),
	}, nil
}

func (p *FuseSidecar) GetName() string {
	return p.name
}

func (p *FuseSidecar) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "FuseSidecar.Mutate",
			"pod.name", pod.GetName(), "pvc.namespace", pod.GetNamespace())
	}

	for len(runtimeInfos) > 0 {
		var injector inject.Injector = fuse.NewInjector(p.client)
		out, err := injector.InjectPod(pod, runtimeInfos)
		if err != nil {
			return shouldStop, errors.Wrapf(err, "failed to inject pod \"%v/%v\" with runtimeInfos", pod.Namespace, pod.Name)
		}
		out.DeepCopyInto(pod)
		pvcNames := kubeclient.GetPVCNamesFromPod(pod)
		runtimeInfos, err = webhookutils.CollectRuntimeInfosFromPVCs(p.client, pvcNames, pod.Namespace, p.log)
		if err != nil {
			return shouldStop, errors.Wrapf(err, "failed to collect runtime infos from PVCs %v", pvcNames)
		}
	}
	p.labelInjectionDone(pod)
	return
}

func (p *FuseSidecar) labelInjectionDone(pod *corev1.Pod) {
	if pod.ObjectMeta.Labels == nil {
		pod.ObjectMeta.Labels = map[string]string{}
	}
	pod.ObjectMeta.Labels[common.InjectSidecarDone] = common.True
	pod.ObjectMeta.Labels[common.LabelAnnotationManagedBy] = common.Fluid
}
