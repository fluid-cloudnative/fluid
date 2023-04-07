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

package cachefs

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// SyncRuntime syncs the runtime spec
func (j *CacheFSEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	runtime, err := j.getRuntime()
	if err != nil {
		return changed, err
	}

	// 1. sync workers
	workerChanged, err := j.syncWorkerSpec(ctx, runtime)
	if err != nil {
		return
	}
	if workerChanged {
		j.Log.Info("Worker Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return workerChanged, err
	}

	// 2. sync fuse
	fuseChanged, err := j.syncFuseSpec(ctx, runtime)
	if err != nil {
		return
	}
	if fuseChanged {
		j.Log.Info("Fuse Spec is updated", "name", ctx.Name, "namespace", ctx.Namespace)
		return fuseChanged, err
	}
	return
}

func (j *CacheFSEngine) syncWorkerSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheFSRuntime) (changed bool, err error) {
	j.Log.V(1).Info("syncWorkerSpec")
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(j.Client,
			types.NamespacedName{Namespace: j.namespace, Name: j.getWorkerName()})
		if err != nil {
			return err
		}

		workersToUpdate := workers.DeepCopy()
		if len(workersToUpdate.Spec.Template.Spec.Containers) == 1 {
			workerResources := runtime.Spec.Worker.Resources
			if !utils.ResourceRequirementsEqual(workersToUpdate.Spec.Template.Spec.Containers[0].Resources, workerResources) {
				j.Log.Info("The resource requirement is different.", "worker sts", workersToUpdate.Spec.Template.Spec.Containers[0].Resources, "runtime", workerResources)
				workersToUpdate.Spec.Template.Spec.Containers[0].Resources = workerResources
				changed = true
			} else {
				j.Log.V(1).Info("The resource requirement of workers is the same, skip")
			}

			if changed {
				if reflect.DeepEqual(workers, workersToUpdate) {
					changed = false
					j.Log.V(1).Info("The resource requirement of worker is not changed, skip")
					return nil
				}
				j.Log.Info("The resource requirement of worker is updated")

				err = j.Client.Update(context.TODO(), workersToUpdate)
				if err != nil {
					j.Log.Error(err, "Failed to update the sts spec")
				}
			} else {
				j.Log.V(1).Info("The resource requirement of worker is not set, skip")
			}
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		j.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}

	return
}

func (j *CacheFSEngine) syncFuseSpec(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.CacheFSRuntime) (changed bool, err error) {
	j.Log.V(1).Info("syncFuseSpec")
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		fuses, err := kubeclient.GetDaemonset(j.Client, j.getFuseDaemonsetName(), j.namespace)
		if err != nil {
			return err
		}

		fusesToUpdate := fuses.DeepCopy()
		if len(fusesToUpdate.Spec.Template.Spec.Containers) == 1 {
			fuseResource := runtime.Spec.Fuse.Resources
			if !utils.ResourceRequirementsEqual(fusesToUpdate.Spec.Template.Spec.Containers[0].Resources, fuseResource) {
				j.Log.Info("The resource requirement is different.", "fuse ds", fuses.Spec.Template.Spec.Containers[0].Resources, "runtime", fuseResource)
				fusesToUpdate.Spec.Template.Spec.Containers[0].Resources = fuseResource
				changed = true
			} else {
				j.Log.V(1).Info("The resource requirement of fuse is the same, skip")
			}

			if changed {
				if reflect.DeepEqual(fuses, fusesToUpdate) {
					changed = false
					j.Log.V(1).Info("The resource requirement of fuse is not changed, skip")
					return nil
				}
				j.Log.Info("The resource requirement of fuse is updated")
				err = j.Client.Update(context.TODO(), fusesToUpdate)
				if err != nil {
					j.Log.Error(err, "Failed to update the sts spec")
				}
			} else {
				j.Log.V(1).Info("The resource requirement of fuse is not set, skip")
			}
		}

		return err
	})

	if fluiderrs.IsDeprecated(err) {
		j.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, update specs are not supported. To support these features, please create a new dataset", "details", err)
		return false, nil
	}
	return
}
