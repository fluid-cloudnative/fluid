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

package mutating

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	webhookutils "github.com/fluid-cloudnative/fluid/pkg/webhook/utils"
	"github.com/pkg/errors"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// FluidMutatingHandler mutates a pod and has implemented admission.DecoderInjector
type FluidMutatingHandler struct {
	Client client.Client
	// A decoder will be automatically injected
	decoder *admission.Decoder
}

func (a *FluidMutatingHandler) Setup(client client.Client) {
	a.Client = client
}

// Handle is the mutating logic of pod
func (a *FluidMutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	defer utils.TimeTrack(time.Now(), "CreateUpdatePodForSchedulingHandler.Handle",
		"req.name", req.Name, "req.namespace", req.Namespace)

	if utils.GetBoolValueFromEnv(common.EnvDisableInjection, false) {
		return admission.Allowed("skip mutating the pod because global injection is disabled")
	}

	var setupLog = ctrl.Log.WithName("handle")
	pod := &corev1.Pod{}
	err := a.decoder.Decode(req, pod)
	if err != nil {
		setupLog.Error(err, "unable to decoder pod from req")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Before K8s 1.24, pod.Namespace may not be trustworthy so we deny invalid requests for security concern.
	// See related bugfix at https://github.com/kubernetes/kubernetes/pull/94637
	if len(pod.Namespace) != 0 && pod.Namespace != req.Namespace {
		return admission.Denied("found invalid pod.metadata.namespace, it must either be empty or equal to request's namespace")
	}

	var undoNamespaceOverride bool = false
	if len(pod.Namespace) == 0 {
		if len(req.Namespace) == 0 {
			return admission.Errored(http.StatusInternalServerError, fmt.Errorf("unexepcted error: both pod.metadata.namespace and request's namespace is empty"))
		}
		// Override pod.Namespace with req.Namespace in order to pass namespace info to deeper functions.
		// But we must revert the overriding to avoid a side effect of the mutation.
		setupLog.Info("detecting empty pod.metadata.namespace, overriding it with request.namespace", "request.namespace", req.Namespace)
		pod.Namespace = req.Namespace
		undoNamespaceOverride = true
	}

	// check whether should inject
	if common.CheckExpectValue(pod.Labels, common.EnableFluidInjectionFlag, common.False) {
		setupLog.Info("skip mutating the pod because injection is disabled", "Pod", pod.Name, "Namespace", pod.Namespace)
		return admission.Allowed("skip mutating the pod because injection is disabled")
	}
	if common.CheckExpectValue(pod.Labels, common.InjectSidecarDone, common.True) {
		setupLog.Info("skip mutating the pod because injection is done", "Pod", pod.Name, "Namespace", pod.Namespace)
		return admission.Allowed("skip mutating the pod because injection is done")
	}

	err = a.MutatePod(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if undoNamespaceOverride {
		pod.Namespace = ""
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		setupLog.Error(err, "unable to marshal pod")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	resp := admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
	setupLog.V(1).Info("patch response", "name", pod.GetName(), "namespace", pod.GetNamespace(), "patches", utils.DumpJSON(resp.Patch))
	return resp
}

// InjectDecoder injects the decoder.
func (a *FluidMutatingHandler) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

// MutatePod will call all plugins to get total prefer info
func (a *FluidMutatingHandler) MutatePod(pod *corev1.Pod) (err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "AddScheduleInfoToPod",
			"pod.name", pod.GetName(), "pod.namespace", pod.GetNamespace())
	}
	var setupLog = ctrl.Log.WithName("AddScheduleInfoToPod")
	setupLog.V(1).Info("start to add schedule info", "Pod", pod.Name, "Namespace", pod.Namespace)
	pvcNames := kubeclient.GetPVCNamesFromPod(pod)
	runtimeInfos, err := webhookutils.CollectRuntimeInfosFromPVCs(a.Client, pvcNames, pod.Namespace, setupLog)
	if err != nil {
		setupLog.Error(err, "failed to collect runtime infos from PVCs", "pvcNames", pvcNames)
		return errors.Wrapf(err, "failed to collect runtime infos from PVCs %v", pvcNames)
	}

	// get plugins registry and get the need plugins list from it
	pluginsRegistry := plugins.GetRegistryHandler()
	var pluginsList []api.MutatingHandler

	// handle the pods interact with fluid
	switch {
	case utils.ServerlessEnabled(pod.GetLabels()):
		if len(runtimeInfos) == 0 {
			pluginsList = pluginsRegistry.GetServerlessPodWithoutDatasetHandler()
		} else {
			pluginsList = pluginsRegistry.GetServerlessPodWithDatasetHandler()
		}
	case utils.ServerfulFuseEnabled(pod.GetLabels()):
		if len(runtimeInfos) == 0 {
			pluginsList = pluginsRegistry.GetPodWithoutDatasetHandler()
		} else {
			pluginsList = pluginsRegistry.GetPodWithDatasetHandler()
		}
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
			setupLog.V(1).Info("the plugin return true, no need to hand over other plugins", "plugin", plugin.GetName())
			break
		}
		setupLog.V(1).Info("the plugin return false, will hand over next plugin", "plugin", plugin.GetName())
	}

	return

}
