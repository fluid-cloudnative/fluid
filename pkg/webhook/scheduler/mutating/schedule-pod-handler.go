/*

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

package mutating

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdatePodForSchedulingHandler mutates a pod and has implemented admission.DecoderInjector
type CreateUpdatePodForSchedulingHandler struct {
	Client client.Client
	// A decoder will be automatically injected
	decoder *admission.Decoder
}

func (a *CreateUpdatePodForSchedulingHandler) Setup(client client.Client) {
	a.Client = client
}

// Handle is the mutating logic of pod
func (a *CreateUpdatePodForSchedulingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	var setupLog = ctrl.Log.WithName("handle")
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		setupLog.Error(err, "unable to decoder pod from req")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check whether should inject
	if common.CheckExpectValue(pod.Labels, common.EnableFluidInjectionFlag, common.False) {
		setupLog.Info("skip mutating the pod because injection is disabled", "Pod", pod.Name, "Namespace", pod.Namespace)
		return admission.Allowed("skip mutating the pod because injection is disabled")
	}
	if pod.Labels["app"] == "alluxio" || pod.Labels["app"] == "jindofs" {
		setupLog.Info("skip mutating the pod because it's fluid Pods", "Pod", pod.Name, "Namespace", pod.Namespace)
		return admission.Allowed("skip mutating the pod because it's fluid Pods")
	}

	// inject affinity info into pod
	err = a.AddScheduleInfoToPod(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		setupLog.Error(err, "unable to marshal pod")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// InjectDecoder injects the decoder.
func (a *CreateUpdatePodForSchedulingHandler) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// AddScheduleInfoToPod will call all plugins to get total prefer info
func (a *CreateUpdatePodForSchedulingHandler) AddScheduleInfoToPod(pod *corev1.Pod) (err error) {
	var setupLog = ctrl.Log.WithName("AddScheduleInfoToPod")
	setupLog.Info("start to add schedule info", "Pod", pod.Name, "Namespace", pod.Namespace)
	pvcNames := kubeclient.GetPVCNamesFromPod(pod)
	var runtimeInfos []base.RuntimeInfoInterface
	for _, pvcName := range pvcNames {
		isDatasetPVC, err := kubeclient.IsDatasetPVC(a.Client, pvcName, pod.Namespace)
		if err != nil {
			setupLog.Error(err, "unable to check pvc, will ignore it", "pvc", pvcName)
			return err
		}
		if isDatasetPVC {
			runtimeInfo, err := base.GetRuntimeInfo(a.Client, pvcName, pod.Namespace)
			if err != nil {
				setupLog.Error(err, "unable to get runtimeInfo, will ignore it", "runtime", pvcName)
				return err
			}
			runtimeInfo.SetDeprecatedNodeLabel(false)
			runtimeInfos = append(runtimeInfos, runtimeInfo)
		}
	}

	// get plugins Registry and get the need plugins list from it
	pluginsRegistry := plugins.Registry(a.Client)
	var pluginsList []plugins.MutatingHandler
	if len(runtimeInfos) == 0 {
		pluginsList = pluginsRegistry.GetPodWithoutDatasetHandler()
	} else {
		pluginsList = pluginsRegistry.GetPodWithDatasetHandler()
	}

	// call every plugin in the plugins list in the defined order
	// if a plugin return shouldStop, stop to call other plugins
	for _, plugin := range pluginsList {
		shouldStop, err := plugin.Mutate(pod, runtimeInfos)
		if err != nil {
			setupLog.Error(err, "Failed to mutate pod")
			return err
		}

		if shouldStop {
			setupLog.Info("the plugin return true, no need to call other plugins", "plugin", plugin.GetName())
			break
		}
		setupLog.Info("the plugin return false, will call next plugin until last", "plugin", plugin.GetName())
	}

	return

}
