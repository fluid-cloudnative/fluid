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

package ddc

import (
	"strings"

	fluidv1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/deploy"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindocache"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/thin"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"fmt"
)

type buildFunc func(id string, ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error)

var buildFuncMap map[string]buildFunc

func init() {
	buildFuncMap = map[string]buildFunc{
		common.AlluxioEngineImpl:    alluxio.Build,
		common.JindoFSEngineImpl:    jindo.Build,
		common.JindoFSxEngineImpl:   jindofsx.Build,
		common.JindoCacheEngineImpl: jindocache.Build,
		common.GooseFSEngineImpl:    goosefs.Build,
		common.JuiceFSEngineImpl:    juicefs.Build,
		common.ThinEngineImpl:       thin.Build,
		common.EFCEngineImpl:        efc.Build,
	}

	deploy.SetPrecheckFunc(map[string]deploy.CheckFunc{
		"alluxioruntime-controller": alluxio.Precheck,
		"jindoruntime-controller":   jindofsx.Precheck,
		"juicefsruntime-controller": juicefs.Precheck,
		"goosefsruntime-controller": goosefs.Precheck,
		"thinruntime-controller":    thin.Precheck,
		"efcruntime-controller":     efc.Precheck,
	})
}

// CreateEngine chooses one engine implementation according to `ctx.EngineImpl` and builds a concrete engine.
func CreateEngine(id string, ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error) {

	if buildFunc, found := buildFuncMap[ctx.EngineImpl]; found {
		engine, err = buildFunc(id, ctx)
	} else {
		err = fmt.Errorf("failed to build the engine due to the type %s is not found", ctx.RuntimeType)
	}

	return
}

// GenerateEngineID generates an Engine ID
func GenerateEngineID(namespacedName types.NamespacedName) string {
	return fmt.Sprintf("%s-%s",
		namespacedName.Namespace, namespacedName.Name)
}

// InferEngineImpl infers which engineImpl should be use for a given runtime.
// For a new runtime which has not been set up, returns the default engineImpl.
// NOTE: for backward compatibility, the func checks runtimeStatus.ValueFileConfigmap to identify
// the engine implementation for a running Runtime.
// TODO: This could change in the future by checking runtimeStatus.engineImpl instead.
func InferEngineImpl(runtimeStatus fluidv1alpha1.RuntimeStatus, defaultImpl string) string {
	if len(runtimeStatus.ValueFileConfigmap) == 0 {
		return defaultImpl
	}

	// e.g. <dataset_name>-<engine_impl>-values, <dataset_name> may contain "-"
	parts := strings.Split(runtimeStatus.ValueFileConfigmap, "-")
	if len(parts) >= 3 {
		engineImpl := parts[len(parts)-2]
		if _, exists := buildFuncMap[engineImpl]; exists {
			return engineImpl
		}
	}

	return defaultImpl
}
