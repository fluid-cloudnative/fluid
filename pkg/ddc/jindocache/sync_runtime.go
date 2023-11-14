/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// SyncRuntime syncs the runtime spec
func (e *JindoCacheEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
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

func (e *JindoCacheEngine) syncMasterSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JindoRuntime) (changed bool, err error) {
	e.Log.V(1).Info("syncMasterSpec")
	if runtime.Spec.Master.Disabled {
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		master, err := kubeclient.GetStatefulSet(e.Client, e.getMasterName(), e.namespace)
		if err != nil {
			return err
		}

		// if len(runtime.Spec.Master.Resources.Limits) == 0 && len(runtime.Spec.Master.Resources.Requests) == 0 {
		// 	e.Log.V(1).Info("The resource requirement is not set, skip")
		// 	return nil
		// }

		masterToUpdate := master.DeepCopy()
		if len(masterToUpdate.Spec.Template.Spec.Containers) == 1 {
			// masterResources := utils.TranformResourcesWithTieredStore(runtime.Spec.Master.Resources,
			// 	masterToUpdate.Spec.Template.Spec.Containers[0].Resources)
			masterResources := runtime.Spec.Master.Resources
			if !utils.ResourceRequirementsEqual(masterToUpdate.Spec.Template.Spec.Containers[0].Resources, masterResources) {
				e.Log.Info("The resource requirement is different.", "master sts", masterToUpdate.Spec.Template.Spec.Containers[0].Resources,
					"runtime", masterResources)
				masterToUpdate.Spec.Template.Spec.Containers[0].Resources =
					masterResources
				changed = true
			} else {
				e.Log.V(1).Info("The resource requirement of master is the same, skip")
			}
			if changed {
				if reflect.DeepEqual(master, masterToUpdate) {
					changed = false
					e.Log.V(1).Info("The resource requirement of master is not changed, skip")
					return nil
				}
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

func (e *JindoCacheEngine) syncWorkerSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JindoRuntime) (changed bool, err error) {
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

		// if len(runtime.Spec.Worker.Resources.Limits) == 0 &&
		// 	len(runtime.Spec.Worker.Resources.Requests) == 0 {
		// 	e.Log.V(1).Info("The resource requirement of worker is not set, skip")
		// 	return nil
		// }

		workersToUpdate := workers.DeepCopy()
		if len(workersToUpdate.Spec.Template.Spec.Containers) == 1 {
			// workerResources := utils.TranformResourcesWithTieredStore(runtime.Spec.Worker.Resources,
			// 	workersToUpdate.Spec.Template.Spec.Containers[0].Resources)
			workerResources := runtime.Spec.Worker.Resources
			if !utils.ResourceRequirementsEqual(workersToUpdate.Spec.Template.Spec.Containers[0].Resources, workerResources) {
				e.Log.Info("The resource requirement is different.", "worker sts", workersToUpdate.Spec.Template.Spec.Containers[0].Resources,
					"runtime", workerResources)
				workersToUpdate.Spec.Template.Spec.Containers[0].Resources =
					workerResources
				changed = true
			} else {
				e.Log.V(1).Info("The resource requirement of workers is the same, skip")
			}

			if changed {
				if reflect.DeepEqual(workers, workersToUpdate) {
					changed = false
					e.Log.V(1).Info("The resource requirement of worker is not changed, skip")
					return nil
				}
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

func (e *JindoCacheEngine) syncFuseSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.JindoRuntime) (changed bool, err error) {
	e.Log.V(1).Info("syncFuseSpec")
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		fuses, err := kubeclient.GetDaemonset(e.Client, e.getFuseName(), e.namespace)
		if err != nil {
			return err
		}

		// if len(runtime.Spec.Fuse.Resources.Limits) == 0 && len(runtime.Spec.Fuse.Resources.Requests) == 0 {
		// 	e.Log.V(1).Info("The resource requirement of Fuse is not set, skip")
		// 	return nil
		// }

		fusesToUpdate := fuses.DeepCopy()
		if len(fusesToUpdate.Spec.Template.Spec.Containers) == 1 {
			fuseResource :=
				runtime.Spec.Fuse.Resources
			if !utils.ResourceRequirementsEqual(fusesToUpdate.Spec.Template.Spec.Containers[0].Resources, fuseResource) {
				e.Log.Info("The resource requirement is different.", "fuse ds", fuses.Spec.Template.Spec.Containers[0].Resources,
					"runtime", fuseResource)
				fusesToUpdate.Spec.Template.Spec.Containers[0].Resources = fuseResource
				changed = true
			} else {
				e.Log.V(1).Info("The resource requirement of fuse is the same, skip")
			}

			if changed {
				if reflect.DeepEqual(fuses, fusesToUpdate) {
					changed = false
					e.Log.V(1).Info("The resource requirement of fuse is not changed, skip")
					return nil
				}
				e.Log.Info("The resource requirement of fuse is updated")
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
