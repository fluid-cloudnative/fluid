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
		"jindocache": jindocache.Build,
		"goosefs":    goosefs.Build,
		"juicefs":    juicefs.Build,
		"thin":       thin.Build,
		"efc":        efc.Build,
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
