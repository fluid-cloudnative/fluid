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
	"os"
	"reflect"
	"strings"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e *CacheEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
	// permitSyncEngineStatus avoids frequent rpcs with engines with rate limited retries
	permitSyncEngineStatus := e.permitSync()
	if permitSyncEngineStatus {
		defer e.setTimeOfLastSync()
	}

	defer utils.TimeTrack(time.Now(), "CacheEngine.Sync", "ctx", ctx)

	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}
	runtimeClass, err := e.getRuntimeClass(runtime.Spec.RuntimeClassName)
	if err != nil {
		return err
	}

	err = e.syncRuntimeValueConfigMap(ctx, runtime)
	if err != nil {
		return err
	}

	// handle ufs change - support dynamic mount updates
	err = e.UpdateOnUFSChange(runtime)
	if err != nil {
		e.Log.Error(err, "Failed to update UFS")
		return err
	}

	// Use lightweight getRuntimeStatusValue instead of full transform for status update
	statusValue, err := e.getRuntimeStatusValue(runtime, runtimeClass)
	if err != nil {
		return err
	}

	// sync runtime spec changes to Master/Worker components
	err = e.syncRuntimeSpec(ctx, runtime, runtimeClass)
	if err != nil {
		return err
	}

	// sync runtime status
	runtimeReady, err := e.CheckAndUpdateRuntimeStatus(statusValue)
	if err != nil {
		return err
	}

	if !runtimeReady {
		// update dataset status when runtime not ready
		err = e.UpdateDatasetStatus(datav1alpha1.FailedDatasetPhase, runtime, runtimeClass)
		if err != nil {
			return err
		}
	} else if permitSyncEngineStatus {
		// sync dataset cache states when runtime is ready and sync permitted
		e.Log.Info("sync dataset cache states")
		err = e.syncDatasetCacheStates(ctx, runtime, runtimeClass)
		if err != nil {
			return err
		}
	}

	// add dataset related labels for worker nodes
	info, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}
	err = lifecycle.SyncScheduleInfoToCacheNodes(info, e.Client)
	if err != nil {
		return err
	}

	return nil
}

func (e *CacheEngine) syncRuntimeValueConfigMap(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheRuntime) error {
	configMap, err := kubeclient.GetConfigmapByNameWithContext(ctx, e.Client, common.GetCacheRuntimeConfigConfigMapName(e.name), e.namespace)
	if err != nil {
		return err
	}
	data, err := e.generateRuntimeConfigData(ctx, runtime)
	if err != nil {
		return err
	}

	var True = true
	owner := []metav1.OwnerReference{
		{
			APIVersion:         runtime.APIVersion,
			Kind:               runtime.Kind,
			Name:               runtime.Name,
			UID:                runtime.UID,
			Controller:         &True,
			BlockOwnerDeletion: &True,
		},
	}

	if configMap == nil {
		return kubeclient.CreateConfigMapWithOwnerWithContext(ctx, e.Client, common.GetCacheRuntimeConfigConfigMapName(e.name), e.namespace, data, owner)
	}

	configMapToUpdate := configMap.DeepCopy()
	configMapToUpdate.Data = data
	if !reflect.DeepEqual(configMapToUpdate, configMap) {
		err = kubeclient.UpdateConfigMapWithContext(ctx, e.Client, configMapToUpdate)
		if err != nil {
			return err
		}
	}

	return err
}
func getSyncRetryDuration() (d *time.Duration, err error) {
	if value, existed := os.LookupEnv(syncRetryDurationEnv); existed {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return d, err
		}
		d = &duration
	}
	return
}

// setTimeOfLastSync updates the synchronization timestamp for the CacheEngine.
// This function sets the internal timeOfLastSync field to the current time and
// logs the updated time value for tracking purposes.
func (e *CacheEngine) setTimeOfLastSync() {
	e.timeOfLastSync = time.Now()
	e.Log.V(1).Info("Set timeOfLastSync", "timeOfLastSync", e.timeOfLastSync)
}
func (e *CacheEngine) permitSync() (permit bool) {
	if time.Since(e.timeOfLastSync) < e.syncRetryDuration {
		permit = false
	} else {
		permit = true
	}
	return
}

