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
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// +kubebuilder:webhook:path=/mutate-fluid-io-v1alpha1-schedulepod,mutating=true,failurePolicy=fail,sideEffects=None,admissionReviewVersions=v1;v1beta1,groups="",resources=pods,verbs=create;update,versions=v1,name=schedulepod.fluid.io

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]common.AdmissionHandler{
		common.WebhookSchedulePodPath: &FluidMutatingHandler{},
	}
)
