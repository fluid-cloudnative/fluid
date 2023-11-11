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

package utils

import (
	"net/http"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func TestNewPprofServer(t *testing.T) {
	setupLog := ctrl.Log.WithName("test")
	pprofAddr := "127.0.0.1:6060"
	NewPprofServer(setupLog, pprofAddr, false)
	paths := []string{
		"/debug/pprof/", "/debug/pprof/cmdline", "/debug/pprof/profile", "/debug/pprof/symbol", "/debug/pprof/trace",
	}
	time.Sleep(10 * time.Second)
	for _, p := range paths {
		resp, err := http.Get("http://127.0.0.1:6060" + p)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code is %d", resp.StatusCode)
		}
	}
}

func TestNewPprofServerWithDevelopment(t *testing.T) {
	setupLog := ctrl.Log.WithName("test")
	pprofAddr := "127.0.0.1:6061"
	NewPprofServer(setupLog, pprofAddr, true)
	paths := []string{
		"/debug/pprof/", "/debug/pprof/cmdline", "/debug/pprof/profile", "/debug/pprof/symbol", "/debug/pprof/trace", "/debug/fgprof",
	}
	time.Sleep(10 * time.Second)
	for _, p := range paths {
		resp, err := http.Get("http://127.0.0.1:6061" + p)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status code is %d", resp.StatusCode)
		}
	}
}
