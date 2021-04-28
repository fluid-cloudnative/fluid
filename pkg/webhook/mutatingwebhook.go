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
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	"net/http"
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
	setupLog.Info("start to handle")
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		setupLog.Error(err, "unable to decoder pod from req")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check whether should inject
	if pod.Labels["Fluid-Injection"] == "disabled" {
		return admission.Allowed("injection is disabled, no need to prefer")
	}
	if pod.Labels["app"] == "alluxio" || pod.Labels["app"] == "jindofs" {
		return admission.Allowed("fluid Pods, will not prefer")
	}

	// inject prefer and not prefer info to pod
	a.InjectPreferToPod(pod)

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

// InjectPreferToPod will call all plugins to get total prefer info
func (a *MutatingHandler) InjectPreferToPod(pod *corev1.Pod) {
	var (
		nodePreferTerms     []corev1.PreferredSchedulingTerm
		podPreferTerms      []corev1.WeightedPodAffinityTerm
		podAntiPreferTerms  []corev1.WeightedPodAffinityTerm
		nodeRequireTerms    *corev1.NodeSelector
		podRequireTerms     []corev1.PodAffinityTerm
		podAntiRequireTerms []corev1.PodAffinityTerm
	)
	// get the origin prefer and require terms from pod
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			nodePreferTerms = pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			nodeRequireTerms = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
		if pod.Spec.Affinity.PodAffinity != nil {
			podPreferTerms = pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			podRequireTerms = pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			podPreferTerms = pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			podAntiRequireTerms = pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}

	// todo: Parallel execution of all plugins
	for _, plugin := range plugins.Registry(a.Client) {
		pluginPreferredSchedulingTerms := plugin.NodePrefer(*pod)
		nodePreferTerms = append(nodePreferTerms, pluginPreferredSchedulingTerms...)

		pluginPodAffinityTerms := plugin.PodPrefer(*pod)
		podPreferTerms = append(podPreferTerms, pluginPodAffinityTerms...)

		pluginPodAntiAffinityTerms := plugin.PodNotPrefer(*pod)
		podAntiPreferTerms = append(podAntiPreferTerms, pluginPodAntiAffinityTerms...)

	}

	// generate the final affinity
	nodeAffinity := corev1.NodeAffinity{}
	if len(nodePreferTerms) != 0 {
		nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = nodePreferTerms
	}
	if nodeRequireTerms != nil {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = nodeRequireTerms
	}

	podAffinity := corev1.PodAffinity{}
	if len(podPreferTerms) != 0 {
		podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = podPreferTerms
	}
	if len(podRequireTerms) != 0 {
		podAffinity.RequiredDuringSchedulingIgnoredDuringExecution = podRequireTerms
	}

	podAntiAffinity := corev1.PodAntiAffinity{}
	if len(podAntiPreferTerms) != 0 {
		podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = podAntiPreferTerms
	}
	if len(podAntiRequireTerms) != 0 {
		podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = podAntiRequireTerms
	}

	pod.Spec.Affinity = &corev1.Affinity{
		NodeAffinity:    &nodeAffinity,
		PodAffinity:     &podAffinity,
		PodAntiAffinity: &podAntiAffinity,
	}
}
