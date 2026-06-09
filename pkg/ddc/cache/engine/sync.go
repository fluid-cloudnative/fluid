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
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *CacheEngine) Sync(ctx cruntime.ReconcileRequestContext) (err error) {
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

	// TODO: implement other logic like inplace update and replica scaling

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
	_, err = e.CheckAndUpdateRuntimeStatus(statusValue)
	if err != nil {
		return err
	}

	// handle runtime spec change

	// sync metadata

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
	configMap, err := kubeclient.GetConfigmapByNameWithContext(ctx, e.Client, e.getRuntimeConfigConfigMapName(), e.namespace)
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
		return kubeclient.CreateConfigMapWithOwnerWithContext(ctx, e.Client, e.getRuntimeConfigConfigMapName(), e.namespace, data, owner)
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

// syncRuntimeSpec synchronizes CacheRuntime spec changes to component workloads
// This enables in-place updates for compatible fields without pod recreation
// Note: Only Master and Worker components support in-place update (using AdvancedStatefulSet)
// Client component uses DaemonSet and does NOT support in-place update
func (e *CacheEngine) syncRuntimeSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) error {
	e.Log.V(1).Info("start syncing runtime spec")
	defer func() {
		e.Log.V(1).Info("finished syncing runtime spec")
	}()

	// Sync Master component if enabled (supports in-place update via AdvancedStatefulSet)
	if runtimeClass.Topology.Master != nil && !runtime.Spec.Master.Disabled {
		masterIdentity := &common.ComponentIdentity{
			Name:      GetComponentName(e.name, common.ComponentTypeMaster),
			Namespace: e.namespace,
		}
		manager := component.NewComponentHelper(common.ComponentTypeMaster, e.Client)
		// Only sync resources if they are explicitly set (not zero-value)
		// This prevents overwriting template defaults when user hasn't specified resources
		var resources corev1.ResourceRequirements
		if runtime.Spec.Master.Resources.Requests != nil || runtime.Spec.Master.Resources.Limits != nil {
			resources = runtime.Spec.Master.Resources
		}
		masterSpec := component.ComponentSpec{
			Version:   runtime.Spec.Master.RuntimeVersion,
			Resources: resources,
			Replicas:  &runtime.Spec.Master.Replicas,
		}
		if err := manager.SyncComponentSpec(ctx.Context, masterIdentity, masterSpec); err != nil {
			e.Log.Error(err, "failed to sync master component spec", "component", masterIdentity.Name)
			return err
		}
	}

	// Sync Worker component if enabled (supports in-place update via AdvancedStatefulSet)
	if runtimeClass.Topology.Worker != nil && !runtime.Spec.Worker.Disabled {
		workerIdentity := &common.ComponentIdentity{
			Name:      GetComponentName(e.name, common.ComponentTypeWorker),
			Namespace: e.namespace,
		}
		manager := component.NewComponentHelper(common.ComponentTypeWorker, e.Client)
		// Only sync resources if they are explicitly set (not zero-value)
		// This prevents overwriting template defaults when user hasn't specified resources
		var workerResources corev1.ResourceRequirements
		if runtime.Spec.Worker.Resources.Requests != nil || runtime.Spec.Worker.Resources.Limits != nil {
			workerResources = runtime.Spec.Worker.Resources
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
