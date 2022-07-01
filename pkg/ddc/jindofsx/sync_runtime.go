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

package jindofsx

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// SyncRuntime syncs the runtime spec
func (e *JindoFSxEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return changed, err
	}

	// 1. sync master
	masterChanged, err := e.syncMasterSpec(ctx, runtime)
	if err != nil {
		return
	}
	if masterChanged {
		e.Log.Info("Master Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return masterChanged, err
	}

	// 2. sync workers
	workerChanged, err := e.syncWorkerSpec(ctx, runtime)
	if err != nil {
		return
	}
	if workerChanged {
		e.Log.Info("Worker Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return workerChanged, err
	}

	// 3. sync fuse
	fuseChanged, err := e.syncFuseSpec(ctx, runtime)
	if err != nil {
		return
	}
	if fuseChanged {
		e.Log.Info("Fuse Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return fuseChanged, err
	}

	return
}

func (e *JindoFSxEngine) syncMasterSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JindoRuntime) (changed bool, err error) {
	e.Log.V(1).Info("syncMasterSpec")
	if runtime.Spec.Master.Disabled {
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		master, err := kubeclient.GetStatefulSet(e.Client, e.getMasterName(), e.namespace)
		if err != nil {
			return err
		}

		if len(runtime.Spec.Master.Resources.Limits) == 0 && len(runtime.Spec.Master.Resources.Requests) == 0 {
			e.Log.V(1).Info("The resource requirement is not set, skip")
			return nil
		}

		masterToUpdate := master.DeepCopy()
		if len(masterToUpdate.Spec.Template.Spec.Containers) == 1 {
			if !reflect.DeepEqual(masterToUpdate.Spec.Template.Spec.Containers[0].Resources, runtime.Spec.Master.Resources) {
				e.Log.Info("The resource requirement is different.", "worker sts", masterToUpdate.Spec.Template.Spec.Containers[0].Resources,
					"runtime", runtime.Spec.Master.Resources)
				masterToUpdate.Spec.Template.Spec.Containers[0].Resources =
					tranformResources(runtime.Spec.Master.Resources,
						masterToUpdate.Spec.Template.Spec.Containers[0].Resources)
				changed = true
			} else {
				e.Log.V(1).Info("The resource requirement of master is the same.")
			}
			if changed {
				e.Log.Info("The resource requirement of master is updated")
				err = e.Client.Update(context.TODO(), masterToUpdate)
				if err != nil {
					e.Log.Error(err, "Failed to update the sts spec")
				}
			}
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}

	return
}

func (e *JindoFSxEngine) syncWorkerSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JindoRuntime) (changed bool, err error) {
	e.Log.V(1).Info("syncWorkerSpec")
	if runtime.Spec.Worker.Disabled {
		return
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
		if err != nil {
			return err
		}

		if len(runtime.Spec.Worker.Resources.Limits) == 0 &&
			len(runtime.Spec.Worker.Resources.Requests) == 0 {
			e.Log.V(1).Info("The resource requirement of worker is not set, skip")
			return nil
		}

		workersToUpdate := workers.DeepCopy()
		if len(workersToUpdate.Spec.Template.Spec.Containers) == 1 {
			if !reflect.DeepEqual(workersToUpdate.Spec.Template.Spec.Containers[0].Resources, runtime.Spec.Worker.Resources) {
				e.Log.Info("The resource requirement is different.", "worker sts", workersToUpdate.Spec.Template.Spec.Containers[0].Resources,
					"runtime", runtime.Spec.Worker.Resources)
				workersToUpdate.Spec.Template.Spec.Containers[0].Resources =
					tranformResources(runtime.Spec.Worker.Resources,
						workersToUpdate.Spec.Template.Spec.Containers[0].Resources)
				changed = true
			} else {
				e.Log.Info("The resource requirement is the same.")
			}

			if changed {
				e.Log.Info("The resource requirement of worker is updated")
				err = e.Client.Update(context.TODO(), workersToUpdate)
				if err != nil {
					e.Log.Error(err, "Failed to update the sts spec")
				}
			} else {
				e.Log.V(1).Info("The resource requirement of worker is not set, skip")
			}
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}

	return
}

func (e *JindoFSxEngine) syncFuseSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JindoRuntime) (changed bool, err error) {
	e.Log.V(1).Info("syncFuseSpec")
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		fuses, err := kubeclient.GetDaemonset(e.Client, e.getFuseName(), e.namespace)
		if err != nil {
			return err
		}

		if len(runtime.Spec.Fuse.Resources.Limits) == 0 && len(runtime.Spec.Fuse.Resources.Requests) == 0 {
			e.Log.V(1).Info("The resource requirement of Fuse is not set, skip")
			return nil
		}

		fusesToUpdate := fuses.DeepCopy()
		if len(fusesToUpdate.Spec.Template.Spec.Containers) == 1 {
			if !reflect.DeepEqual(fusesToUpdate.Spec.Template.Spec.Containers[0].Resources, runtime.Spec.Fuse.Resources) {
				e.Log.Info("The resource requirement is different.", "fuse ds", fuses.Spec.Template.Spec.Containers[0].Resources,
					"runtime", runtime.Spec.Fuse.Resources)
				fusesToUpdate.Spec.Template.Spec.Containers[0].Resources =
					tranformResources(runtime.Spec.Fuse.Resources,
						fusesToUpdate.Spec.Template.Spec.Containers[0].Resources)
				changed = true
			}

			if changed {
				err = e.Client.Update(context.TODO(), fusesToUpdate)
				if err != nil {
					e.Log.Error(err, "Failed to update the sts spec")
				}
			} else {
				e.Log.V(1).Info("The resource requirement of fuse is not set, skip")
			}
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}
	return
}

func tranformResources(runtimeResources corev1.ResourceRequirements,
	current corev1.ResourceRequirements) corev1.ResourceRequirements {

	if len(runtimeResources.Requests) == 0 || runtimeResources.Requests.Memory() == nil {
		if current.Requests.Memory() != nil {
			if len(runtimeResources.Requests) == 0 {
				runtimeResources.Requests = make(corev1.ResourceList)
			}
			runtimeResources.Requests[corev1.ResourceMemory] =
				*current.Requests.Memory()
		}
	}

	if len(runtimeResources.Limits) == 0 || runtimeResources.Limits.Memory() == nil {
		if current.Limits.Memory() != nil {
			if len(runtimeResources.Limits) == 0 {
				runtimeResources.Limits = make(corev1.ResourceList)
			}
			runtimeResources.Limits[corev1.ResourceMemory] =
				*current.Limits.Memory()
		}
	}
	return runtimeResources
}
