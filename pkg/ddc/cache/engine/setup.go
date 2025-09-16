package engine

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/componenthelper"
	"github.com/fluid-cloudnative/fluid/pkg/metrics"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
)

func (e *CacheEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	defer func() {
		if err != nil {
			metrics.GetOrCreateRuntimeMetrics(ctx.Runtime.GetObjectKind().GroupVersionKind().Kind, ctx.Namespace, ctx.Name).SetupErrorInc()
		}
	}()

	runtime := ctx.Runtime.(*datav1alpha1.CacheRuntime)

	runtimeClass, err := utils.GetCacheRuntimeClass(ctx.Client, runtime.Spec.RuntimeClassName)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get CacheRuntimeClass %s", runtime.Spec.RuntimeClassName)
	}

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return false, err
	}

	configmap := corev1.ConfigMap{}
	if err := e.Get(context.TODO(), types.NamespacedName{
		Name:      e.getRuntimeConfigCmName(),
		Namespace: e.namespace,
	}, &configmap); err != nil {
		if apierrors.IsNotFound(err) {
			b, _ := json.Marshal("")
			configMapToCreate := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      e.getRuntimeConfigCmName(),
					Namespace: e.namespace,
				},
				Data: map[string]string{
					"config.json": string(b),
				},
			}
			if err := e.Client.Create(context.TODO(), configMapToCreate); err != nil {
				return false, err
			}
			e.Log.Info("Initialized runtime config", "configmapName", e.getRuntimeConfigCmName())
		}
	}

	runtimeValue, err := e.transform(dataset, runtime, runtimeClass)
	if err != nil {
		return false, err
	}
	e.value = runtimeValue

	e.Log.Info("Setup runtime", "runtime", ctx.Runtime)
	if runtimeValue.Master.Enabled {
		if e.masterHelper == nil {
			e.masterHelper = componenthelper.NewComponentHelper(runtimeValue.Master.WorkloadType, e.Scheme, e.Client)
		}
		e.Log.Info("Setup master", "runtime", ctx.Runtime)
		ready, err = e.SetupMasterComponent(runtimeValue.Master)
		if !ready || err != nil {
			return
		}
	}

	if runtimeValue.Worker.Enabled {
		if e.workerHelper == nil {
			e.workerHelper = componenthelper.NewComponentHelper(runtimeValue.Worker.WorkloadType, e.Scheme, e.Client)
		}
		e.Log.Info("Setup worker", "runtime", ctx.Runtime)
		ready, err = e.SetupWorkerComponent(runtimeValue.Worker)
		if !ready || err != nil {
			return
		}
	}

	if runtimeValue.Client.Enabled {
		if e.clientHelper == nil {
			e.clientHelper = componenthelper.NewComponentHelper(runtimeValue.Client.WorkloadType, e.Scheme, e.Client)
		}
		e.Log.Info("Setup client", "runtime", ctx.Runtime)
		ready, err = e.SetupClientComponent(runtimeValue.Client)
		if !ready || err != nil {
			return
		}
	}

	ready, err = e.CheckAndUpdateRuntimeStatus(runtimeValue)
	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to check if the runtime is ready", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return
	}
	if !ready {
		return
	}

	if err = e.BindToDataset(); err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to bind the dataset", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return false, err
	}

	return true, nil
}
