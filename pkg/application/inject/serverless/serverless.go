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

package serverless

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var (
	injectScheme      = runtime.NewScheme()
	fuseContainerName = "fluid-fuse"
)

// InjectObject injects sidecar into
func InjectObject(in runtime.Object, sidecarTemplate common.ServerlessInjectionTemplate) (out interface{}, err error) {
	out = in.DeepCopyObject()

	var containers []corev1.Container
	var volumes []corev1.Volume
	var metadata *metav1.ObjectMeta
	var typeMeta metav1.TypeMeta

	// Handle Lists
	if list, ok := out.(*corev1.List); ok {
		result := list

		for i, item := range list.Items {
			obj, err := fromRawToObject(item.Raw)
			if runtime.IsNotRegisteredError(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			r, err := InjectObject(obj, sidecarTemplate)
			if err != nil {
				return nil, err
			}

			re := runtime.RawExtension{}
			re.Object = r.(runtime.Object)
			result.Items[i] = re
		}
		return result, nil
	}

	switch v := out.(type) {
	case *corev1.Pod:
		pod := v
		containers = pod.Spec.Containers
		volumes = pod.Spec.Volumes
		typeMeta = pod.TypeMeta
		metadata = &pod.ObjectMeta
	default:
		log.Info("No supported K8s Type", "v", v)
		return out, fmt.Errorf("no support for K8s Type %v", v)
	}

	name := types.NamespacedName{
		Namespace: metadata.Namespace,
		Name:      metadata.Name,
	}
	kind := typeMeta.Kind

	// Skip injection for injected container
	if len(containers) > 0 {
		for _, c := range containers {
			if c.Name == fuseContainerName {
				warningStr := fmt.Sprintf("===> Skipping injection because %v has injected %q sidecar already\n",
					name, fuseContainerName)
				if len(kind) != 0 {
					warningStr = fmt.Sprintf("===> Skipping injection because Kind %s: %v has injected %q sidecar already\n",
						kind, name, fuseContainerName)
				}
				log.Info(warningStr)
				return out, nil
			}
		}
	}

	// 1.Modify the volumes
	for i, v := range volumes {
		for _, toUpdate := range sidecarTemplate.VolumesToUpdate {
			if v.Name == toUpdate.Name {
				volumes[i] = toUpdate
				// break
			}
		}
	}

	// 2.Add the volumes
	if len(sidecarTemplate.VolumeMountsToAdd) > 0 {
		volumes = append(volumes, sidecarTemplate.VolumesToAdd...)
	}

	// 2.Modify and add the volumeMounts of containers
	for _, c := range containers {
		shouldInsert := false
		for i, v := range c.VolumeMounts {
			for _, toUpdate := range sidecarTemplate.VolumeMountsToUpdate {
				if toUpdate.Name == v.Name {
					c.VolumeMounts[i] = toUpdate
					shouldInsert = true
					// break
				}
			}
		}

		if shouldInsert {
			if len(sidecarTemplate.VolumeMountsToAdd) > 0 {
				c.VolumeMounts = append(c.VolumeMounts, sidecarTemplate.VolumeMountsToAdd...)
			}
		}

	}

	// 3.Add sidecar as the first container
	containers = append([]corev1.Container{sidecarTemplate.FuseContainer}, containers...)

	return out, err
}

// fromRawToObject is used to convert from raw to the runtime object
func fromRawToObject(raw []byte) (runtime.Object, error) {
	var typeMeta metav1.TypeMeta
	if err := yaml.Unmarshal(raw, &typeMeta); err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(typeMeta.APIVersion, typeMeta.Kind)
	obj, err := injectScheme.New(gvk)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(raw, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
