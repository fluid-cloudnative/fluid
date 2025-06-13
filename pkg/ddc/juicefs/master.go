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
	"context"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func (j JuiceFSEngine) CheckMasterReady() (ready bool, err error) {
	// JuiceFS Runtime has no master role
	return true, nil
}

// ShouldSetupMaster checks if a further call of func `SetupMaster` is needed.
// JuiceFS Runtime has no master role, so the function check runtime.status.WorkerPhase
// to know if juicefs is installed and set up.
func (j JuiceFSEngine) ShouldSetupMaster() (should bool, err error) {
	runtime, err := j.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.WorkerPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}
	return
}

// SetupMaster installs juicefs components into the cluster.
// JuiceFS Runtime has no master role, implementing func `SetupMaster` here
// is just for a same lifecycle as other runtimes (other runtimes may have master component)
func (j JuiceFSEngine) SetupMaster() (err error) {
	workerName := j.getWorkerName()

	// 1. Setup
	_, err = kubeclient.GetStatefulSet(j.Client, workerName, j.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		//1. Is not found error
		j.Log.V(1).Info("SetupMaster", "worker", workerName)
		return j.installJuiceFS()
	} else if err != nil {
		//2. Other errors
		return
	} else {
		//3.The fuse has been set up
		j.Log.V(1).Info("The worker has been set.")
	}

	// 2. Update the status of the runtime
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()

		// Init selector for worker
		runtimeToUpdate.Status.Selector = j.getWorkerSelectors()
		runtimeToUpdate.Status.ValueFileConfigmap = j.getHelmValuesConfigMapName()

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return j.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	if err != nil {
		j.Log.Error(err, "Update runtime status")
		return err
	}

	return
}
