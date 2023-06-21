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

package thin

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/thin/referencedataset"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

type ThinEngine struct {
	runtime        *datav1alpha1.ThinRuntime
	runtimeProfile *datav1alpha1.ThinRuntimeProfile
	name           string
	namespace      string
	runtimeType    string
	Log            logr.Logger
	client.Client
	//When reaching this gracefulShutdownLimits, the system is forced to clean up.
	gracefulShutdownLimits int32
	MetadataSyncDoneCh     chan base.MetadataSyncResult
	runtimeInfo            base.RuntimeInfoInterface
	UnitTest               bool
	retryShutdown          int32
	*ctrl.Helper
}

func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	if ctx.Runtime == nil {
		return nil, fmt.Errorf("engine %s is failed due to runtime is nil", ctx.Name)
	}
	runtime, ok := ctx.Runtime.(*datav1alpha1.ThinRuntime)
	if !ok {
		return nil, fmt.Errorf("engine %s is failed due to type conversion", ctx.Name)
	}

	isRef, err := CheckReferenceDatasetRuntime(ctx.Client, runtime)
	if err != nil {
		return nil, err
	}

	if isRef {
		return referencedataset.BuildReferenceDatasetThinEngine(id, ctx)
	} else {
		return buildThinEngine(id, ctx)
	}
}

// buildThinEngine build engine for handling file system dataset
func buildThinEngine(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &ThinEngine{
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		Client:                 ctx.Client,
		Log:                    ctx.Log,
		runtimeType:            ctx.RuntimeType,
		gracefulShutdownLimits: 5,
		retryShutdown:          0,
		MetadataSyncDoneCh:     nil,
	}

	runtime := ctx.Runtime.(*datav1alpha1.ThinRuntime)
	engine.runtime = runtime

	runtimeProfile, err := utils.GetThinRuntimeProfile(ctx.Client, runtime.Spec.ThinRuntimeProfileName)
	if err != nil {
		return nil, errors.Wrapf(err, "error when getting thinruntime profile %s", runtime.Spec.ThinRuntimeProfileName)
	}
	engine.runtimeProfile = runtimeProfile

	// Build and setup runtime info
	runtimeInfo, err := engine.getRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get runtime info", ctx.Name)
	}

	engine.Helper = ctrl.BuildHelper(runtimeInfo, ctx.Client, engine.Log)
	templateEngine := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return templateEngine, err
}

func Precheck(client client.Client, key types.NamespacedName) (found bool, err error) {
	var obj datav1alpha1.ThinRuntime
	return utils.CheckObject(client, key, &obj)
}

// CheckReferenceDatasetRuntime judge if this runtime is used for handling dataset mounting another dataset.
func CheckReferenceDatasetRuntime(client client.Client, runtime *datav1alpha1.ThinRuntime) (bool, error) {
	dataset, err := utils.GetDataset(client, runtime.Name, runtime.Namespace)
	if err != nil && utils.IgnoreNotFound(err) != nil {
		// if err is not found, try to GetMountedDatasetNamespacedName from runtime.status.mounts, don't return here.
		return false, err
	}

	var mounted []types.NamespacedName
	if dataset != nil {
		// getMountedDataset from dataset first
		mounted = base.GetMountedDatasetNamespacedName(dataset.Spec.Mounts)
	} else if runtime.Status.Mounts != nil && len(runtime.Status.Mounts) != 0 {
		// then try to getMountedDataset from runtime
		mounted = base.GetMountedDatasetNamespacedName(runtime.Status.Mounts)
	}
	// not mount other datasets
	if len(mounted) == 0 {
		return false, nil
	}

	return true, nil
}
