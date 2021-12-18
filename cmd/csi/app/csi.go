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
package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	csi "github.com/fluid-cloudnative/fluid/pkg/csi/fuse"
)

var (
	endpoint string
	nodeID   string

	metricsAddr string
	pprofAddr   string
)

var scheme = runtime.NewScheme()

var csiCmd = &cobra.Command{
	Use:   "start",
	Short: "start fluid driver on node",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
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

	csiCmd.Flags().StringVarP(&nodeID, "nodeid", "", "", "node id")
	if err := csiCmd.MarkFlagRequired("nodeid"); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	csiCmd.Flags().StringVarP(&endpoint, "endpoint", "", "", "CSI endpoint")
	if err := csiCmd.MarkFlagRequired("endpoint"); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	csiCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metrics endpoint binds to.")
	csiCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")

	csiCmd.Flags().AddGoFlagSet(flag.CommandLine)

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
