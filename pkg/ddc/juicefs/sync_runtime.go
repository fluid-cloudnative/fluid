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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
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

	var latestValue *JuiceFS
	latestValue, err = j.transform(runtime)
	if err != nil {
		return
	}

	// Syncing the runtime spec in a atomic way: if anything unexpected happens in the middle, the process can be retry and inconsitency will be fixed.
	// 1. get old value from configmap
	// 2. sync worker spec given old value, latest value, and runtime spec. Meanwhile, old value will be updated to match what has been synced.
	// 3. sync fuse spec given old value, latest value, and runtime spec. Meanwhile, old value will be updated to match what has been synced.
	// 4. Commit value changes to complete the process

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		valueToSync, innerErr := j.GetValueFromConfigmap()
		if innerErr != nil {
			return innerErr
		}

		// 1. sync workers. syncWorkerSpec should not have a retryOnConflict logic because we want the process to be atomic
		workerChanged, innerErr := j.syncWorkerSpec(ctx, runtime, valueToSync, latestValue)
		if innerErr != nil {
			return innerErr
		}
		if workerChanged {
			j.Log.Info("Worker Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		}

		// 2. sync fuse. syncFuseSpec should not have a retryOnConflict logic because we want the process to be atomic
		fuseChanged, innerErr := j.syncFuseSpec(ctx, runtime, valueToSync, latestValue)
		if innerErr != nil {
			return innerErr
		}
		if fuseChanged {
			j.Log.Info("Fuse Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		}

		if workerChanged || fuseChanged {
			j.Log.Info("Committing changed value to configmap", "name", ctx.Name, "namespace", ctx.Namespace)
			if innerErr = j.SaveValueToConfigmap(valueToSync); innerErr != nil {
				j.Log.Error(innerErr, "failed to save changed value to configmap")
				return innerErr
			}
		}

		return nil
	})

	if err != nil {
		j.Log.Error(err, "Failed to update runtime")
		return false, err
	}

	return
}

func (j *JuiceFSEngine) syncWorkerSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JuiceFSRuntime, oldValue, latestValue *JuiceFS) (changed bool, err error) {
	j.Log.V(1).Info("entering syncWorkerSpec")
	defer func() {
		j.Log.V(1).Info("exiting syncWorkerSpec")
	}()
	var cmdChanged bool
	workers, err := ctrl.GetWorkersAsStatefulset(j.Client,
		types.NamespacedName{Namespace: j.namespace, Name: j.getWorkerName()})
	if err != nil {
		return
	}

	if workers.Spec.UpdateStrategy.Type != appsv1.OnDeleteStatefulSetStrategyType {
		j.Log.V(1).Info("Worker Sts's update strategy is not safe to sync worker spec, skipping", "updateStrategy", workers.Spec.UpdateStrategy.Type)
		return
	}

	workersToUpdate := workers.DeepCopy()

	changed = j.checkAndSetWorkerChanges(oldValue, latestValue, runtime, workersToUpdate)

	// options -> configmap
	workerCommand, err := j.getWorkerCommand()
	if err != nil || workerCommand == "" {
		j.Log.Error(err, "Failed to get worker command")
		return
	}
	cmdChanged, _ = j.isCommandChanged(workerCommand, latestValue.Worker.Command)

	if cmdChanged {
		j.Log.Info("syncWorkerSpec: the worker command or options are updated, trying to update worker config")
		err = j.updateWorkerScript(latestValue.Worker.Command)
		if err != nil {
			j.Log.Error(err, "Failed to update the sts config")
			return
		}
		oldValue.Worker.Command = latestValue.Worker.Command
		if !changed {
			// if worker sts not changed, rollout worker sts to reload the script
			j.Log.Info("syncWorkerSpec: rollout restart worker", "sts", workersToUpdate.Name)
			if len(workersToUpdate.Spec.Template.ObjectMeta.Annotations) == 0 {
				workersToUpdate.Spec.Template.ObjectMeta.Annotations = map[string]string{}
			}
			workersToUpdate.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
			changed = true
		}
	} else {
		j.Log.V(1).Info("syncWorkerSpec: the worker config is not changed")
	}

	if changed {
		if reflect.DeepEqual(workers, workersToUpdate) {
			changed = false
			j.Log.V(1).Info("syncWorkerSpec: no differences detected about worker after equality check", "worker sts", types.NamespacedName{Namespace: workersToUpdate.Namespace, Name: workersToUpdate.Name})
			return
		}

		j.Log.Info("syncWorkerSpec: some fields are changed in worker, try to update worker sts", "worker sts", types.NamespacedName{Namespace: workersToUpdate.Namespace, Name: workersToUpdate.Name})
		err = j.Client.Update(context.TODO(), workersToUpdate)
		if err != nil {
			j.Log.Error(err, "failed to update the sts spec")
			return
		}
	} else {
		j.Log.V(1).Info("syncWorkerSpec: no differences detected about worker", "worker sts", types.NamespacedName{Namespace: workersToUpdate.Namespace, Name: workersToUpdate.Name})
	}

	return
}

