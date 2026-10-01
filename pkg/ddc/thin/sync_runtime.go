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

package thin

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	runtimeOpts "github.com/fluid-cloudnative/fluid/pkg/utils/runtimes/options"
)

// fuseContainerName is the name of the fuse container rendered by the thin chart,
// see charts/thin/templates/fuse/daemonset.yaml.
const fuseContainerName = "thin-fuse"

// SyncRuntime reconciles the fuse template fields of a ThinRuntime that has already been set up.
// Without it, editing a ready ThinRuntime's spec.fuse never reaches the rendered values or the fuse
// DaemonSet, because they are only ever generated once during setup.
//
// The helm values ConfigMap is treated as the last synced state: the desired state is re-rendered
// from the ThinRuntime and its ThinRuntimeProfile, diffed against that last synced state, pushed to
// the DaemonSet, and only then committed back to the ConfigMap. Doing it in that order means an
// interrupted sync is retried on the next reconciliation instead of being silently forgotten.
//
// The fuse DaemonSet uses the OnDelete update strategy, so updating its template does not restart
// running fuse pods. They pick up the new template whenever they are deleted, which is why the
// change is also surfaced as an event.
func (t *ThinEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	if runtimeOpts.ShouldSkipSyncingRuntime() {
		t.Log.V(1).Info("Skipping runtime sync due to CONTROLLER_SKIP_SYNCING_RUNTIME being enabled")
		return
	}

	runtime, err := t.getRuntime()
	if err != nil {
		return
	}

	// The engine is cached across reconciliations, so t.runtime and t.runtimeProfile may be stale.
	profile, err := utils.GetThinRuntimeProfile(t.Client, runtime.Spec.ThinRuntimeProfileName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			t.Log.Info("ThinRuntimeProfile not found, skip syncing the runtime spec",
				"profile", runtime.Spec.ThinRuntimeProfileName)
			return false, nil
		}
		return false, err
	}

	latestValue, err := t.transform(runtime, profile)
	if err != nil {
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		valueToSync, innerErr := t.GetValueFromConfigmap()
		if innerErr != nil {
			return innerErr
		}
		if valueToSync == nil {
			// The user opted out of the values ConfigMap, so there is no last synced state to diff
			// against. Degrade to not syncing rather than failing the whole reconciliation.
			t.Log.Info("Helm value configmap not found, skip syncing the runtime spec",
				"configmap", t.getHelmValuesConfigMapName())
			return nil
		}

		fuseChanged, innerErr := t.syncFuseSpec(valueToSync, latestValue)
		if innerErr != nil {
			return innerErr
		}

		changed = fuseChanged
		if !changed {
			return nil
		}

		t.Log.Info("Committing the changed value to the helm value configmap", "name", t.name, "namespace", t.namespace)
		if innerErr = t.SaveValueToConfigmap(valueToSync); innerErr != nil {
			t.Log.Error(innerErr, "failed to save the changed value to the helm value configmap")
			return innerErr
		}

		return nil
	})

	if err != nil {
		t.Log.Error(err, "Failed to sync the runtime spec")
		return false, err
	}

	if changed {
		t.recordFuseTemplateUpdated(runtime)
	}

	return
}

func (t *ThinEngine) recordFuseTemplateUpdated(runtime *datav1alpha1.ThinRuntime) {
	if t.Recorder == nil {
		return
	}
	t.Recorder.Eventf(runtime, corev1.EventTypeNormal, common.FuseTemplateUpdated,
		"Updated the template of fuse daemonset %s. Its update strategy is OnDelete, so running fuse pods keep the previous template until they are deleted.",
		t.getFuseName())
}

