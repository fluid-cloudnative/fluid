/*
  Copyright 2022 The Fluid Authors.

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

package efc

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EFCEngine struct {
	runtime     *datav1alpha1.EFCRuntime
	name        string
	namespace   string
	runtimeType string
	runtimeInfo base.RuntimeInfoInterface
	UnitTest    bool
	Log         logr.Logger
	*ctrl.Helper
	client.Client
	//When reaching this gracefulShutdownLimits, the system is forced to clean up.
	gracefulShutdownLimits int32
	retryShutdown          int32
	// initImage              string
}

func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &EFCEngine{
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		Client:                 ctx.Client,
		Log:                    ctx.Log,
		runtimeType:            ctx.RuntimeType,
		gracefulShutdownLimits: 5,
		retryShutdown:          0,
	}
	err := engine.parseRuntime(ctx)
	if err != nil {
		return nil, err
	}

	// Build and setup runtime info
	runtimeInfo, err := engine.getRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get runtime info", ctx.Name)
	}

	// Build the helper
	engine.Helper = ctrl.BuildHelper(runtimeInfo, ctx.Client, engine.Log)

	template := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return template, err
}

func (e *EFCEngine) parseRuntime(ctx cruntime.ReconcileRequestContext) error {
	if ctx.Runtime != nil {
		runtime, ok := ctx.Runtime.(*datav1alpha1.EFCRuntime)
		if !ok {
			return fmt.Errorf("engine %s is failed to parse", ctx.Name)
		}
		e.runtime = runtime
	} else {
		return fmt.Errorf("engine %s is failed to parse", ctx.Name)
	}
	return nil
}

// Precheck checks if the given key can be found in the current runtime types
func Precheck(client client.Client, key types.NamespacedName) (found bool, err error) {
	var obj datav1alpha1.EFCRuntime
	return utils.CheckObject(client, key, &obj)
}
