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

package fuse

import (
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/defaultapp"
	podapp "github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	"github.com/fluid-cloudnative/fluid/pkg/utils/applications/unstructured"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructuredtype "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log logr.Logger
)

func init() {
	log = ctrl.Log.WithName("fuse")
}

type Injector struct {
	client client.Client
}

func NewInjector(client client.Client) *Injector {
	return &Injector{
		client: client,
	}
}

// InjectPod injects pod with runtimeInfo which key is pvcName, value is runtimeInfo
func (s *Injector) InjectPod(in *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (out *corev1.Pod, err error) {
	match := false
	outObj, err := s.Inject(in, runtimeInfos)
	if err != nil {
		return out, err
	}
	out, match = outObj.(*corev1.Pod)
	if !match {
		return out, fmt.Errorf("failed to match the object %v to %v", outObj, out)
	}

	return
}

func (s *Injector) inject(in runtime.Object, pvcName string, runtimeInfo base.RuntimeInfoInterface) (out runtime.Object, err error) {
	out = in.DeepCopyObject()

	var (
		application common.FluidApplication
		objectMeta  metav1.ObjectMeta
		typeMeta    metav1.TypeMeta
	)

	// Handle Lists
	if list, ok := out.(*corev1.List); ok {
		result := list

		for i, item := range list.Items {
			obj, err := utils.FromRawToObject(item.Raw)
			if runtime.IsNotRegisteredError(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			r, err := s.inject(obj, pvcName, runtimeInfo)
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
		typeMeta = pod.TypeMeta
		objectMeta = pod.ObjectMeta
		application = podapp.NewApplication(pod)
	case *unstructuredtype.Unstructured:
		obj := v
		typeMeta = metav1.TypeMeta{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
		}
		objectMeta = metav1.ObjectMeta{
			Name:        obj.GetName(),
			Namespace:   obj.GetNamespace(),
			Annotations: obj.GetAnnotations(),
			Labels:      obj.GetLabels(),
		}
		application = unstructured.NewApplication(obj)
	case runtime.Object:
		obj := v
		outValue := reflect.ValueOf(obj).Elem()
		typeMeta = outValue.FieldByName("TypeMeta").Interface().(metav1.TypeMeta)
		objectMeta = outValue.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
		application = defaultapp.NewApplication(obj)
	default:
		log.Info("No supported K8s Type", "v", v)
		return out, fmt.Errorf("no supported K8s Type %v", v)
	}

	namespacedName := types.NamespacedName{
		Namespace: objectMeta.Namespace,
		Name:      objectMeta.Name,
	}
	kind := typeMeta.Kind

	pods, err := application.GetPodSpecs()
	if err != nil {
		return
	}

	for _, pod := range pods {
		metaObj, err := pod.GetMetaObject()
		if err != nil {
			return out, err
		}

		// if it's not serverless enable or injection is done, skip
		if !utils.ServerlessEnabled(metaObj.Labels) || utils.InjectSidecarDone(metaObj.Labels) {
			continue
		}

		// 1. check if the pod spec has fluid volume claim
		injectFuseContainer := true
		enableCacheDir := utils.InjectCacheDirEnabled(metaObj.Labels)
		template, err := runtimeInfo.GetTemplateToInjectForFuse(pvcName, enableCacheDir)
		if err != nil {
			return out, err
		}

		// 2. Determine if the volumeMounts contain the target pvc, if not found, skip. The reason is that if this `pod` spec doesn't have volumeMounts
		volumeMounts, err := pod.GetVolumeMounts()
		if err != nil {
			return out, err
		}

		// 3. find the volumes with the target pvc name, and replace it with the fuse's hostpath volume
		volumes, err := pod.GetVolumes()
		if err != nil {
			return out, err
		}

		pvcNames := kubeclient.PVCNames(volumeMounts, volumes)
		found := utils.ContainsString(pvcNames, pvcName)
		if !found {
			log.Info("Not able to find the fluid pvc in pod spec, skip",
				"name", objectMeta.Name,
				"namespace", objectMeta.Namespace,
				"pvc", pvcName,
				"candidate pvcs", pvcNames)
			continue
		}

		volumeNames := []string{}
		datasetVolumeNames := []string{}
		for i, volume := range volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				name := volume.Name
				volumes[i] = template.VolumesToUpdate[0]
				volumes[i].Name = name
				datasetVolumeNames = append(datasetVolumeNames, name)
			}
			volumeNames = append(volumeNames, volume.Name)
		}

		// 4. add the volumes
		// key is the old volume name, value is the new volume name
		volumeNamesConflict := map[string]string{}
		if len(template.VolumesToAdd) > 0 {
			log.V(1).Info("Before append volume", "original", volumes)
			// volumes = append(volumes, template.VolumesToAdd...)
			for _, volumeToAdd := range template.VolumesToAdd {
				if !utils.ContainsString(volumeNames, volumeToAdd.Name) {
					volumes = append(volumes, volumeToAdd)
				} else {
					i := 0
					newVolumeName := utils.ReplacePrefix(volumeToAdd.Name, common.Fluid)
					for {
						if !utils.ContainsString(volumeNames, newVolumeName) {
							break
						} else {
							if i > 100 {
								log.Info("retry  the volume name because duplicate name more than 100 times, then give up",
									"name", objectMeta.Name,
									"namespace", objectMeta.Namespace,
									"volumeName", newVolumeName)
								return out, fmt.Errorf("retry  the volume name %v for object %v because duplicate name more than 100 times, then give up", newVolumeName, types.NamespacedName{
									Namespace: objectMeta.Namespace,
									Name:      objectMeta.Name,
								})
							}
							suffix := common.Fluid + "-" + utils.RandomAlphaNumberString(3)
							newVolumeName = utils.ReplacePrefix(volumeToAdd.Name, suffix)
							log.Info("retry  the volume name because duplicate name",
								"name", objectMeta.Name,
								"namespace", objectMeta.Namespace,
								"volumeName", newVolumeName)
							i++
						}
					}

					volume := volumeToAdd.DeepCopy()
					volume.Name = newVolumeName
					volumeNamesConflict[volumeToAdd.Name] = volume.Name
					volumeNames = append(volumeNames, newVolumeName)
					// return out, err
					volumes = append(volumes, *volume)
				}
			}

			log.V(1).Info("After append volume", "original", volumes)
		}
		err = pod.SetVolumes(volumes)
		if err != nil {
			return out, err
		}

		// 5.Add sidecar as the first container
		containers, err := pod.GetContainers()
		if err != nil {
			return out, err
		}

		// Skip injection for injected container
		for _, c := range containers {
			if c.Name == common.FuseContainerName {
				warningStr := fmt.Sprintf("===> Skipping injection because %v has injected %q sidecar already\n",
					namespacedName, common.FuseContainerName)
				if len(kind) != 0 {
					warningStr = fmt.Sprintf("===> Skipping injection because Kind %s: %v has injected %q sidecar already\n",
						kind, namespacedName, common.FuseContainerName)
				}
				log.Info(warningStr)
				injectFuseContainer = false
				break
			}
		}
		fuseContainer := template.FuseContainer
		for oldName, newName := range volumeNamesConflict {
			for i, volumeMount := range fuseContainer.VolumeMounts {
				if volumeMount.Name == oldName {
					fuseContainer.VolumeMounts[i].Name = newName
				}
			}
		}
		if injectFuseContainer {
			containers = append([]corev1.Container{fuseContainer}, containers...)
		}

		// 6. Set mountPropagationHostToContainer to the dataset volume mount point
		mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
		for _, container := range containers {
			if container.Name != common.FuseContainerName {
				for i, volumeMount := range container.VolumeMounts {
					if utils.ContainsString(datasetVolumeNames, volumeMount.Name) {
						container.VolumeMounts[i].MountPropagation = &mountPropagationHostToContainer
					}
				}
			}
		}

		err = pod.SetContainers(containers)
		if err != nil {
			return out, err
		}

		// 7. Set the injection phase done to avoid re-injection
		metaObj.Labels[common.InjectSidecarDone] = common.True
		err = pod.SetMetaObject(metaObj)
		if err != nil {
			return out, err
		}
	}

	// kubeclient.IsVolumeMountForPVC(pvcName, )

	err = application.SetPodSpecs(pods)
	if err != nil {
		return out, err
	}

	out = application.GetObject()
	if err != nil {
		return out, err
	}
	return out, nil
}

func (s *Injector) Inject(in runtime.Object, runtimeInfos map[string]base.RuntimeInfoInterface) (out runtime.Object, err error) {
	for pvcName, runtimeInfo := range runtimeInfos {
		out, err = s.inject(in, pvcName, runtimeInfo)
		if err != nil {
			return
		}
		if len(runtimeInfos) > 1 {
			in = out.DeepCopyObject()
		}
	}

	return
}

func (s *Injector) InjectUnstructured(in *unstructuredtype.Unstructured, runtimeInfos map[string]base.RuntimeInfoInterface) (out *unstructuredtype.Unstructured, err error) {
	return nil, fmt.Errorf("not implemented yet")
}