func (j *JuiceFSEngine) checkAndSetWorkerChanges(oldValue, latestValue *JuiceFS, runtime *datav1alpha1.JuiceFSRuntime, workersToUpdate *appsv1.StatefulSet) (workerChanged bool) {
	// nodeSelector
	if nodeSelectorChanged, newSelector := j.isNodeSelectorChanged(oldValue.Worker.NodeSelector, latestValue.Worker.NodeSelector); nodeSelectorChanged {
		j.Log.Info("syncWorkerSpec: node selector changed", "old", oldValue.Worker.NodeSelector, "new", newSelector)
		workersToUpdate.Spec.Template.Spec.NodeSelector =
			utils.UnionMapsWithOverride(utils.GetMapsDifference(workersToUpdate.Spec.Template.Spec.NodeSelector, oldValue.Worker.NodeSelector), newSelector)
		oldValue.Worker.NodeSelector = latestValue.Worker.NodeSelector
		workerChanged = true
	}

	// volumes
	if volumeChanged, newVolumes := j.isVolumesChanged(oldValue.Worker.Volumes, latestValue.Worker.Volumes); volumeChanged {
		j.Log.Info("syncWorkerSpec: volumes changed", "old", oldValue.Worker.Volumes, "new", newVolumes)
		workersToUpdate.Spec.Template.Spec.Volumes = append(utils.GetVolumesDifference(workersToUpdate.Spec.Template.Spec.Volumes, oldValue.Worker.Volumes), newVolumes...)
		oldValue.Worker.Volumes = latestValue.Worker.Volumes
		workerChanged = true
	}

	// labels
	if labelChanged, newLabels := j.isLabelsChanged(oldValue.Worker.Labels, latestValue.Worker.Labels); labelChanged {
		j.Log.Info("syncWorkerSpec: labels changed", "old", oldValue.Worker.Labels, "new", newLabels)
		workersToUpdate.Spec.Template.ObjectMeta.Labels =
			utils.UnionMapsWithOverride(utils.GetMapsDifference(workersToUpdate.Spec.Template.ObjectMeta.Labels, oldValue.Worker.Labels), newLabels)
		oldValue.Worker.Labels = latestValue.Worker.Labels
		workerChanged = true
	}

	// annotations
	if annoChanged, newAnnos := j.isAnnotationsChanged(oldValue.Worker.Annotations, latestValue.Worker.Annotations); annoChanged {
		j.Log.Info("syncWorkerSpec: annotations changed", "old", oldValue.Worker.Annotations, "new", newAnnos)
		workersToUpdate.Spec.Template.ObjectMeta.Annotations =
			utils.UnionMapsWithOverride(utils.GetMapsDifference(workersToUpdate.Spec.Template.ObjectMeta.Annotations, oldValue.Worker.Annotations), newAnnos)
		oldValue.Worker.Annotations = latestValue.Worker.Annotations
		workerChanged = true
	}

	containerIdx := -1
	for i, ctr := range workersToUpdate.Spec.Template.Spec.Containers {
		if ctr.Name == JuiceFSWorkerContainerName {
			containerIdx = i
			break
		}
	}

	if containerIdx >= 0 {
		// resource
		// TODO: check if we can simply compare worker's resources and runtime.spec.worker.resources
		if resourcesChanged, newResources := j.isResourcesChanged(workersToUpdate.Spec.Template.Spec.Containers[containerIdx].Resources, runtime.Spec.Worker.Resources); resourcesChanged {
			j.Log.Info("syncWorkerSpec: resources changed", "old", workersToUpdate.Spec.Template.Spec.Containers[containerIdx].Resources, "new", newResources)
			workersToUpdate.Spec.Template.Spec.Containers[containerIdx].Resources = newResources
			workerChanged = true
		}

		// env
		if envChanged, newEnvs := j.isEnvsChanged(oldValue.Worker.Envs, latestValue.Worker.Envs); envChanged {
			j.Log.Info("syncWorkerSpec: env variables changed", "old", oldValue.Worker.Envs, "new", newEnvs)
			workersToUpdate.Spec.Template.Spec.Containers[containerIdx].Env =
				append(utils.GetEnvsDifference(workersToUpdate.Spec.Template.Spec.Containers[containerIdx].Env, oldValue.Worker.Envs), newEnvs...)
			oldValue.Worker.Envs = latestValue.Worker.Envs
			workerChanged = true
		}

		// volumeMounts
		if volumeMountChanged, newVolumeMounts := j.isVolumeMountsChanged(oldValue.Worker.VolumeMounts, latestValue.Worker.VolumeMounts); volumeMountChanged {
			j.Log.Info("syncWorkerSpec: volume mounts changed", "old", oldValue.Worker.VolumeMounts, "new", newVolumeMounts)
			workersToUpdate.Spec.Template.Spec.Containers[containerIdx].VolumeMounts =
				append(utils.GetVolumeMountsDifference(workersToUpdate.Spec.Template.Spec.Containers[containerIdx].VolumeMounts,
					oldValue.Worker.VolumeMounts), newVolumeMounts...)
			oldValue.Worker.VolumeMounts = latestValue.Worker.VolumeMounts
			workerChanged = true
		}

		// image
		// For image, we assume once image/imageTag is set, it shall not be removed by user.
		// It's hard for Fluid to detect the removal and find a way to rollout image back to the default image.
		if len(runtime.Spec.JuiceFSVersion.Image) == 0 && len(runtime.Spec.JuiceFSVersion.ImageTag) == 0 {
			// Do not touch image info because user are using the default image
			j.Log.Info("syncWorkerSpec: no user-defined image info on Runtime, skip syncing image")
		} else {
			latestWorkerImage := latestValue.Image
			if latestValue.ImageTag != "" {
				latestWorkerImage = latestWorkerImage + ":" + latestValue.ImageTag
			}

			oldWorkerImage := oldValue.Image
			if oldValue.ImageTag != "" {
				oldWorkerImage = oldWorkerImage + ":" + oldValue.ImageTag
			}

			if imageChanged, newImage := j.isImageChanged(oldWorkerImage, latestWorkerImage); imageChanged {
				j.Log.Info("syncWorkerSpec: image changed", "old", oldWorkerImage, "new", newImage)
				workersToUpdate.Spec.Template.Spec.Containers[containerIdx].Image = newImage
				oldValue.Image = latestValue.Image
				oldValue.ImageTag = latestValue.ImageTag
				workerChanged = true
			}
		}
	}

	return
}

