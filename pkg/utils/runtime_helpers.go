package utils

import (
	"context"
	"fmt"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/jindo"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// CheckAlluxioRuntime checks Alluxio Runtime object with the given name and namespace
func CheckAlluxioRuntime(client client.Client, name, namespace string) (*data.AlluxioRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.AlluxioRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// CheckJindoRuntime checks Jindo Runtime object with the given name and namespace
func CheckJindoRuntime(client client.Client, name, namespace string) (*data.JindoRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.JindoRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// CheckGooseFSRuntime checks GooseFS Runtime object with the given name and namespace
func CheckGooseFSRuntime(client client.Client, name, namespace string) (*data.GooseFSRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.GooseFSRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// CheckJuiceFSRuntime checks JuiceFS Runtime object with the given name and namespace
func CheckJuiceFSRuntime(client client.Client, name, namespace string) (*data.JuiceFSRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.JuiceFSRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func InvokeRuntimeContollerOnDemand(ctx cruntime.ReconcileRequestContext, runtimeClient client.Client, dataset *datav1alpha1.Dataset) (controllerName string,
	err error) {

	if dataset != nil {
		err = fmt.Errorf("the dataset is nil")
		return
	}

	key := types.NamespacedName{
		Namespace: dataset.Namespace,
		Name:      dataset.Name,
	}

	checkRuntime(runtimeClient, key, GetJindoRuntime)

	var fluidRuntime client.Object
	fluidRuntime, err = GetAlluxioRuntime(runtimeClient, dataset.Name, dataset.Namespace)
	if err != nil {
	}

	switch ctx.RuntimeType {
	case common.AlluxioRuntime:

	case common.JindoRuntime:
		fluidRuntime, err = GetJindoRuntime(runtimeClient, dataset.Name, dataset.Namespace)
		ctx.RuntimeType = jindo.GetRuntimeType()
	case common.GooseFSRuntime:
		fluidRuntime, err = GetGooseFSRuntime(runtimeClient, dataset.Name, dataset.Namespace)
	case common.JuiceFSRuntime:
		fluidRuntime, err = GetJuiceFSRuntime(runtimeClient, dataset.Name, dataset.Namespace)
	default:
		ctx.Log.Error(fmt.Errorf("RuntimeNotSupported"), "The runtime is not supported yet", "dataset", dataset)
	}

	if err != nil {
		if IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("The runtime is not found", "runtime", ctx.NamespacedName)
			return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get the ddc runtime")
			return RequeueIfError(errors.Wrap(err, "Unable to get ddc runtime"))
		}
	}

}

func checkRuntime(runtimeClient client.Client,
	key types.NamespacedName,
	getRuntime func(c client.Client, name string, namespace string) (o client.Object, err error)) (match bool, err error) {
	var runtime client.Object
	runtime, err = getRuntime(runtimeClient, key.Name, key.Namespace)
	if err != nil {
		if IgnoreNotFound(err) == nil {
			err = nil
		}
		return
	}
	log.Info("Succeed in finding the object", "runtime", runtime)
	return
}
