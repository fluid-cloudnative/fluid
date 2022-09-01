/*
Copyright 2022 The Fluid Author.

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

package deploy

import (
	"context"
	"errors"
	"reflect"
	"strconv"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilpointer "k8s.io/utils/pointer"
)

type CheckFunc func(client.Client, types.NamespacedName) (bool, error)

var precheckFuncs map[string]CheckFunc

func SetPrecheckFunc(checks map[string]CheckFunc) {
	precheckFuncs = checks
}

func ScaleoutRuntimeContollerOnDemand(c client.Client, datasetKey types.NamespacedName, log logr.Logger) (
	controllerName string, scaleout bool, err error) {

	for myControllerName, checkRuntime := range precheckFuncs {
		match, err := checkRuntime(c, datasetKey)
		if err != nil {
			return controllerName, scaleout, err
		}

		if match {
			scaleout, err = scaleoutDeploymentIfNeeded(c, types.NamespacedName{
				Namespace: common.NamespaceFluidSystem,
				Name:      myControllerName,
			}, log)
			if err != nil {
				return controllerName, scaleout, err
			}
			// if it's match, the skip checking other runtime controller
			return myControllerName, scaleout, nil
		}

	}

	// no matched controller
	return controllerName, scaleout, errors.New("no matched controller")
}

// scale out deployment If needed
func scaleoutDeploymentIfNeeded(c client.Client, key types.NamespacedName, log logr.Logger) (scale bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		deploy := &appsv1.Deployment{}
		err = c.Get(context.TODO(), key, deploy)
		if err != nil {
			return err
		}
		deployToUpdate := deploy.DeepCopy()
		// scale out at least 1
		if *deployToUpdate.Spec.Replicas == 0 {
			replicasStr, ok := deployToUpdate.Annotations[common.RuntimeControllerReplicas]
			var replicas int32 = 0
			if ok {
				replicasInt64, _ := strconv.ParseInt(replicasStr, 10, 32)
				replicas = int32(replicasInt64)
			}
			if replicas <= 1 {
				replicas = 1
			}
			deployToUpdate.Spec.Replicas = utilpointer.Int32(replicas)
			scale = true
		} else {
			log.V(1).Info("No need to scale out runtime controller, skip", "key", key)
			return nil
		}

		if !reflect.DeepEqual(deploy, deployToUpdate) {
			err = c.Update(context.TODO(), deployToUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error(err, "Failed to scale deployment", "key", key)
	}

	return
}
