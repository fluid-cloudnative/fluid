/*
Copyright 2022 The Fluid Authors.

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

package juicefs

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// SyncRuntime syncs the runtime spec
func (j *JuiceFSEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	runtime, err := j.getRuntime()
	if err != nil {
		return changed, err
	}

	var value *JuiceFS
	value, err = j.transform(runtime)
	if err != nil {
		return
	}

	// 1. sync workers
	workerChanged, err := j.syncWorkerSpec(ctx, runtime, value)
	if err != nil {
		return
	}
	if workerChanged {
		j.Log.Info("Worker Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return workerChanged, err
	}

	// 2. sync fuse
	fuseChanged, err := j.syncFuseSpec(ctx, runtime, value)
	if err != nil {
		return
	}
	if fuseChanged {
		j.Log.Info("Fuse Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return fuseChanged, err
	}
	return
}

func (j *JuiceFSEngine) syncWorkerSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (changed bool, err error) {
	j.Log.V(1).Info("syncWorkerSpec")
	var cmdChanged bool
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(j.Client,
			types.NamespacedName{Namespace: j.namespace, Name: j.getWorkerName()})
		if err != nil {
			return err
		}

		workersToUpdate := workers.DeepCopy()

		// nodeSelector
		if nodeSelectorChanged, newSelector := j.isNodeSelectorChanged(workersToUpdate.Spec.Template.Spec.NodeSelector, value.Worker.NodeSelector); nodeSelectorChanged {
			workersToUpdate.Spec.Template.Spec.NodeSelector = newSelector
			changed = true
		}

		// volumes
		if volumeChanged, newVolumes := j.isVolumesChanged(workersToUpdate.Spec.Template.Spec.Volumes, value.Worker.Volumes); volumeChanged {
			workersToUpdate.Spec.Template.Spec.Volumes = newVolumes
			changed = true
		}

		// labels
		if labelChanged, newLabels := j.isLabelsChanged(workersToUpdate.Spec.Template.ObjectMeta.Labels, value.Worker.Labels); labelChanged {
			workersToUpdate.Spec.Template.ObjectMeta.Labels = newLabels
			changed = true
		}

		// annotations
		if annoChanged, newAnnos := j.isAnnotationsChanged(workersToUpdate.Spec.Template.ObjectMeta.Annotations, value.Worker.Annotations); annoChanged {
			workersToUpdate.Spec.Template.ObjectMeta.Annotations = newAnnos
			changed = true
		}

		// options -> configmap
		workerCommand, err := j.getWorkerCommand()
		if err != nil || workerCommand == "" {
			j.Log.Error(err, "Failed to get worker command")
			return err
		}
		cmdChanged, _ = j.isCommandChanged(workerCommand, value.Worker.Command)

		if len(workersToUpdate.Spec.Template.Spec.Containers) == 1 {
			// resource
			if resourcesChanged, newResources := j.isResourcesChanged(workersToUpdate.Spec.Template.Spec.Containers[0].Resources, runtime.Spec.Worker.Resources); resourcesChanged {
				workersToUpdate.Spec.Template.Spec.Containers[0].Resources = newResources
				changed = true
			}

			// env
			if envChanged, newEnvs := j.isEnvsChanged(workersToUpdate.Spec.Template.Spec.Containers[0].Env, value.Worker.Envs); envChanged {
				workersToUpdate.Spec.Template.Spec.Containers[0].Env = newEnvs
				changed = true
			}

			// volumeMounts
			if volumeMountChanged, newVolumeMounts := j.isVolumeMountsChanged(workersToUpdate.Spec.Template.Spec.Containers[0].VolumeMounts, value.Worker.VolumeMounts); volumeMountChanged {
				workersToUpdate.Spec.Template.Spec.Containers[0].VolumeMounts = newVolumeMounts
				changed = true
			}

			// image
			runtimeImage := value.Image
			if value.ImageTag != "" {
				runtimeImage = runtimeImage + ":" + value.ImageTag
			}
			if imageChanged, newImage := j.isImageChanged(workersToUpdate.Spec.Template.Spec.Containers[0].Image, runtimeImage); imageChanged {
				workersToUpdate.Spec.Template.Spec.Containers[0].Image = newImage
				changed = true
			}
		}

		if cmdChanged {
			j.Log.Info("The worker config is updated")
			err = j.updateWorkerScript(value.Worker.Command)
			if err != nil {
				j.Log.Error(err, "Failed to update the sts config")
				return err
			}
			if !changed {
				// if worker sts not changed, rollout worker sts to reload the script
				j.Log.Info("rollout restart worker", "sts", workersToUpdate.Name)
				workersToUpdate.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
				changed = true
			}
		} else {
			j.Log.V(1).Info("The worker config is not changed")
		}

		if changed {
			if reflect.DeepEqual(workers, workersToUpdate) {
				changed = false
				j.Log.V(1).Info("The worker is not changed, skip")
				return nil
			}
			j.Log.Info("The worker is updated")

			err = j.Client.Update(context.TODO(), workersToUpdate)
			if err != nil {
				j.Log.Error(err, "Failed to update the sts spec")
			}
		} else {
			j.Log.V(1).Info("The worker is not changed")
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		j.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}

	return
}

func (j *JuiceFSEngine) syncFuseSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (changed bool, err error) {
	j.Log.V(1).Info("syncFuseSpec")
	var cmdChanged bool
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		fuses, err := kubeclient.GetDaemonset(j.Client, j.getFuseName(), j.namespace)
		if err != nil {
			return err
		}

		fusesToUpdate := fuses.DeepCopy()

		// nodeSelector
		if nodeSelectorChanged, newSelector := j.isNodeSelectorChanged(fusesToUpdate.Spec.Template.Spec.NodeSelector, value.Fuse.NodeSelector); nodeSelectorChanged {
			j.Log.Info("syncFuseSpec nodeSelectorChanged")
			fusesToUpdate.Spec.Template.Spec.NodeSelector = newSelector
			changed = true
		}

		// volumes
		if volumeChanged, newVolumes := j.isVolumesChanged(fusesToUpdate.Spec.Template.Spec.Volumes, value.Fuse.Volumes); volumeChanged {
			j.Log.Info("syncFuseSpec volumeChanged")
			fusesToUpdate.Spec.Template.Spec.Volumes = newVolumes
			changed = true
		}

		// labels
		if labelChanged, newLabels := j.isLabelsChanged(fusesToUpdate.Spec.Template.ObjectMeta.Labels, value.Fuse.Labels); labelChanged {
			j.Log.Info("syncFuseSpec labelChanged")
			fusesToUpdate.Spec.Template.ObjectMeta.Labels = newLabels
			changed = true
		}

		// annotations
		if annoChanged, newAnnos := j.isAnnotationsChanged(fusesToUpdate.Spec.Template.ObjectMeta.Annotations, value.Fuse.Annotations); annoChanged {
			j.Log.Info("syncFuseSpec annoChanged")
			fusesToUpdate.Spec.Template.ObjectMeta.Annotations = newAnnos
			changed = true
		}

		// options -> configmap
		fuseCommand, err := j.getFuseCommand()
		if err != nil || fuseCommand == "" {
			j.Log.Error(err, "Failed to get fuse command")
			return err
		}
		cmdChanged, _ = j.isCommandChanged(fuseCommand, value.Fuse.Command)

		if len(fusesToUpdate.Spec.Template.Spec.Containers) == 1 {
			// resource
			if resourcesChanged, newResources := j.isResourcesChanged(fusesToUpdate.Spec.Template.Spec.Containers[0].Resources, runtime.Spec.Fuse.Resources); resourcesChanged {
				j.Log.Info("syncFuseSpec resourcesChanged")
				fusesToUpdate.Spec.Template.Spec.Containers[0].Resources = newResources
				changed = true
			}

			// env
			if envChanged, newEnvs := j.isEnvsChanged(fusesToUpdate.Spec.Template.Spec.Containers[0].Env, value.Fuse.Envs); envChanged {
				j.Log.Info("syncFuseSpec envChanged")
				fusesToUpdate.Spec.Template.Spec.Containers[0].Env = newEnvs
				changed = true
			}

			// volumeMounts
			if volumeMountChanged, newVolumeMounts := j.isVolumeMountsChanged(fusesToUpdate.Spec.Template.Spec.Containers[0].VolumeMounts, value.Fuse.VolumeMounts); volumeMountChanged {
				j.Log.Info("syncFuseSpec volumeMountChanged")
				fusesToUpdate.Spec.Template.Spec.Containers[0].VolumeMounts = newVolumeMounts
				changed = true
			}

			// image
			fuseImage := value.Fuse.Image
			if value.ImageTag != "" {
				fuseImage = fuseImage + ":" + value.Fuse.ImageTag
			}
			if imageChanged, newImage := j.isImageChanged(fusesToUpdate.Spec.Template.Spec.Containers[0].Image, fuseImage); imageChanged {
				j.Log.Info("syncFuseSpec imageChanged")
				fusesToUpdate.Spec.Template.Spec.Containers[0].Image = newImage
				changed = true
			}
		}

		if cmdChanged {
			j.Log.Info("The fuse config is updated")
			err = j.updateFuseScript(value.Fuse.Command)
			if err != nil {
				j.Log.Error(err, "Failed to update the ds config")
				return err
			}
		} else {
			j.Log.V(1).Info("The fuse config is not changed")
		}

		if changed {
			if reflect.DeepEqual(fuses, fusesToUpdate) {
				changed = false
				j.Log.V(1).Info("The fuse is not changed, skip")
				return nil
			}
			j.Log.Info("The fuse is updated")

			if currentGeneration, exist := fusesToUpdate.Spec.Template.Labels[common.LabelRuntimeFuseGeneration]; exist {
				currentGenerationInt, err := strconv.Atoi(currentGeneration)
				if err != nil {
					j.Log.Error(err, "Failed to parse current fuse generation from the ds label")
					return nil
				}
				newGeneration := strconv.FormatInt(int64(currentGenerationInt+1), 10)
				fusesToUpdate.Spec.Template.Labels[common.LabelRuntimeFuseGeneration] = newGeneration

				pvc, err := kubeclient.GetPersistentVolumeClaim(j.Client, j.name, j.namespace)
				if err != nil {
					return err
				}

				labelsToModify := common.LabelsToModify{}
				if _, exist := pvc.Labels[common.LabelRuntimeFuseGeneration]; exist {
					labelsToModify.Update(common.LabelRuntimeFuseGeneration, newGeneration)
				} else {
					labelsToModify.Add(common.LabelRuntimeFuseGeneration, newGeneration)
				}

				if _, err = utils.PatchLabels(j.Client, pvc, labelsToModify); err != nil {
					j.Log.Error(err, fmt.Sprintf("imageChanged but failed to update image info on pvc %s/%s", j.namespace, j.name))
				}
			}
			err = j.Client.Update(context.TODO(), fusesToUpdate)
			if err != nil {
				j.Log.Error(err, "Failed to update the ds spec")
			}

		} else {
			j.Log.V(1).Info("The fuse is not changed")
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		j.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}
	return
}

func (j *JuiceFSEngine) isVolumeMountsChanged(crtVolumeMounts, runtimeVolumeMounts []corev1.VolumeMount) (changed bool, newVolumeMounts []corev1.VolumeMount) {
	mounts := make(map[string]corev1.VolumeMount)
	for _, mount := range crtVolumeMounts {
		mounts[mount.Name] = mount
	}
	for _, mount := range runtimeVolumeMounts {
		if m, ok := mounts[mount.Name]; !ok || !reflect.DeepEqual(m, mount) {
			j.Log.Info("The volumeMounts is different.", "current sts", crtVolumeMounts, "runtime", runtimeVolumeMounts)
			mounts[mount.Name] = mount
			changed = true
		}
	}
	vms := []corev1.VolumeMount{}
	for _, mount := range mounts {
		vms = append(vms, mount)
	}
	newVolumeMounts = vms
	return
}

func (j JuiceFSEngine) isEnvsChanged(crtEnvs, runtimeEnvs []corev1.EnvVar) (changed bool, newEnvs []corev1.EnvVar) {
	envMap := make(map[string]corev1.EnvVar)
	for _, env := range crtEnvs {
		envMap[env.Name] = env
	}
	for _, env := range runtimeEnvs {
		if envMap[env.Name].Value != env.Value || !reflect.DeepEqual(envMap[env.Name].ValueFrom, env.ValueFrom) {
			j.Log.Info("The env is different.", "current sts", crtEnvs, "runtime", runtimeEnvs)
			envMap[env.Name] = env
			changed = true
		}
	}
	envs := []corev1.EnvVar{}
	for _, env := range envMap {
		envs = append(envs, env)
	}
	newEnvs = envs
	return
}

func (j JuiceFSEngine) isResourcesChanged(crtResources, runtimeResources corev1.ResourceRequirements) (changed bool, newResources corev1.ResourceRequirements) {
	if !utils.ResourceRequirementsEqual(crtResources, runtimeResources) {
		j.Log.Info("The resource requirement is different.", "current sts", crtResources, "runtime", runtimeResources)
		changed = true
	}
	newResources = runtimeResources
	return
}

func (j JuiceFSEngine) isVolumesChanged(crtVolumes, runtimeVolumes []corev1.Volume) (changed bool, newVolumes []corev1.Volume) {
	volumes := make(map[string]corev1.Volume)
	for _, mount := range crtVolumes {
		volumes[mount.Name] = mount
	}
	for _, volume := range runtimeVolumes {
		if m, ok := volumes[volume.Name]; !ok || !reflect.DeepEqual(m, volume) {
			j.Log.Info("The volumes is different.", "current sts", crtVolumes, "runtime", runtimeVolumes)
			volumes[volume.Name] = volume
			changed = true
		}
	}
	vs := []corev1.Volume{}
	for _, volume := range volumes {
		vs = append(vs, volume)
	}
	newVolumes = vs
	return
}

func (j JuiceFSEngine) isLabelsChanged(crtLabels, runtimeLabels map[string]string) (changed bool, newLabels map[string]string) {
	newLabels = crtLabels
	for k, v := range runtimeLabels {
		if crtv, ok := crtLabels[k]; !ok || crtv != v {
			j.Log.Info("The labels is different.", "current sts", crtLabels, "runtime", runtimeLabels)
			newLabels[k] = v
			changed = true
		}
	}
	return
}

func (j JuiceFSEngine) isAnnotationsChanged(crtAnnotations, runtimeAnnotations map[string]string) (changed bool, newAnnotations map[string]string) {
	newAnnotations = crtAnnotations
	for k, v := range runtimeAnnotations {
		if crtv, ok := crtAnnotations[k]; !ok || crtv != v {
			j.Log.Info("The annotations is different.", "current sts", crtAnnotations, "runtime", runtimeAnnotations)
			newAnnotations[k] = v
			changed = true
		}
	}
	return
}

func (j JuiceFSEngine) isImageChanged(crtImage, runtimeImage string) (changed bool, newImage string) {
	if crtImage != runtimeImage {
		j.Log.Info("The image is different.", "current sts", crtImage, "runtime", runtimeImage)
		changed = true
	}
	newImage = runtimeImage
	return
}

func (j JuiceFSEngine) isNodeSelectorChanged(crtNodeSelector, runtimeNodeSelector map[string]string) (changed bool, newNodeSelector map[string]string) {
	if crtNodeSelector == nil {
		crtNodeSelector = map[string]string{}
	}
	if runtimeNodeSelector == nil {
		runtimeNodeSelector = map[string]string{}
	}
	if !reflect.DeepEqual(crtNodeSelector, runtimeNodeSelector) {
		j.Log.Info("The nodeSelector is different.", "current sts", crtNodeSelector, "runtime", runtimeNodeSelector)
		changed = true
	}
	newNodeSelector = runtimeNodeSelector
	return
}

func (j JuiceFSEngine) isCommandChanged(crtCommand, runtimeCommand string) (changed bool, newCommand string) {
	getOption := func(command string) map[string]string {
		commands := strings.Split(command, "-o")
		if len(commands) == 1 {
			return map[string]string{}
		}
		options := strings.Split(commands[1], ",")
		optionMap := make(map[string]string)
		for _, option := range options {
			// ignore metrics option, because it may be different when using hostNetwork
			if strings.Contains(option, "metrics") {
				continue
			}
			o := strings.TrimSpace(option)
			os := strings.Split(o, "=")
			if len(os) == 1 {
				optionMap[o] = ""
			} else {
				optionMap[os[0]] = os[1]
			}
		}
		return optionMap
	}
	workerOption := getOption(crtCommand)
	runtimeOption := getOption(runtimeCommand)
	for k, v := range runtimeOption {
		if wv, ok := workerOption[k]; !ok || wv != v {
			j.Log.Info("The command is different.", "current sts", crtCommand, "runtime", runtimeCommand)
			changed = true
		}
	}
	newCommand = runtimeCommand
	return
}
