/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// initComponentValue initializes common fields for a component value
// Returns the initialized component value and an error if validation fails
func (e *CacheEngine) initComponentValue(
	componentType common.ComponentType,
	componentDefinition *datav1alpha1.RuntimeComponentDefinition,
	owner *common.OwnerReference,
	replicas int32,
) (*common.CacheRuntimeComponentValue, error) {
	componentValue := &common.CacheRuntimeComponentValue{
		Name:          common.GetCacheComponentName(e.name, componentType),
		Namespace:     e.namespace,
		Enabled:       true,
		ComponentType: componentType,
		// use deep copy to avoid modifying the original Template
		PodTemplateSpec: *componentDefinition.Template.DeepCopy(),
		Owner:           owner,
		Replicas:        replicas,
	}

	// Set service configuration if headless service is defined
	if componentDefinition.Service.Headless != nil {
		componentValue.Service = &common.CacheRuntimeComponentServiceConfig{
			Name: GetComponentServiceName(e.name, componentType),
		}
	}

	// Validate that at least one container is defined
	if len(componentValue.PodTemplateSpec.Spec.Containers) == 0 {
		return nil, fmt.Errorf("component %s must define at least one container", componentType)
	}

	return componentValue, nil
}

// transformComponentPodTemplate transforms common pod template configurations for master/worker/client components
// This includes image, resources, args, env, nodeSelector, tolerations and pod metadata
func (e *CacheEngine) transformComponentPodTemplate(runtimeCompSpec datav1alpha1.RuntimeComponentCommonSpec,
	dataset *datav1alpha1.Dataset, componentValue *common.CacheRuntimeComponentValue) {
	podTemplate := &componentValue.PodTemplateSpec

	// Pod Meta - Labels and Annotations
	if runtimeCompSpec.PodMetadata.Labels != nil {
		podTemplate.Labels = utils.UnionMapsWithOverride(podTemplate.Labels, runtimeCompSpec.PodMetadata.Labels)
	}
	if runtimeCompSpec.PodMetadata.Annotations != nil {
		podTemplate.Annotations = utils.UnionMapsWithOverride(podTemplate.Annotations, runtimeCompSpec.PodMetadata.Annotations)
	}

	// transform NodeSelector, runtime component takes higher priority
	podTemplate.Spec.NodeSelector = utils.UnionMapsWithOverride(podTemplate.Spec.NodeSelector, runtimeCompSpec.NodeSelector)

	// dataset tolerations apply to all components
	if len(dataset.Spec.Tolerations) > 0 {
		podTemplate.Spec.Tolerations = append(podTemplate.Spec.Tolerations, dataset.Spec.Tolerations...)
	}
	if len(runtimeCompSpec.Tolerations) > 0 {
		podTemplate.Spec.Tolerations = append(podTemplate.Spec.Tolerations, runtimeCompSpec.Tolerations...)
	}

	// envs
	serviceName := ""
	if componentValue.Service != nil {
		serviceName = componentValue.Service.Name
	}
	addEnvs := []corev1.EnvVar{
		{
			Name:  "FLUID_DATASET_NAME",
			Value: e.name,
		},
		{
			Name:  "FLUID_DATASET_NAMESPACE",
			Value: e.namespace,
		},
		{
			Name:  "FLUID_RUNTIME_CONFIG_PATH",
			Value: e.getRuntimeConfigPath(),
		},
		{
			Name:  "FLUID_RUNTIME_MOUNT_PATH",
			Value: e.getFuseMountPoint(),
		},
		{
			Name:  "FLUID_RUNTIME_COMPONENT_TYPE",
			Value: string(componentValue.ComponentType),
		},
		{
			// curvine master sets the CURVINE_MASTER_HOSTNAME with service name
			Name:  "FLUID_RUNTIME_COMPONENT_SVC_NAME",
			Value: serviceName,
		},
	}

	// transform container related config, currently only modify the first container
	if len(podTemplate.Spec.Containers) > 0 {
		// transform Container Image name etc.
		if len(runtimeCompSpec.RuntimeVersion.Image) > 0 && len(runtimeCompSpec.RuntimeVersion.ImageTag) > 0 {
			podTemplate.Spec.Containers[0].Image = runtimeCompSpec.RuntimeVersion.Image + ":" + runtimeCompSpec.RuntimeVersion.ImageTag
		}
		if len(runtimeCompSpec.RuntimeVersion.ImagePullPolicy) > 0 {
			podTemplate.Spec.Containers[0].ImagePullPolicy = (corev1.PullPolicy)(runtimeCompSpec.RuntimeVersion.ImagePullPolicy)
		}

		// Overlay the runtime component resources on the CacheRuntimeClass template
		// baseline key by key: a partially specified resources only moves the keys it
		// names and leaves the rest of the template's requirements in place.
		podTemplate.Spec.Containers[0].Resources = mergeResourceRequirements(
			podTemplate.Spec.Containers[0].Resources, runtimeCompSpec.Resources)

		if runtimeCompSpec.Args != nil {
			podTemplate.Spec.Containers[0].Args = runtimeCompSpec.Args
		}

		if runtimeCompSpec.Env != nil {
			podTemplate.Spec.Containers[0].Env = append(podTemplate.Spec.Containers[0].Env, runtimeCompSpec.Env...)
		}

		// inject envs should come first.
		componentValue.PodTemplateSpec.Spec.Containers[0].Env = append(addEnvs, componentValue.PodTemplateSpec.Spec.Containers[0].Env...)
	}

	if len(componentValue.PodTemplateSpec.Spec.InitContainers) > 0 {
		componentValue.PodTemplateSpec.Spec.InitContainers[0].Env = append(addEnvs, componentValue.PodTemplateSpec.Spec.InitContainers[0].Env...)
	}
}