// syncFuseSpec pushes the differences between the last synced value and the latest value into the
// fuse DaemonSet. oldValue is advanced in place for every field that was actually pushed, so that
// the caller can commit it as the new last synced state.
func (t *ThinEngine) syncFuseSpec(oldValue, latestValue *ThinValue) (changed bool, err error) {
	t.Log.V(1).Info("entering syncFuseSpec")
	defer func() {
		t.Log.V(1).Info("exiting syncFuseSpec")
	}()

	fuses, err := kubeclient.GetDaemonset(t.Client, t.getFuseName(), t.namespace)
	if err != nil {
		return false, err
	}

	if fuses.Spec.UpdateStrategy.Type != appsv1.OnDeleteDaemonSetStrategyType {
		// Updating the template of a RollingUpdate daemonset would restart the running fuse pods and
		// break the applications mounting them, so switch the strategy first and let the resulting
		// update event trigger a new reconciliation.
		t.Log.V(1).Info("Fuse daemonset's update strategy is not safe to sync fuse spec",
			"updateStrategy", fuses.Spec.UpdateStrategy.Type)
		if err = kubeclient.UpdateDaemonSetUpdateStrategy(t.Client, fuses.Name, fuses.Namespace,
			appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType}); err != nil {
			return false, err
		}
		t.Log.Info("syncFuseSpec: successfully updated fuse daemonset's update strategy to OnDelete",
			"fuse ds", types.NamespacedName{Namespace: fuses.Namespace, Name: fuses.Name})
		return false, nil
	}

	fusesToUpdate := fuses.DeepCopy()
	changed, err = t.checkAndSetFuseChanges(oldValue, latestValue, fusesToUpdate)
	if err != nil {
		return false, err
	}
	if !changed {
		t.Log.V(1).Info("syncFuseSpec: no differences detected about fuse")
		return false, nil
	}

	if reflect.DeepEqual(fuses, fusesToUpdate) {
		t.Log.V(1).Info("syncFuseSpec: no differences detected about fuse after equality check")
		return false, nil
	}

	t.Log.Info("syncFuseSpec: some fields are changed in fuse, try to update the fuse daemonset",
		"fuse ds", types.NamespacedName{Namespace: fusesToUpdate.Namespace, Name: fusesToUpdate.Name})
	if err = t.Client.Update(context.TODO(), fusesToUpdate); err != nil {
		t.Log.Error(err, "syncFuseSpec: failed to update the fuse daemonset spec",
			"fuse ds", types.NamespacedName{Namespace: fusesToUpdate.Namespace, Name: fusesToUpdate.Name})
		return false, err
	}

	return true, nil
}

