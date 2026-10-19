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
	"sort"
	"strings"

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
// This includes image, resources, args, env, nodeSelector, tolerations, image pull secrets and pod metadata.
//
// Fields carried by both runtimeSpec and runtimeCompSpec are layered the way spec.options already is:
// the CacheRuntimeClass template first, then the runtime-level spec, then the component-level spec.
func (e *CacheEngine) transformComponentPodTemplate(runtimeSpec datav1alpha1.CacheRuntimeSpec,
	runtimeCompSpec datav1alpha1.RuntimeComponentCommonSpec,
	dataset *datav1alpha1.Dataset, componentValue *common.CacheRuntimeComponentValue) {
	podTemplate := &componentValue.PodTemplateSpec

	// Pod Meta - Labels and Annotations, template < runtime level < component level.
	// UnionMapsWithOverride copies both operands, so a nil map at any layer is a no-op.
	podTemplate.Labels = utils.UnionMapsWithOverride(
		utils.UnionMapsWithOverride(podTemplate.Labels, runtimeSpec.PodMetadata.Labels),
		runtimeCompSpec.PodMetadata.Labels)
	podTemplate.Annotations = utils.UnionMapsWithOverride(
		utils.UnionMapsWithOverride(podTemplate.Annotations, runtimeSpec.PodMetadata.Annotations),
		runtimeCompSpec.PodMetadata.Annotations)

	// ImagePullSecrets has no component-level counterpart, so the runtime-level list is merged
	// onto whatever the template already declares. The CRD marks the field with a merge patch
	// strategy keyed on name, so a secret named at both layers must not appear twice.
	podTemplate.Spec.ImagePullSecrets = appendMissingImagePullSecrets(
		podTemplate.Spec.ImagePullSecrets, runtimeSpec.ImagePullSecrets)

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

// appendMissingImagePullSecrets appends the secrets that are not already referenced, comparing
// by name so that a secret declared both on the CacheRuntimeClass template and on the
// CacheRuntime is carried once.
func appendMissingImagePullSecrets(existing []corev1.LocalObjectReference,
	toAdd []corev1.LocalObjectReference) []corev1.LocalObjectReference {
	if len(toAdd) == 0 {
		return existing
	}

	present := make(map[string]bool, len(existing))
	for _, secret := range existing {
		present[secret.Name] = true
	}

	for _, secret := range toAdd {
		if present[secret.Name] {
			continue
		}
		present[secret.Name] = true
		existing = append(existing, secret)
	}

	return existing
}

// mergeResourceRequirements overlays the CacheRuntime component resources onto the
// CacheRuntimeClass template baseline key by key: a key the overlay names wins, any other
// keeps its template value, so a template key can be overridden but not removed. Claims
// are matched by name, replacing on a match and appending otherwise.
func mergeResourceRequirements(base, overlay corev1.ResourceRequirements) corev1.ResourceRequirements {
	merged := *base.DeepCopy()
	mergeResourceList(&merged.Limits, overlay.Limits)
	mergeResourceList(&merged.Requests, overlay.Requests)
	mergeResourceClaims(&merged.Claims, overlay.Claims)
	return merged
}

// mergeResourceList applies the overlay entries onto *base in place: an entry with the
// same name is replaced and any other entry is added. An empty overlay changes nothing,
// so `limits: {}` on the CacheRuntime reads the same as leaving limits out.
func mergeResourceList(base *corev1.ResourceList, overlay corev1.ResourceList) {
	if len(overlay) == 0 {
		return
	}
	if *base == nil {
		*base = corev1.ResourceList{}
	}
	for name, quantity := range overlay {
		(*base)[name] = quantity.DeepCopy()
	}
}

// mergeResourceClaims applies the overlay claims onto *base by name in place: a claim
// with the same name is replaced and any other claim is appended.
func mergeResourceClaims(base *[]corev1.ResourceClaim, overlay []corev1.ResourceClaim) {
	for _, claim := range overlay {
		replaced := false
		for i := range *base {
			if (*base)[i].Name == claim.Name {
				(*base)[i] = *claim.DeepCopy()
				replaced = true
				break
			}
		}
		if !replaced {
			*base = append(*base, *claim.DeepCopy())
		}
	}
}

// validateResourceRequirements reports every request whose relation to its limit the API
// server would reject: a request above its limit, and, for resources that cannot be
// overcommitted (huge pages and extended resources), a request that differs from its limit
// or has no limit at all. Merging per key can produce such a pair from a template and a
// CacheRuntime that are each valid on their own, and the API server would only reject it
// once a Pod is created. Other Pod resource rules, such as integer extended resources or
// huge page sizes, are left to the API server: a merge cannot break them when both inputs
// satisfy them.
func validateResourceRequirements(resources corev1.ResourceRequirements) error {
	var conflicts []string
	for name, request := range resources.Requests {
		limit, ok := resources.Limits[name]
		switch {
		case !ok:
			if !isOvercommitAllowed(name) {
				conflicts = append(conflicts, fmt.Sprintf("%s request %s has no limit", name, request.String()))
			}
		case !isOvercommitAllowed(name) && request.Cmp(limit) != 0:
			conflicts = append(conflicts, fmt.Sprintf("%s request %s must equal limit %s", name, request.String(), limit.String()))
		case request.Cmp(limit) > 0:
			conflicts = append(conflicts, fmt.Sprintf("%s request %s exceeds limit %s", name, request.String(), limit.String()))
		}
	}
	if len(conflicts) == 0 {
		return nil
	}
	sort.Strings(conflicts)
	return fmt.Errorf("%s", strings.Join(conflicts, "; "))
}

// isOvercommitAllowed mirrors k8s.io/kubernetes/pkg/apis/core/v1/helper.IsOvercommitAllowed,
// which is not vendored: only native resources other than huge pages may set a request
// below their limit.
func isOvercommitAllowed(name corev1.ResourceName) bool {
	native := !strings.Contains(string(name), "/") ||
		strings.Contains(string(name), corev1.ResourceDefaultNamespacePrefix)
	return native && !strings.HasPrefix(string(name), corev1.ResourceHugePagesPrefix)
}

// splitDeclaredClaims splits container claims into those the sync may apply to a workload and
// the names of those it may not. The sync patches container resources only, never the
// pod-level resourceClaims entries, so a claim the template added to both after creation would
// reach the container alone, and with DynamicResourceAllocation enabled the API server would
// then reject every replacement Pod. A claim is applied when the workload's pod spec declares
// it, or when its first container already carries it: dropping such a claim would roll the
// workload without changing what its Pods get.
func splitDeclaredClaims(claims []corev1.ResourceClaim, podSpec *corev1.PodSpec) (declared []corev1.ResourceClaim, undeclared []string) {
	known := map[string]bool{}
	if podSpec != nil {
		for _, podClaim := range podSpec.ResourceClaims {
			known[podClaim.Name] = true
		}
		if len(podSpec.Containers) > 0 {
			for _, claim := range podSpec.Containers[0].Resources.Claims {
				known[claim.Name] = true
			}
		}
	}
	for _, claim := range claims {
		if known[claim.Name] {
			declared = append(declared, claim)
		} else {
			undeclared = append(undeclared, claim.Name)
		}
	}
	sort.Strings(undeclared)
	return declared, undeclared
}

// validateComponentResources checks the resources the creation path rendered into the
// component's first container, the only one transformComponentPodTemplate fills in.
func validateComponentResources(componentValue *common.CacheRuntimeComponentValue) error {
	containers := componentValue.PodTemplateSpec.Spec.Containers
	if len(containers) == 0 {
		return nil
	}
	if err := validateResourceRequirements(containers[0].Resources); err != nil {
		return fmt.Errorf("invalid %s resources: %w", componentValue.ComponentType, err)
	}
	return nil
}

// checkComponentResources rejects the creation of a component whose rendered resources
// conflict, emitting the same RuntimeResourcesConflict Warning event the sync path emits so
// the reason is visible on the CacheRuntime and not only in the generic setup failure.
func (e *CacheEngine) checkComponentResources(runtime *datav1alpha1.CacheRuntime, componentValue *common.CacheRuntimeComponentValue) error {
	err := validateComponentResources(componentValue)
	if err != nil && e.Recorder != nil {
		e.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeResourcesConflict,
			"%s cannot be created: the CacheRuntime and its CacheRuntimeClass template merge into %v",
			componentValue.ComponentType, err)
	}
	return err
}
