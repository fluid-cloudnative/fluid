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

package goosefs

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// GooseFSEngine implements the Engine interface.
type GooseFSEngine struct {
	// *base.TemplateEngine
	runtime     *datav1alpha1.GooseFSRuntime
	name        string
	namespace   string
	runtimeType string
	Log         logr.Logger
	client.Client
	// gracefulShutdownLimits is the limit for the system to forcibly clean up.
	gracefulShutdownLimits int32
	retryShutdown          int32
	initImage              string
	MetadataSyncDoneCh     chan base.MetadataSyncResult
	runtimeInfo            base.RuntimeInfoInterface
	UnitTest               bool
	lastCacheHitStates     *cacheHitStates
	*ctrl.Helper
	Recorder record.EventRecorder
}

// Build function builds the GooseFS Engine
func Build(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	engine := &GooseFSEngine{
		name:                   ctx.Name,
		namespace:              ctx.Namespace,
		Client:                 ctx.Client,
		Recorder:               ctx.Recorder,
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
		runtime, ok := ctx.Runtime.(*datav1alpha1.GooseFSRuntime)
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
		return nil, fmt.Errorf("engine %s failed to get runtime info", ctx.Name)
	}

	// Build the helper
	engine.Helper = ctrl.BuildHelper(runtimeInfo, ctx.Client, engine.Log)

	template := base.NewTemplateEngine(engine, id, ctx)

	err = kubeclient.EnsureNamespace(ctx.Client, ctx.Namespace)
	return template, err
}

// Precheck checks if the given key can be found in the current runtime types
func Precheck(client client.Client, key types.NamespacedName) (found bool, err error) {
	var obj datav1alpha1.GooseFSRuntime
	return utils.CheckObject(client, key, &obj)
}
