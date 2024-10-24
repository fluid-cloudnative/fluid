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
	"context"
	"errors"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/felixge/fgprof"

	"github.com/go-logr/logr"
)

func NewPprofServer(setupLog logr.Logger, pprofAddr string, enableFullGoProfile bool) {
	if pprofAddr != "" {
		setupLog.Info("Enabling pprof", "pprof address", pprofAddr)
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		if enableFullGoProfile {
			mux.Handle("/debug/fgprof", fgprof.Handler())
		}

		pprofServer := http.Server{
			Addr:    pprofAddr,
			Handler: mux,
		}
		setupLog.Info("Starting pprof HTTP server", "pprof server address", pprofServer.Addr)

		go func() {
			go func() {
				ctx := context.Background()
				<-ctx.Done()

				ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Minute)
				defer cancelFunc()

				if err := pprofServer.Shutdown(ctx); err != nil {
					setupLog.Error(err, "Failed to shutdown debug HTTP server")
				}
			}()

			if err := pprofServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				setupLog.Error(err, "Failed to start debug HTTP server")
				panic(err)
			}
		}()
	}
}
