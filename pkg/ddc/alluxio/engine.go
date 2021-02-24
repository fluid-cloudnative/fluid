/*

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
	"os"
	"regexp"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// AlluxioEngine implements the Engine interface.
type AlluxioEngine struct {
	// *base.TemplateEngine
	runtime     *datav1alpha1.AlluxioRuntime
	name        string
	namespace   string
	runtimeType string
	Log         logr.Logger
	client.Client
	// gracefulShutdownLimits is the limit for the system to forcibly clean up.
	gracefulShutdownLimits int32
	retryShutdown          int32
	initImage              string
	MetadataSyncDoneCh     chan MetadataSyncResult
	runtimeInfo            base.RuntimeInfoInterface
	UnitTest               bool
	lastCacheHitStates     *cacheHitStates
}

// Build function builds the Alluxio Engine
func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &AlluxioEngine{
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		Client:                 ctx.Client,
		Log:                    ctx.Log,
		runtimeType:            ctx.RuntimeType,
		gracefulShutdownLimits: 5,
		retryShutdown:          0,
		MetadataSyncDoneCh:     nil,
		lastCacheHitStates:     nil,
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

	// Setup runtime Info
	runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, engine.runtimeType, engine.runtime.Spec.Tieredstore)
	if err != nil {
		return nil, err
	}
	if engine.runtime.Spec.Fuse.Global {
		runtimeInfo.SetupFuseDeployMode(engine.runtime.Spec.Fuse.Global, engine.runtime.Spec.Fuse.NodeSelector)
		ctx.Log.Info("Enable global mode for fuse")
	} else {
		ctx.Log.Info("Disable global mode for fuse")
	}
	engine.runtimeInfo = runtimeInfo

	// Setup init image for Alluxio Engine
	if value, existed := os.LookupEnv(common.ALLUXIO_INIT_IMAGE_ENV); existed {
		if matched, err := regexp.MatchString("^\\S+:\\S+$", value); err == nil && matched {
			engine.initImage = value
		} else {
			ctx.Log.Info("Failed to parse the ALLUXIO_INIT_IMAGE_ENV", "ALLUXIO_INIT_IMAGE_ENV", value, "error", err)
		}
		ctx.Log.Info("Get INIT_IMAGE from Env", common.ALLUXIO_INIT_IMAGE_ENV, value)
	} else {
		ctx.Log.Info("Use Default ALLUXIO_INIT_IMAGE_ENV", "ALLUXIO_INIT_IMAGE_ENV", common.DEFAULT_ALLUXIO_INIT_IMAGE)
	}
	if len(engine.initImage) == 0 {
		engine.initImage = common.DEFAULT_ALLUXIO_INIT_IMAGE
	}

	ctx.Log.Info("check alluxio engine initImage", "engine.initImage", engine.initImage)

	template := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return template, err
}