// checkAndSetFuseChanges applies the supported fuse template changes onto fusesToUpdate.
//
// Fields the thin chart renders verbatim (resources, image, imagePullPolicy) are compared against
// the live daemonset, so that manual drift is corrected as well. Fields the chart merges with
// entries of its own (envs, volumes, volumeMounts, labels, annotations) are compared against the
// last synced value instead, and only the value-derived entries are replaced, so that the chart's
// own entries survive.
//
// Some fuse fields are deliberately left out:
//   - nodeSelector, because transformFuse injects the fuse scheduling label that CSI relies on to
//     place fuse pods, so changing it after creation breaks mounting.
//   - hostNetwork, hostPID, targetPath, ports, command, args and the probes, whose post-ready change
//     semantics need their own discussion.
//   - configValue, which updateFuseConfigOnChange already reconciles.
func (t *ThinEngine) checkAndSetFuseChanges(oldValue, latestValue *ThinValue, fusesToUpdate *appsv1.DaemonSet) (changed bool, err error) {
	// volumes
	if !isSliceEqual(oldValue.Fuse.Volumes, latestValue.Fuse.Volumes) {
		t.Log.Info("syncFuseSpec: volumes changed", "old", oldValue.Fuse.Volumes, "new", latestValue.Fuse.Volumes)
		fusesToUpdate.Spec.Template.Spec.Volumes = append(
			utils.GetVolumesDifference(fusesToUpdate.Spec.Template.Spec.Volumes, oldValue.Fuse.Volumes),
			latestValue.Fuse.Volumes...)
		oldValue.Fuse.Volumes = latestValue.Fuse.Volumes
		changed = true
	}

	// labels
	if !isMapEqual(oldValue.Fuse.Labels, latestValue.Fuse.Labels) {
		t.Log.Info("syncFuseSpec: labels changed", "old", oldValue.Fuse.Labels, "new", latestValue.Fuse.Labels)
		fusesToUpdate.Spec.Template.Labels = utils.UnionMapsWithOverride(
			utils.GetMapsDifference(fusesToUpdate.Spec.Template.Labels, oldValue.Fuse.Labels),
			latestValue.Fuse.Labels)
		oldValue.Fuse.Labels = latestValue.Fuse.Labels
		changed = true
	}

	// annotations
	if !isMapEqual(oldValue.Fuse.Annotations, latestValue.Fuse.Annotations) {
		t.Log.Info("syncFuseSpec: annotations changed", "old", oldValue.Fuse.Annotations, "new", latestValue.Fuse.Annotations)
		fusesToUpdate.Spec.Template.Annotations = utils.UnionMapsWithOverride(
			utils.GetMapsDifference(fusesToUpdate.Spec.Template.Annotations, oldValue.Fuse.Annotations),
			latestValue.Fuse.Annotations)
		oldValue.Fuse.Annotations = latestValue.Fuse.Annotations
		changed = true
	}

	containerIdx := utils.GetContainerIndex(fusesToUpdate.Spec.Template.Spec.Containers, fuseContainerName)
	if containerIdx < 0 {
		t.Log.Info("syncFuseSpec: fuse container not found in the fuse daemonset, skip syncing the container spec",
			"container", fuseContainerName)
		return changed, nil
	}
	container := &fusesToUpdate.Spec.Template.Spec.Containers[containerIdx]

	// resources
	latestResources, err := utils.TransformInternalResourcesToCoreV1Resources(latestValue.Fuse.Resources)
	if err != nil {
		return false, err
	}
	if !utils.ResourceRequirementsEqual(container.Resources, latestResources) {
		t.Log.Info("syncFuseSpec: resources changed", "old", container.Resources, "new", latestResources)
		container.Resources = latestResources
		oldValue.Fuse.Resources = latestValue.Fuse.Resources
		changed = true
	}

	// image
	if latestImage := composeImage(latestValue.Fuse.Image, latestValue.Fuse.ImageTag); container.Image != latestImage {
		t.Log.Info("syncFuseSpec: image changed", "old", container.Image, "new", latestImage)
		container.Image = latestImage
		oldValue.Fuse.Image = latestValue.Fuse.Image
		oldValue.Fuse.ImageTag = latestValue.Fuse.ImageTag
		changed = true
	}

	// imagePullPolicy
	// An empty value means the chart falls back to the Kubernetes default, so leave the daemonset
	// alone instead of clearing a policy that is already in effect.
	if latestPullPolicy := corev1.PullPolicy(latestValue.Fuse.ImagePullPolicy); latestPullPolicy != "" &&
		container.ImagePullPolicy != latestPullPolicy {
		t.Log.Info("syncFuseSpec: image pull policy changed", "old", container.ImagePullPolicy, "new", latestPullPolicy)
		container.ImagePullPolicy = latestPullPolicy
		oldValue.Fuse.ImagePullPolicy = latestValue.Fuse.ImagePullPolicy
		changed = true
	}

	// envs
	if !isSliceEqual(oldValue.Fuse.Envs, latestValue.Fuse.Envs) {
		t.Log.Info("syncFuseSpec: env variables changed", "old", oldValue.Fuse.Envs, "new", latestValue.Fuse.Envs)
		container.Env = append(
			utils.GetEnvsDifference(container.Env, oldValue.Fuse.Envs),
			latestValue.Fuse.Envs...)
		oldValue.Fuse.Envs = latestValue.Fuse.Envs
		changed = true
	}

	// volumeMounts
	if !isSliceEqual(oldValue.Fuse.VolumeMounts, latestValue.Fuse.VolumeMounts) {
		t.Log.Info("syncFuseSpec: volume mounts changed", "old", oldValue.Fuse.VolumeMounts, "new", latestValue.Fuse.VolumeMounts)
		container.VolumeMounts = append(
			utils.GetVolumeMountsDifference(container.VolumeMounts, oldValue.Fuse.VolumeMounts),
			latestValue.Fuse.VolumeMounts...)
		oldValue.Fuse.VolumeMounts = latestValue.Fuse.VolumeMounts
		changed = true
	}

	// lifecycle
	if !reflect.DeepEqual(oldValue.Fuse.Lifecycle, latestValue.Fuse.Lifecycle) {
		t.Log.Info("syncFuseSpec: lifecycle changed", "old", oldValue.Fuse.Lifecycle, "new", latestValue.Fuse.Lifecycle)
		container.Lifecycle = latestValue.Fuse.Lifecycle
		oldValue.Fuse.Lifecycle = latestValue.Fuse.Lifecycle
		changed = true
	}

	return changed, nil
}

func composeImage(image, imageTag string) string {
	if len(imageTag) == 0 {
		return image
	}
	return image + ":" + imageTag
}

// isSliceEqual treats a nil slice and an empty slice as equal, because a value that round trips
// through the values ConfigMap loses that distinction.
func isSliceEqual[T any](a, b []T) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// isMapEqual treats a nil map and an empty map as equal, for the same reason as isSliceEqual.
func isMapEqual(a, b map[string]string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
