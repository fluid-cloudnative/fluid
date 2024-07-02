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

package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"time"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/csi"
	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	endpoint              string
	nodeID                string
	metricsAddr           string
	pprofAddr             string
	pruneFs               []string
	prunePath             string
	kubeletKubeConfigPath string
)

var scheme = runtime.NewScheme()

var startCmd = &cobra.Command{
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

	startCmd.Flags().StringVarP(&nodeID, "nodeid", "", "", "node id")
	if err := startCmd.MarkFlagRequired("nodeid"); err != nil {
		ErrorAndExit(err)
	}

	startCmd.Flags().StringVarP(&endpoint, "endpoint", "", "", "CSI endpoint")
	if err := startCmd.MarkFlagRequired("endpoint"); err != nil {
		ErrorAndExit(err)
	}

	startCmd.Flags().StringSliceVarP(&pruneFs, "prune-fs", "", []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "fuse.juicefs", "fuse.goosefs-fuse", "ossfs"}, "Prune fs to add in /etc/updatedb.conf, separated by comma")
	startCmd.Flags().StringVarP(&prunePath, "prune-path", "", "/runtime-mnt", "Prune path to add in /etc/updatedb.conf")
	startCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metrics endpoint binds to.")
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	startCmd.Flags().StringVarP(&kubeletKubeConfigPath, "kubelet-kube-config", "", "/etc/kubernetes/kubelet.conf", "The file path to kubelet kube config")
	utilfeature.DefaultMutableFeatureGate.AddFlag(startCmd.Flags())
	startCmd.Flags().AddGoFlagSet(flag.CommandLine)
}

func ErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s", err.Error())
	os.Exit(1)
}

func handle() {
	// startReaper()
	fluid.LogVersion()

	if pprofAddr != "" {
		newPprofServer(pprofAddr)
	}

	// the default webserver port is 9443, no need to set.
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
	})

	if err != nil {
		panic(fmt.Sprintf("csi: unable to create controller manager due to error %v", err))
	}

	runningContext := config.RunningContext{
		Config: config.Config{
			NodeId:            nodeID,
			Endpoint:          endpoint,
			PruneFs:           pruneFs,
			PrunePath:         prunePath,
			KubeletConfigPath: kubeletKubeConfigPath,
		},
		VolumeLocks: utils.NewVolumeLocks(),
	}
	if err = csi.SetupWithManager(mgr, runningContext); err != nil {
		panic(fmt.Sprintf("unable to set up manager due to error %v", err))
	}

	ctx := ctrl.SetupSignalHandler()
	if err = mgr.Start(ctx); err != nil {
		panic(fmt.Sprintf("unable to start controller recover due to error %v", err))
	}
}

func newPprofServer(pprofAddr string) {
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