// mergeResourceRequirements overlays the resources declared on a CacheRuntime component
// on top of the baseline rendered from the CacheRuntimeClass template, key by key.
//
// The two are not alternatives: the template carries the runtime's own requirements and
// the CacheRuntime only expresses the deltas an owner wants for their instance, so
// replacing the whole struct would silently drop every requirement the CacheRuntime does
// not restate. A key the overlay does not name keeps its template value; a key it names
// wins, including a key the template never declared.
//
// The corollary is that a key set by the template cannot be removed by omitting it from
// the CacheRuntime, only overridden. Removing a requirement is a change to the template,
// which is where the runtime's requirements are described in the first place.
//
// Claims are name-keyed rather than merged by resource name: an overlay claim replaces
// the template claim with the same name and any other claim is appended, preserving the
// template's order.
func mergeResourceRequirements(base, overlay corev1.ResourceRequirements) corev1.ResourceRequirements {
	merged := *base.DeepCopy()
	merged.Limits = mergeResourceList(merged.Limits, overlay.Limits)
	merged.Requests = mergeResourceList(merged.Requests, overlay.Requests)
	merged.Claims = mergeResourceClaims(merged.Claims, overlay.Claims)
	return merged
}

// mergeResourceList applies the overlay entries onto base, modifying it in place, and
// returns it. Callers own base: mergeResourceRequirements hands over a deep copy. A nil
// base is kept nil when the overlay declares nothing, so that an untouched component
// still compares equal to the workload it was rendered into.
func mergeResourceList(base, overlay corev1.ResourceList) corev1.ResourceList {
	if len(overlay) == 0 {
		return base
	}
	if base == nil {
		base = corev1.ResourceList{}
	}
	for name, quantity := range overlay {
		base[name] = quantity.DeepCopy()
	}
	return base
}

// mergeResourceClaims applies the overlay claims onto base by name, modifying it in
// place, and returns it.
func mergeResourceClaims(base, overlay []corev1.ResourceClaim) []corev1.ResourceClaim {
	if len(overlay) == 0 {
		return base
	}
	for _, claim := range overlay {
		replaced := false
		for i := range base {
			if base[i].Name == claim.Name {
				base[i] = *claim.DeepCopy()
				replaced = true
				break
			}
		}
		if !replaced {
			base = append(base, *claim.DeepCopy())
		}
	}
	return base
}
