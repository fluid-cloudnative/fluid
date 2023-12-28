/*
Copyright 2023 The Fluid Author.

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
package alluxio

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AlluxioEngine implements the Engine interface.
type AlluxioEngine struct {
	// *base.TemplateEngine
	runtime     *datav1alpha1.AlluxioRuntime
	name        string
	namespace   string
	runtimeType string
	engineImpl  string
	Log         logr.Logger
	client.Client
	retryShutdown      int32
	initImage          string
	MetadataSyncDoneCh chan base.MetadataSyncResult
	runtimeInfo        base.RuntimeInfoInterface
	UnitTest           bool
	lastCacheHitStates *cacheHitStates
	*ctrl.Helper
	Recorder record.EventRecorder
}

// Build function builds the Alluxio Engine
func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &AlluxioEngine{
		name:       ctx.Name,
		namespace:  ctx.Namespace,
		Client:     ctx.Client,
		Recorder:   ctx.Recorder,
		Log:        ctx.Log,
		runtimeType: ctx.RuntimeType,
		engineImpl: ctx.EngineImpl,
		// defaultGracefulShutdownLimits:       5,
		// defaultCleanCacheGracePeriodSeconds: 60,
		retryShutdown:      0,
		MetadataSyncDoneCh: nil,
		lastCacheHitStates: nil,
	}
	// var implement base.Implement = engine
	// engine.TemplateEngine = template
	if ctx.Runtime != nil {
		runtime, ok := ctx.Runtime.(*datav1alpha1.AlluxioRuntime)
		if !ok {
			return nil, fmt.Errorf("engine %s is failed to parse", ctx.Name)
		}
		engine.runtime = runtime
	} else {
		return nil, fmt.Errorf("engine %s is failed to parse", ctx.Name)
	}

	// Build and setup runtime info
	runtimeInfo, err := engine.getRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get runtime info, error %s", ctx.Name, err.Error())
	}

	// Build the helper
	engine.Helper = ctrl.BuildHelper(runtimeInfo, ctx.Client, engine.Log)

	template := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return template, err
}

// Precheck checks if the given key can be found in the current runtime types
func Precheck(client client.Client, key types.NamespacedName) (found bool, err error) {
	var obj datav1alpha1.AlluxioRuntime
	return utils.CheckObject(client, key, &obj)
}
