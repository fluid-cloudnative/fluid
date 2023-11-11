/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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

	isRef, err := CheckReferenceDatasetRuntime(ctx, runtime)
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
func CheckReferenceDatasetRuntime(ctx cruntime.ReconcileRequestContext, runtime *datav1alpha1.ThinRuntime) (bool, error) {
	// Reference runtime must have empty spec.profileName
	return len(runtime.Spec.ThinRuntimeProfileName) == 0, nil
}