func (j *JuiceFSEngine) syncFuseSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JuiceFSRuntime, oldValue, latestValue *JuiceFS) (bool, error) {
	j.Log.V(1).Info("entering syncFuseSpec")
	defer func() {
		j.Log.V(1).Info("exiting syncFuseSpec")
	}()

	//1. check if fuse cmd configmap needs to update
	if err := j.updateFuseCmdConfigmapOnChanged(oldValue, latestValue); err != nil {
		return false, err
	}

	//2. check if fuse needs to update
	var fuseChanged, fuseGenerationNeedIncrease bool
	fuses, err := kubeclient.GetDaemonset(j.Client, j.getFuseName(), j.namespace)
	if err != nil {
		return false, err
	}
	fusesToUpdate := fuses.DeepCopy()

	fuseChanged, fuseGenerationNeedIncrease = j.checkAndSetFuseChanges(oldValue, latestValue, runtime, fusesToUpdate)
	if !fuseChanged {
		j.Log.V(1).Info("syncFuseSpec: no differences detected about fuse")
		return fuseChanged, nil
	}

	if fuseGenerationNeedIncrease {
		err := j.increaseFuseGeneration(fusesToUpdate)
		if err != nil {
			j.Log.Error(err, "syncFuseSpec: failed to update the fuse generation on fuse daemonset", "fuse ds", types.NamespacedName{Namespace: fusesToUpdate.Namespace, Name: fusesToUpdate.Name})
			return fuseChanged, err
		}
	}

	if reflect.DeepEqual(fuses, fusesToUpdate) {
		fuseChanged = false
		j.Log.V(1).Info("syncFuseSpec: no differences detected about fuse after equality check")
		return fuseChanged, nil
	}
	j.Log.Info("syncFuseSpec: some fields are changed in fuse, try to update fuse daemonset", "fuse ds", types.NamespacedName{Namespace: fusesToUpdate.Namespace, Name: fusesToUpdate.Name})

	if err := j.Client.Update(context.TODO(), fusesToUpdate); err != nil {
		j.Log.Error(err, "syncFuseSpec: failed to update the ds spec", "fuse ds", types.NamespacedName{Namespace: fusesToUpdate.Namespace, Name: fusesToUpdate.Name})
		return fuseChanged, err
	}

	return fuseChanged, nil
}