// syncRuntimeSpec synchronizes CacheRuntime spec changes to component workloads
// This enables in-place updates for compatible fields without pod recreation
// Note: Only Master and Worker components support in-place update (using AdvancedStatefulSet)
// Client component uses DaemonSet and does NOT support in-place update
func (e *CacheEngine) syncRuntimeSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) error {
	e.Log.V(1).Info("start syncing runtime spec")
	defer func() {
		e.Log.V(1).Info("finished syncing runtime spec")
	}()

	// getRuntimeClass rejects a class without a topology, so this only guards direct callers;
	// like the transform path, a missing topology leaves no component to sync.
	if runtimeClass.Topology == nil {
		return nil
	}

	// Sync Master component if enabled (supports in-place update via AdvancedStatefulSet)
	if runtimeClass.Topology.Master != nil && !runtime.Spec.Master.Disabled {
		masterIdentity := &common.ComponentIdentity{
			Name:      common.GetCacheComponentName(e.name, common.ComponentTypeMaster),
			Namespace: e.namespace,
		}
		manager := component.NewComponentHelper(common.ComponentTypeMaster, e.Client)
		podSpec, err := manager.GetPodSpec(ctx.Context, masterIdentity)
		if err != nil {
			e.Log.Error(err, "failed to read the master pod spec", "component", masterIdentity.Name)
			return err
		}
		masterSpec := component.ComponentSpec{
			Version: runtime.Spec.Master.RuntimeVersion,
			Resources: e.validDesiredResources(runtime, common.ComponentTypeMaster,
				desiredComponentResources(runtime.Spec.Master.Resources, runtimeClass.Topology.Master), podSpec),
			Replicas: &runtime.Spec.Master.Replicas,
		}
		if err := manager.SyncComponentSpec(ctx.Context, masterIdentity, masterSpec); err != nil {
			e.Log.Error(err, "failed to sync master component spec", "component", masterIdentity.Name)
			return err
		}
	}

	// Sync Worker component if enabled (supports in-place update via AdvancedStatefulSet)
	if runtimeClass.Topology.Worker != nil && !runtime.Spec.Worker.Disabled {
		workerIdentity := &common.ComponentIdentity{
			Name:      common.GetCacheComponentName(e.name, common.ComponentTypeWorker),
			Namespace: e.namespace,
		}
		manager := component.NewComponentHelper(common.ComponentTypeWorker, e.Client)
		// Mirror the creation path: the tiered store memory quota is charged on top
		// of the resolved baseline, whether that came from the CacheRuntime or from
		// the CacheRuntimeClass template. See TransformRuntimeTieredStore.
		//
		// The quota is read back from the workload rather than recomputed from the
		// spec: tieredStore is not a supported update field, and SyncComponentSpec
		// only patches resources, so a spec-derived quota would move the container's
		// memory while the tmpfs volumes kept their original size.
		podSpec, err := manager.GetPodSpec(ctx.Context, workerIdentity)
		if err != nil {
			e.Log.Error(err, "failed to read the worker pod spec", "component", workerIdentity.Name)
			return err
		}
		workerResources := e.validDesiredResources(runtime, common.ComponentTypeWorker,
			desiredComponentResources(runtime.Spec.Worker.Resources, runtimeClass.Topology.Worker), podSpec)
		// nil only when the merged resources conflict: leave them unsynced.
		if workerResources != nil {
			resources := withTieredStoreMemoryQuota(*workerResources, chargedTieredStoreMemoryQuota(podSpec))
			workerResources = &resources
		}

		workerSpec := component.ComponentSpec{
			Version:   runtime.Spec.Worker.RuntimeVersion,
			Resources: workerResources,
			Replicas:  &runtime.Spec.Worker.Replicas,
		}
		if err := manager.SyncComponentSpec(ctx.Context, workerIdentity, workerSpec); err != nil {
			e.Log.Error(err, "failed to sync worker component spec", "component", workerIdentity.Name)
			return err
		}
	}

	// Note: Client component is NOT synced here because it uses DaemonSet which does not support in-place update
	// Client component will be recreated when spec changes

	return nil
}

