/*
Copyright 2018 The Kubernetes Authors.

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
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
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
	// the serviceName of webhook-manager
	serviceName string
}

func NewMutatingHandler(c client.Client, serviceName string) *MutatingHandler {
	return &MutatingHandler{
		Client:      c,
		serviceName: serviceName,
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
	if pod.Labels["app"] == "alluxio" || pod.Labels["Fluid-Injection"] == "disabled" {
		return admission.Allowed("no need to prefer")
	}
	pvcNames := kubeclient.GetPVCNamesFromPod(pod)

	var datasetList = &v1alpha1.DatasetList{}
	var alluxioRuntime = &v1alpha1.AlluxioRuntime{}
	err = a.Client.List(ctx, datasetList)
	if err != nil {
		setupLog.Error(err, "unable to list dataset")
		return admission.Errored(http.StatusBadRequest, err)
	}

	var preferredSchedulingTerms []corev1.PreferredSchedulingTerm
	for _, dataset := range datasetList.Items {
		ifMount := false
		for _, pvcName := range pvcNames {
			if dataset.Name == pvcName && dataset.Namespace == pod.Namespace {
				// pod has a mounted dataset
				ifMount = true
				err = a.Client.Get(ctx, types.NamespacedName{
					Namespace: dataset.Namespace,
					Name:      dataset.Name,
				}, alluxioRuntime)
				if err != nil {
					setupLog.Error(err, "unable to get alluxioRuntime")
					continue
				}
				if alluxioRuntime.Spec.Fuse.Global == true {
					// Pod should prefer to choose nodes with this dataset
					preferredSchedulingTerms = append(preferredSchedulingTerms, corev1.PreferredSchedulingTerm{
						Weight: 1,
						Preference: corev1.NodeSelectorTerm{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "data.fluid.io/storage-" + dataset.Namespace + "-" + dataset.Name,
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"true"},
								},
							},
						},
					})
				}
			}
		}
		if !ifMount {
			// pod has no relationship with this dataset
			preferredSchedulingTerms = append(preferredSchedulingTerms, corev1.PreferredSchedulingTerm{
				Weight: 1,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "data.fluid.io/storage-" + dataset.Namespace + "-" + dataset.Name,
							Operator: corev1.NodeSelectorOpNotIn,
							Values:   []string{"true"},
						},
					},
				},
			})
		}
	}

	pod.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTerms,
		},
	}

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
