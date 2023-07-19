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
		"alluxio":    alluxio.Build,
		"jindo":      jindo.Build,
		"jindofsx":   jindofsx.Build,
		"goosefs":    goosefs.Build,
		"juicefs":    juicefs.Build,
		"thin":       thin.Build,
		"efc":        efc.Build,
		"jindocache": jindocache.Build,
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

/**
* Build Engine from config
 */
func CreateEngine(id string, ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error) {

	if buildeFunc, found := buildFuncMap[ctx.RuntimeType]; found {
		engine, err = buildeFunc(id, ctx)
	} else {
		err = fmt.Errorf("failed to build the engine due to the type %s is not found", ctx.RuntimeType)
	}

	return
}

/**
* GenerateEngineID generates Engine ID
 */
func GenerateEngineID(namespacedName types.NamespacedName) string {
	return fmt.Sprintf("%s-%s",
		namespacedName.Namespace, namespacedName.Name)
}
