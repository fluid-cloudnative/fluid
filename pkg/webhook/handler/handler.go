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

package handler

import (
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	setupLog = ctrl.Log.WithName("handler")
)

type GateFunc func() (enabled bool)

var (
	// HandlerMap contains all admission webhook handlers.
	HandlerMap   = map[string]common.AdmissionHandler{}
	handlerGates = map[string]GateFunc{}
)

func addHandlers(m map[string]common.AdmissionHandler) {
	addHandlersWithGate(m, nil)
}

func addHandlersWithGate(m map[string]common.AdmissionHandler, fn GateFunc) {
	for path, handler := range m {
		if len(path) == 0 {
			setupLog.Info("Skip handler with empty path.", "handler", handler)
			continue
		}
		if path[0] != '/' {
			path = "/" + path
		}
		_, found := HandlerMap[path]
		if found {
			setupLog.Info("error: conflicting webhook builder path in handler map", "path", path)
			os.Exit(1)
		}
		HandlerMap[path] = handler
		if fn != nil {
			handlerGates[path] = fn
		}
	}
}

func filterActiveHandlers() {
	disablePaths := sets.NewString()
	for path := range HandlerMap {
		if fn, ok := handlerGates[path]; ok {
			if !fn() {
				disablePaths.Insert(path)
			}
		}
	}
	for _, path := range disablePaths.List() {
		delete(HandlerMap, path)
	}
}
