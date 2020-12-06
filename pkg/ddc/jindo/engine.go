package jindo

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

type JindoEngine struct {
	runtime     *datav1alpha1.JindoRuntime
	name        string
	namespace   string
	runtimeType string
	Log         logr.Logger
	client.Client
	//When reaching this gracefulShutdownLimits, the system is forced to clean up.
	gracefulShutdownLimits int32
	retryShutdown          int32
	AccessModes            []v1.PersistentVolumeAccessMode
	//initImage              string
}

func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &JindoEngine{
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		Client:                 ctx.Client,
		Log:                    ctx.Log,
		runtimeType:            ctx.RuntimeType,
		gracefulShutdownLimits: 5,
		retryShutdown:          0,
	}
	// var implement base.Implement = engine
	// engine.TemplateEngine = template
	if ctx.Runtime != nil {
		runtime, ok := ctx.Runtime.(*datav1alpha1.JindoRuntime)
		if !ok {
			return nil, fmt.Errorf("engine %s is failed to parse", ctx.Name)
		}
		engine.runtime = runtime
	} else {
		return nil, fmt.Errorf("engine %s is failed to parse", ctx.Name)
	}
	// Get access mode
	dataset, err := utils.GetDataset(ctx.Client, ctx.Name, ctx.Namespace)
	if err != nil {
		return nil, err
	}
	engine.AccessModes = dataset.Spec.AccessModes

	template := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return template, err
}
