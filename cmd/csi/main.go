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

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"net/http"
	"net/http/pprof"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"

	"github.com/fluid-cloudnative/fluid"
	"github.com/fluid-cloudnative/fluid/pkg/csi/fuse"
	"github.com/spf13/cobra"
)

var (
	endpoint    string
	nodeID      string
	short       bool
	metricsAddr string
	pprofAddr   string
)

var scheme = runtime.NewScheme()

var cmd = &cobra.Command{
	Use:   "fluid-csi",
	Short: "CSI based fluid driver for Fuse",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start fluid driver on node",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fluid.PrintVersion(short)
	},
}

func init() {
	// Register k8s-native resources and Fluid CRDs
	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	if err := flag.Set("logtostderr", "true"); err != nil {
		fmt.Printf("Failed to flag.set due to %v", err)
		os.Exit(1)
	}

	startCmd.Flags().StringVarP(&nodeID, "nodeid", "", "", "node id")
	if err := startCmd.MarkFlagRequired("nodeid"); err != nil {
		errorAndExit(err)
	}

	startCmd.Flags().StringVarP(&endpoint, "endpoint", "", "", "CSI endpoint")
	if err := startCmd.MarkFlagRequired("endpoint"); err != nil {
		errorAndExit(err)
	}

	startCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metrics endpoint binds to.")
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")

	versionCmd.Flags().BoolVar(&short, "short", false, "print just the short version info")

	cmd.AddCommand(startCmd)
	cmd.AddCommand(versionCmd)
}

func main() {
	startCmd.Flags().AddGoFlagSet(flag.CommandLine)

	if err := cmd.Execute(); err != nil {
		errorAndExit(err)
	}

	os.Exit(0)
}

func handle() {
	// startReaper()
	fluid.LogVersion()

	if pprofAddr != "" {
		glog.Infof("Enabling pprof with address %s", pprofAddr)
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		pprofServer := http.Server{
			Addr:    pprofAddr,
			Handler: mux,
		}
		glog.Infof("Starting pprof HTTP server at %s", pprofServer.Addr)

		go func() {
			go func() {
				ctx := context.Background()
				<-ctx.Done()

				ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Minute)
				defer cancelFunc()

				if err := pprofServer.Shutdown(ctx); err != nil {
					glog.Error(err, "Failed to shutdown debug HTTP server")
				}
			}()

			if err := pprofServer.ListenAndServe(); !errors.Is(http.ErrServerClosed, err) {
				glog.Error(err, "Failed to start debug HTTP server")
				panic(err)
			}
		}()
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
	})

	if err != nil {
		panic(fmt.Sprintf("csi: unable to create controller manager due to error %v", err))
	}

	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			panic(fmt.Sprintf("unable to start controller manager due to error %v", err))
		}
	}()

	d := csi.NewDriver(nodeID, endpoint, mgr.GetClient())
	d.Run()
}

func errorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s", err.Error())
	os.Exit(1)
}

// /*
// Based on https://github.com/openshift/origin/blob/master/pkg/util/proc/reaper.go

// */
// func startReaper() {
// 	glog.V(4).Infof("Launching reaper")
// 	go func() {
// 		sigs := make(chan os.Signal, 1)
// 		signal.Notify(sigs, syscall.SIGCHLD)
// 		for {
// 			// Wait for a child to terminate
// 			sig := <-sigs
// 			glog.V(4).Infof("Signal received: %v", sig)
// 			for {
// 				// Reap processes
// 				cpid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
// 				if cpid < 1 {
// 					break
// 				}

// 				glog.V(4).Infof("Reaped process with pid %d", cpid)
// 			}
// 		}
// 	}()
// }
