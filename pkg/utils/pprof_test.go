/*
Copyright 2023 The Fluid Authors.

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