// TODO: move the default configurations defined in helm fuse template to the logic of transformFuse,
// ensuring that checkAndSetFuseChanges don't need to care about the configuration in actual daemonset
func (j *JuiceFSEngine) checkAndSetFuseChanges(oldValue, latestValue *JuiceFS, runtime *datav1alpha1.JuiceFSRuntime, fusesToUpdate *appsv1.DaemonSet) (fuseChanged bool, fuseGenerationNeedUpdate bool) {
	// nodeSelector
	if nodeSelectorChanged, newSelector := j.isNodeSelectorChanged(oldValue.Fuse.NodeSelector, latestValue.Fuse.NodeSelector); nodeSelectorChanged {
		j.Log.Info("syncFuseSpec: node selector changed", "old", oldValue.Fuse.NodeSelector, "new", newSelector)
		fusesToUpdate.Spec.Template.Spec.NodeSelector =
			utils.UnionMapsWithOverride(utils.GetMapsDifference(fusesToUpdate.Spec.Template.Spec.NodeSelector, oldValue.Fuse.NodeSelector), newSelector)
		oldValue.Fuse.NodeSelector = latestValue.Fuse.NodeSelector
		fuseChanged = true
	}

	// volumes
	if volumeChanged, newVolumes := j.isVolumesChanged(oldValue.Fuse.Volumes, latestValue.Fuse.Volumes); volumeChanged {
		j.Log.Info("syncFuseSpec: volume changed", "old", oldValue.Fuse.Volumes, "new", newVolumes)
		fusesToUpdate.Spec.Template.Spec.Volumes =
			append(utils.GetVolumesDifference(fusesToUpdate.Spec.Template.Spec.Volumes, oldValue.Fuse.Volumes), newVolumes...)
		oldValue.Fuse.Volumes = latestValue.Fuse.Volumes
		fuseChanged = true
	}

	// labels
	if labelChanged, newLabels := j.isLabelsChanged(oldValue.Fuse.Labels, latestValue.Fuse.Labels); labelChanged {
		j.Log.Info("syncFuseSpec: labels changed", "old", oldValue.Fuse.Labels, "new", newLabels)
		fusesToUpdate.Spec.Template.ObjectMeta.Labels =
			utils.UnionMapsWithOverride(utils.GetMapsDifference(fusesToUpdate.Spec.Template.ObjectMeta.Labels, oldValue.Fuse.Labels), newLabels)
		oldValue.Fuse.Labels = latestValue.Fuse.Labels
		fuseChanged = true
	}

	// annotations
	if annoChanged, newAnnos := j.isAnnotationsChanged(oldValue.Fuse.Annotations, latestValue.Fuse.Annotations); annoChanged {
		j.Log.Info("syncFuseSpec annotations changed", "old", oldValue.Fuse.Annotations, "new", newAnnos)
		fusesToUpdate.Spec.Template.ObjectMeta.Annotations =
			utils.UnionMapsWithOverride(utils.GetMapsDifference(fusesToUpdate.Spec.Template.ObjectMeta.Annotations, oldValue.Fuse.Annotations), newAnnos)
		oldValue.Fuse.Annotations = latestValue.Fuse.Annotations
		fuseChanged = true
	}

	containerIdx := -1
	for i, ctr := range fusesToUpdate.Spec.Template.Spec.Containers {
		if ctr.Name == JuiceFSFuseContainerName {
			containerIdx = i
			break
		}
	}

	if containerIdx >= 0 {
		// resource
		if resourcesChanged, newResources := j.isResourcesChanged(fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].Resources, runtime.Spec.Fuse.Resources); resourcesChanged {
			j.Log.Info("syncFuseSpec: resources changed", "old", fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].Resources, "new", newResources)
			fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].Resources = newResources
			fuseChanged = true
		}

		// env
		if envChanged, newEnvs := j.isEnvsChanged(oldValue.Fuse.Envs, latestValue.Fuse.Envs); envChanged {
			j.Log.Info("syncFuseSpec: env variables changed", "old", oldValue.Fuse.Envs, "new", newEnvs)
			fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].Env =
				append(utils.GetEnvsDifference(fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].Env, oldValue.Fuse.Envs), newEnvs...)
			oldValue.Fuse.Envs = latestValue.Fuse.Envs
			fuseChanged = true
		}

		// volumeMounts
		if volumeMountChanged, newVolumeMounts := j.isVolumeMountsChanged(oldValue.Fuse.VolumeMounts, latestValue.Fuse.VolumeMounts); volumeMountChanged {
			j.Log.Info("syncFuseSpec: volume mounts changed", "old", oldValue.Fuse.VolumeMounts, "new", newVolumeMounts)
			fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].VolumeMounts =
				append(utils.GetVolumeMountsDifference(fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].VolumeMounts,
					oldValue.Fuse.VolumeMounts), newVolumeMounts...)
			oldValue.Fuse.VolumeMounts = latestValue.Fuse.VolumeMounts
			fuseChanged = true
		}

		// image
		// For image, we assume once image/imageTag is set, it shall not be removed by user.
		// It's hard for Fluid to detect the removal and find a way to rollout image back to the default image.
		if len(runtime.Spec.Fuse.Image) == 0 && len(runtime.Spec.Fuse.ImageTag) == 0 {
			// Do not touch image info because user are using the default image
			j.Log.Info("syncFuseSpec: no user-defined image info on Runtime, skip syncing image")
		} else {
			latestFuseImage := latestValue.Fuse.Image
			if latestValue.Fuse.ImageTag != "" {
				latestFuseImage = latestFuseImage + ":" + latestValue.Fuse.ImageTag
			}

			currentFuseImage := oldValue.Fuse.Image
			if oldValue.Fuse.ImageTag != "" {
				currentFuseImage = currentFuseImage + ":" + oldValue.Fuse.ImageTag
			}

			if imageChanged, newImage := j.isImageChanged(currentFuseImage, latestFuseImage); imageChanged {
				j.Log.Info("syncFuseSpec: image changed", "old", currentFuseImage, "new", latestFuseImage)
				fusesToUpdate.Spec.Template.Spec.Containers[containerIdx].Image = newImage
				oldValue.Fuse.Image = latestValue.Fuse.Image
				oldValue.Fuse.ImageTag = latestValue.Fuse.ImageTag
				fuseChanged, fuseGenerationNeedUpdate = true, true
			}
		}
	}

	return fuseChanged, fuseGenerationNeedUpdate
}

