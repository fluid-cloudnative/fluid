package utils

import (
	"context"
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilpointer "k8s.io/utils/pointer"
)

type checkFunc func(client.Client, types.NamespacedName) (bool, error)

var checkFuncs map[string]checkFunc = map[string]checkFunc{
	"alluxioruntime-controller": CheckAlluxioRuntime,
	"jindoruntime-controller":   CheckJindoRuntime,
	"juicefsruntime-controller": CheckJuiceFSRuntime,
	"goosefsruntime-controller": CheckGooseFSRuntime,
}

func InvokeRuntimeContollerOnDemand(c client.Client, dataset *datav1alpha1.Dataset) (
	controllerName string, err error) {

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
			return controllerName, err
		}

		if match {
			err = createRuntimeControllerIfNeeded(c, types.NamespacedName{
				Namespace: "fluid-system",
				Name:      myControllerName,
			})
			if err != nil {
				return controllerName, err
			}
			// if it's match, the skip checking other runtime controller
			return myControllerName, nil
		}

	}

	return
}

func createRuntimeControllerIfNeeded(c client.Client, key types.NamespacedName) (err error) {
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
