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

package jindo

import (
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// SyncRuntime syncs the runtime spec
func (e *JindoEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	if e.Client == nil {
		e.Log.V(1).Info("Client is nil, skipping SyncRuntime")
		return false, nil
	}

	// 1. Get the workers StatefulSet
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.V(1).Info("Workers not found", "name", e.getWorkerName())
			return false, nil
		}
		return false, err
	}

	// 2. Extract current image
	var currentImage string
	var containerIndex int = -1
	for i, container := range workers.Spec.Template.Spec.Containers {
		if container.Name == WorkerContainerName {
			currentImage = container.Image
			containerIndex = i
			break
		}
	}

	if len(currentImage) == 0 || containerIndex == -1 {
		e.Log.V(1).Info("Worker container not found", "name", e.getWorkerName())
		return false, nil
	}

	// 3. Compute desired image using shared helper
	desiredImage := resolveSmartDataImage()

	// 4. Compare desired image vs current image
	if currentImage != desiredImage {
		e.Log.Info("Image drift detected, triggering rolling upgrade", "current", currentImage, "desired", desiredImage)

		// Update the image
		workersToUpdate := workers.DeepCopy()
		workersToUpdate.Spec.Template.Spec.Containers[containerIndex].Image = desiredImage

		err = e.Client.Update(ctx.Context, workersToUpdate)
		if err != nil {
			e.Log.Error(err, "Failed to update worker image", "desired", desiredImage)
			return false, err
		}

		e.Log.Info("Successfully triggered rolling upgrade for worker StatefulSet", "name", e.getWorkerName())

		// Return false, nil to continue reconciliation without short-circuiting
		return false, nil
	}

	return false, nil
}
