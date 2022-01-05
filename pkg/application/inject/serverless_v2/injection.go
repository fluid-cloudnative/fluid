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

package serverless_v2

import (
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	reflectutil "github.com/fluid-cloudnative/fluid/pkg/utils/reflect"
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
func InjectObject(in runtime.Object, sidecarTemplate common.ServerlessInjectionTemplate) (out runtime.Object, err error) {
	out = in.DeepCopyObject()

	var containersPtr *[]corev1.Container
	var volumesPtr *[]corev1.Volume
	var objectMeta metav1.ObjectMeta
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
			re.Object = r
			result.Items[i] = re
		}
		return result, nil
	}

	switch v := out.(type) {
	case *corev1.Pod:
		pod := v
		containersPtr = &pod.Spec.Containers
		volumesPtr = &pod.Spec.Volumes
		typeMeta = pod.TypeMeta
		objectMeta = pod.ObjectMeta
	default:
		log.Info("No supported K8s Type", "v", v)
		outValue := reflect.ValueOf(out).Elem()

		containersReferenceName, containersValue, err := reflectutil.ContainersValueFromObject(out, "", []string{"init"})
		if err != nil {
			return out, fmt.Errorf("get container references failed for K8s Type %v with error %v", v, err)
		}

		log.Info("ContainersValueFromObject", "containersReferenceName", containersReferenceName)

		volumesReferenceName, volumesValue, err := reflectutil.VolumesValueFromObject(out, "", []string{})
		if err != nil {
			return out, fmt.Errorf("get volume Reference volume for K8s Type %v with error %v", v, err)
		}

		log.Info("VolumesValueFromObject", "volumesReferenceName", volumesReferenceName)

		containersPtr = containersValue.Addr().Interface().(*[]corev1.Container)
		volumesPtr = volumesValue.Addr().Interface().(*[]corev1.Volume)
		typeMeta = outValue.FieldByName("TypeMeta").Interface().(metav1.TypeMeta)
		objectMeta = outValue.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
	}

	isServerless := ServerlessEnabled(objectMeta.Annotations)
	if !isServerless {
		return out, nil
	}

	name := types.NamespacedName{
		Namespace: objectMeta.Namespace,
		Name:      objectMeta.Name,
	}
	kind := typeMeta.Kind

	// Skip injection for injected container
	if len(*containersPtr) > 0 {
		for _, c := range *containersPtr {
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
	for i, v := range *volumesPtr {
		for _, toUpdate := range sidecarTemplate.VolumesToUpdate {
			if v.Name == toUpdate.Name {
				log.V(1).Info("Update volume", "original", v, "updated", toUpdate)
				(*volumesPtr)[i] = toUpdate
				// break
			}
		}
	}

	// 2.Add the volumes
	if len(sidecarTemplate.VolumesToAdd) > 0 {
		log.V(1).Info("Before append volume", "original", (*volumesPtr))
		(*volumesPtr) = append((*volumesPtr), sidecarTemplate.VolumesToAdd...)
		log.V(1).Info("After append volume", "original", (*volumesPtr))
	}

	// 3.Add sidecar as the first container
	*containersPtr = append([]corev1.Container{sidecarTemplate.FuseContainer}, *containersPtr...)

	log.V(1).Info("Updated resource", "containers", *containersPtr, "volumes", *volumesPtr)

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

func ServerlessEnabled(annotions map[string]string) (match bool) {
	for key, value := range annotions {
		if key == common.Serverless && value == common.True {
			match = true
			break
		}
	}
	return
}
