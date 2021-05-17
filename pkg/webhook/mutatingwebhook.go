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

package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=ignore,groups="",resources=pods,verbs=create,versions=v1,name=fluid-pod-admission-webhook

// MutatingHandler mutates a pod and has implemented admission.DecoderInjector
type MutatingHandler struct {
	Client client.Client
	// A decoder will be automatically injected
	decoder *admission.Decoder
}

func NewMutatingHandler(c client.Client) *MutatingHandler {
	return &MutatingHandler{
		Client: c,
	}
}

// Handle is the mutating logic of pod
func (a *MutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	var setupLog = ctrl.Log.WithName("handle")
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		setupLog.Error(err, "unable to decoder pod from req")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check whether should inject
	if pod.Labels["Fluid-Injection"] == "disabled" {
		setupLog.Info("skip mutating the pod because injection is disabled", "Pod", pod.Name, "Namespace", pod.Namespace)
		return admission.Allowed("skip mutating the pod because injection is disabled")
	}
	if pod.Labels["app"] == "alluxio" || pod.Labels["app"] == "jindofs" {
		setupLog.Info("skip mutating the pod because it's fluid Pods", "Pod", pod.Name, "Namespace", pod.Namespace)
		return admission.Allowed("skip mutating the pod because it's fluid Pods")
	}

	// inject affinity info into pod
	a.InjectAffinityToPod(pod)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		setupLog.Error(err, "unable to marshal pod")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// InjectDecoder injects the decoder.
func (a *MutatingHandler) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// InjectAffinityToPod will call all plugins to get total prefer info
func (a *MutatingHandler) InjectAffinityToPod(pod *corev1.Pod) {
	var setupLog = ctrl.Log.WithName("InjectAffinityToPod")
	setupLog.Info("start to inject", "Pod", pod.Name, "Namespace", pod.Namespace)
	pvcNames := kubeclient.GetPVCNamesFromPod(pod)
	var runtimeInfos []base.RuntimeInfoInterface
	for _, pvcName := range pvcNames {
		isDatasetPVC, err := kubeclient.IsDatasetPVC(a.Client, pvcName, pod.Namespace)
		if err != nil {
			setupLog.Error(err, "unable to check pvc, will ignore it", "pvc", pvcName)
			continue
		}
		if isDatasetPVC {
			runtimeInfo, err := base.GetRuntimeInfo(a.Client, pvcName, pod.Namespace)
			if err != nil {
				setupLog.Error(err, "unable to get runtimeInfo, will ignore it", "runtime", pvcName)
				continue
			}
			runtimeInfo.SetDeprecatedNodeLabel(false)
			runtimeInfos = append(runtimeInfos, runtimeInfo)
		}
	}

	pluginsList := plugins.Registry(a.Client)
	for _, plugin := range pluginsList {
		shouldStop := plugin.InjectAffinity(pod, runtimeInfos)
		if shouldStop {
			setupLog.Info("the plugin return true, no need to call other plugins", "plugin", plugin.GetName())
			break
		}
		setupLog.Info("the plugin return false, will call next plugin until last", "plugin", plugin.GetName())
	}

}