func (j *JuiceFSEngine) updateFuseCmdConfigmapOnChanged(oldValue, latestValue *JuiceFS) error {
	// options -> configmap
	fuseCommand, err := j.getFuseCommand()
	if err != nil {
		j.Log.Error(err, "Failed to get fuse command")
		return err
	}

	if len(fuseCommand) == 0 {
		j.Log.Info("cannot get fuse command, got an empty fuse command, skip updating Fuse command configmap")
		return nil
	}

	if cmdChanged, _ := j.isCommandChanged(fuseCommand, latestValue.Fuse.Command); cmdChanged {
		j.Log.Info("The fuse config is updated")
		if err := j.updateFuseScript(latestValue.Fuse.Command); err != nil {
			j.Log.Error(err, "Failed to update the ds config")
			return err
		}
		oldValue.Fuse.Command = latestValue.Fuse.Command
		return nil
	}
	j.Log.V(1).Info("The fuse config is not changed")
	return nil
}

func (j *JuiceFSEngine) increaseFuseGeneration(fusesToUpdate *appsv1.DaemonSet) error {
	newGeneration := "1"
	currentGeneration, exist := fusesToUpdate.Spec.Template.Labels[common.LabelRuntimeFuseGeneration]
	if exist {
		currentGenerationInt, err := strconv.Atoi(currentGeneration)
		if err != nil {
			j.Log.Error(err, "Failed to parse current fuse generation from the ds label")
			return nil
		}
		newGeneration = strconv.FormatInt(int64(currentGenerationInt+1), 10)
	}

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
	return nil
}

