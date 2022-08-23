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

package runtime

import (
	"context"
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilpointer "k8s.io/utils/pointer"
)

type checkFunc func(client.Client, types.NamespacedName) (bool, error)

var checkFuncs map[string]checkFunc = map[string]checkFunc{
	"alluxioruntime-controller": utils.CheckAlluxioRuntime,
	"jindoruntime-controller":   utils.CheckJindoRuntime,
	"juicefsruntime-controller": utils.CheckJuiceFSRuntime,
	"goosefsruntime-controller": utils.CheckGooseFSRuntime,
}

func CreateRuntimeContollerOnDemand(c client.Client, dataset *datav1alpha1.Dataset, log logr.Logger) (
	controllerName string, scaleout bool, err error) {

	if dataset != nil {
		err = fmt.Errorf("the dataset is nil")
		return
	}
	key := types.NamespacedName{
		Namespace: dataset.Namespace,
		Name:      dataset.Name,
	}

	for myControllerName, checkRuntime := range checkFuncs {
		match, err := checkRuntime(c, key)
		if err != nil {
			return controllerName, scaleout, err
		}

		if match {
			scaleout, err = createRuntimeControllerIfNeeded(c, types.NamespacedName{
				Namespace: "fluid-system",
				Name:      myControllerName,
			}, log)
			if err != nil {
				return controllerName, scaleout, err
			}
			// if it's match, the skip checking other runtime controller
			return myControllerName, scaleout, nil
		}

	}

	return
}

func createRuntimeControllerIfNeeded(c client.Client, key types.NamespacedName, log logr.Logger) (scale bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		deploy := &appsv1.Deployment{}
		err = c.Get(context.TODO(), key, deploy)
		if err != nil {
			return err
		}
		deployToUpdate := deploy.DeepCopy()
		// scale out
		if *deployToUpdate.Spec.Replicas == 0 {
			deployToUpdate.Spec.Replicas = utilpointer.Int32(1)
			scale = true
		} else {
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
