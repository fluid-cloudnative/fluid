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
		// transform Container Image name etc. A version naming only one of image and imageTag
		// takes the other half from the template rather than being dropped.
		version := desiredComponentVersion(runtimeCompSpec.RuntimeVersion, podTemplate.Spec.Containers[0].Image)
		if len(version.Image) > 0 && len(version.ImageTag) > 0 {
			podTemplate.Spec.Containers[0].Image = version.Image + ":" + version.ImageTag
		}
		if len(runtimeCompSpec.RuntimeVersion.ImagePullPolicy) > 0 {
			podTemplate.Spec.Containers[0].ImagePullPolicy = (corev1.PullPolicy)(runtimeCompSpec.RuntimeVersion.ImagePullPolicy)
		}

		// use runtime component resources if specified, otherwise use default resources
		if runtimeCompSpec.Resources.Limits != nil || runtimeCompSpec.Resources.Requests != nil {
			podTemplate.Spec.Containers[0].Resources = runtimeCompSpec.Resources
		}

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

// desiredComponentVersion completes a partially specified runtime version against the image
// carried by the CacheRuntimeClass template.
//
// VersionSpec declares image and imageTag as independent optional strings, so naming only one
// of them is a legal way to say "the same image on a newer tag". Both the creation path and
// updateImage used to require both halves and otherwise leave the template image alone, which
// silently discards that edit. Completing the missing half here gives both paths the same
// desired image, so a component created with a partial version and one updated to it end up
// identical instead of rolling on the next reconcile.
//
// A version naming neither half is returned untouched, leaving the template image in place. So
// is one whose missing half cannot be recovered - a template with no image, or an image pinned
// by digest - because guessing there would move the workload onto something nobody asked for.
func desiredComponentVersion(runtimeVersion datav1alpha1.VersionSpec, templateImage string) datav1alpha1.VersionSpec {
	if runtimeVersion.Image == "" && runtimeVersion.ImageTag == "" {
		return runtimeVersion
	}
	if runtimeVersion.Image != "" && runtimeVersion.ImageTag != "" {
		return runtimeVersion
	}

	templateRepository, templateTag := splitImageReference(templateImage)
	if runtimeVersion.Image == "" {
		runtimeVersion.Image = templateRepository
	}
	if runtimeVersion.ImageTag == "" {
		runtimeVersion.ImageTag = templateTag
	}

	return runtimeVersion
}

// splitImageReference splits a container image reference into its repository and tag. The tag
// is empty when the reference carries none, and both are empty for a reference pinned by
// digest, where there is no tag to complete and appending one would not be valid.
func splitImageReference(image string) (repository string, tag string) {
	if image == "" || strings.Contains(image, "@") {
		return "", ""
	}

	lastColon := strings.LastIndex(image, ":")
	// A colon before the last slash belongs to a registry port, not to a tag.
	if lastColon == -1 || lastColon < strings.LastIndex(image, "/") {
		return image, ""
	}

	return image[:lastColon], image[lastColon+1:]
}

// componentTemplateImage returns the image the CacheRuntimeClass declares for a component,
// which is the baseline a partially specified runtime version is completed against.
func componentTemplateImage(componentDefinition *datav1alpha1.RuntimeComponentDefinition) string {
	if componentDefinition == nil || len(componentDefinition.Template.Spec.Containers) == 0 {
		return ""
	}

	return componentDefinition.Template.Spec.Containers[0].Image
}
