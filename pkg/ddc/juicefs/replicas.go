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

package juicefs

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// SyncReplicas syncs the replicas
func (j *JuiceFSEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {
	var (
		workerName string = j.getWorkerName()
		namespace  string = j.namespace
	)

	workers, err := kubeclient.GetStatefulSet(j.Client, workerName, namespace)
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		err = j.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
		return err
	})
	if err != nil {
		return utils.LoggingErrorExceptConflict(j.Log, err, "Failed to sync the replicas",
			types.NamespacedName{Namespace: j.namespace, Name: j.name})
	}

	return
}