func (j *JuiceFSEngine) isVolumeMountsChanged(crtVolumeMounts, runtimeVolumeMounts []corev1.VolumeMount) (changed bool, newVolumeMounts []corev1.VolumeMount) {
	newVolumeMounts = runtimeVolumeMounts
	if len(crtVolumeMounts) == 0 && len(runtimeVolumeMounts) == 0 {
		return
	}

	if !reflect.DeepEqual(crtVolumeMounts, runtimeVolumeMounts) {
		changed = true
	}

	return
}

func (j JuiceFSEngine) isEnvsChanged(crtEnvs, runtimeEnvs []corev1.EnvVar) (changed bool, newEnvs []corev1.EnvVar) {
	// TODO: Be careful of flaky detection. Caller should make sure the envs are arranged in a deterministic way (for example sorted).
	// Currently JuiceFS transform env variable in a deterministic way.
	newEnvs = runtimeEnvs
	if len(crtEnvs) == 0 && len(runtimeEnvs) == 0 {
		return
	}

	if !reflect.DeepEqual(crtEnvs, runtimeEnvs) {
		changed = true
	}

	return
}

func (j JuiceFSEngine) isResourcesChanged(crtResources, runtimeResources corev1.ResourceRequirements) (changed bool, newResources corev1.ResourceRequirements) {
	newResources = runtimeResources
	if !utils.ResourceRequirementsEqual(crtResources, runtimeResources) {
		changed = true
	}
	return
}

func (j JuiceFSEngine) isVolumesChanged(crtVolumes, runtimeVolumes []corev1.Volume) (changed bool, newVolumes []corev1.Volume) {
	newVolumes = runtimeVolumes

	// handle cases where nil slice equals to empty slice
	if len(crtVolumes) == 0 && len(runtimeVolumes) == 0 {
		return
	}

	if !reflect.DeepEqual(crtVolumes, runtimeVolumes) {
		changed = true
	}
	return
}

func (j JuiceFSEngine) isLabelsChanged(crtLabels, runtimeLabels map[string]string) (changed bool, newLabels map[string]string) {
	newLabels = runtimeLabels
	// handle cases where nil map equals to empty map
	if len(crtLabels) == 0 && len(runtimeLabels) == 0 {
		return
	}

	if !reflect.DeepEqual(crtLabels, runtimeLabels) {
		changed = true
	}
	return
}

func (j JuiceFSEngine) isAnnotationsChanged(crtAnnotations, runtimeAnnotations map[string]string) (changed bool, newAnnotations map[string]string) {
	newAnnotations = runtimeAnnotations
	// handle cases where nil map equals to empty map
	if len(crtAnnotations) == 0 && len(runtimeAnnotations) == 0 {
		return
	}

	if !reflect.DeepEqual(crtAnnotations, runtimeAnnotations) {
		changed = true
	}

	return
}

func (j JuiceFSEngine) isImageChanged(crtImage, runtimeImage string) (changed bool, newImage string) {
	newImage = runtimeImage
	if crtImage != runtimeImage {
		changed = true
	}
	return
}

func (j JuiceFSEngine) isNodeSelectorChanged(crtNodeSelector, runtimeNodeSelector map[string]string) (changed bool, newNodeSelector map[string]string) {
	newNodeSelector = runtimeNodeSelector
	if len(crtNodeSelector) == 0 && len(runtimeNodeSelector) == 0 {
		return
	}

	if !reflect.DeepEqual(crtNodeSelector, runtimeNodeSelector) {
		changed = true
	}
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
	if len(workerOption) != len(runtimeOption) {
		j.Log.Info("The command is different.", "current sts", crtCommand, "runtime", runtimeCommand)
		changed = true
	} else {
		for k, v := range runtimeOption {
			if wv, ok := workerOption[k]; !ok || wv != v {
				j.Log.Info("The command is different.", "current sts", crtCommand, "runtime", runtimeCommand)
				changed = true
			}
		}
	}
	newCommand = runtimeCommand
	return
}