// desiredComponentResources resolves the resources that should be synced to a
// component's workload. It mirrors the creation path: the values set on the CacheRuntime
// are overlaid key by key on the CacheRuntimeClass template baseline, so a key the
// CacheRuntime does not name keeps the template value. When neither side declares a key,
// it is cleared from the workload.
//
// Only the first container is considered, matching the creation path, which also only
// fills in resources for Containers[0].
func desiredComponentResources(runtimeResources corev1.ResourceRequirements, componentDefinition *datav1alpha1.RuntimeComponentDefinition) *corev1.ResourceRequirements {
	var templateResources corev1.ResourceRequirements
	if componentDefinition != nil && len(componentDefinition.Template.Spec.Containers) > 0 {
		templateResources = componentDefinition.Template.Spec.Containers[0].Resources
	}

	desired := mergeResourceRequirements(templateResources, runtimeResources)
	return &desired
}

func (e *CacheEngine) syncDatasetCacheStates(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) (err error) {
	cacheStates, err := e.GetCacheStates(runtime, runtimeClass)
	if err != nil {
		e.Log.Error(err, "Failed to get cache states, keeping previous cache states in dataset status")
		// not blocking the sync flow when failed to get cache states
		return nil
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()

		datasetToUpdate.Status.CacheStates = cacheStates
		datasetToUpdate.Status.FileNum = cacheStates[common.FileNum]
		datasetToUpdate.Status.UfsTotal = cacheStates[common.UfsTotal]

		if !reflect.DeepEqual(dataset.Status, datasetToUpdate.Status) {
			e.Log.Info("the dataset status", "status", datasetToUpdate.Status)
			err = e.Client.Status().Update(ctx.Context, datasetToUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return utils.LoggingErrorExceptConflict(e.Log, err, "Failed to Update dataset",
			types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}

// validDesiredResources returns the desired resources to sync to the workload whose pod spec
// is given. When the merged requests and limits conflict it emits a Warning event on the
// runtime naming the conflicting keys and returns nil, so the sync leaves the workload's
// resources as they are while image and replicas still sync. Container claims the sync would
// add without the workload's pod spec declaring them are dropped with a Warning event, and the
// rest is still synced:
// without DynamicResourceAllocation the API server ignores claims anyway, so they must not
// hold back a requests or limits edit.
func (e *CacheEngine) validDesiredResources(runtime *datav1alpha1.CacheRuntime, componentType common.ComponentType,
	desired *corev1.ResourceRequirements, podSpec *corev1.PodSpec) *corev1.ResourceRequirements {
	if desired == nil {
		return nil
	}
	if err := validateResourceRequirements(*desired); err != nil {
		e.Log.Info("Skip syncing conflicting resources", "component", componentType, "conflicts", err.Error())
		if e.Recorder != nil {
			e.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeResourcesConflict,
				"%s resources are not synced: the CacheRuntime and its CacheRuntimeClass template merge into %v",
				componentType, err)
		}
		return nil
	}

	declared, undeclared := splitDeclaredClaims(desired.Claims, podSpec)
	if len(undeclared) == 0 {
		return desired
	}
	e.Log.Info("Skip syncing claims the workload does not declare", "component", componentType, "claims", undeclared)
	if e.Recorder != nil {
		e.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeResourcesConflict,
			"%s resource claims %s are not synced: the workload's pod-level resourceClaims do not declare them; recreate the CacheRuntime to add them",
			componentType, strings.Join(undeclared, ", "))
	}
	synced := desired.DeepCopy()
	synced.Claims = declared
	return synced
}
