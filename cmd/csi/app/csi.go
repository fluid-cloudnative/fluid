/*
Copyright 2021 The Fluid Authors.

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
	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	csi "github.com/fluid-cloudnative/fluid/pkg/csi/fuse"
	csimanager "github.com/fluid-cloudnative/fluid/pkg/csi/manager"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/pprof"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

var (
	endpoint          string
	nodeID            string
	metricsAddr       string
	pprofAddr         string
	recoverFusePeriod int
)

const defaultKubeletTimeout = 10

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

	startCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metrics endpoint binds to.")
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	startCmd.Flags().AddGoFlagSet(flag.CommandLine)

	// start csi manager
	startCmd.Flags().IntVar(&recoverFusePeriod, "recover-fuse-period", -1, "CSI manager sync pods period, in seconds")
}

func ErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "%s", err.Error())
	os.Exit(1)
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

	ctx := ctrl.SetupSignalHandler()
	go func() {
		if err := mgr.Start(ctx); err != nil {
			panic(fmt.Sprintf("unable to start controller manager due to error %v", err))
		}
	}()

	if recoverFusePeriod > 0 {
		if err := manageStart(mgr.GetClient()); err != nil {
			panic(fmt.Sprintf("unable to start manager due to error %v", err))
		}
	}

	d := csi.NewDriver(nodeID, endpoint, mgr.GetClient())
	d.Run()
}

func manageStart(k8sClient client.Client) (err error) {
	glog.V(3).Infoln("start csi manager")

	// get CSI sa token
	tokenByte, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return
	}
	token := string(tokenByte)

	glog.V(3).Infoln("start kubelet client")
	nodeIp := os.Getenv("NODE_IP")
	kubeletClientCert := os.Getenv("KUBELET_CLIENT_CERT")
	kubeletClientKey := os.Getenv("KUBELET_CLIENT_KEY")
	var kubeletTimeout int
	if os.Getenv("KUBELET_TIMEOUT") != "" {
		if kubeletTimeout, err = strconv.Atoi(os.Getenv("KUBELET_TIMEOUT")); err != nil {
			glog.Errorf("parse kubelet timeout error: %v", err)
			return
		}
	} else {
		kubeletTimeout = defaultKubeletTimeout
	}
	glog.V(3).Infof("get node ip: %s", nodeIp)
	kubeletClient, err := kubelet.NewKubeletClient(&kubelet.KubeletClientConfig{
		Address: nodeIp,
		Port:    10250,
		TLSClientConfig: rest.TLSClientConfig{
			ServerName: "kubelet",
			CertFile:   kubeletClientCert,
			KeyFile:    kubeletClientKey,
		},
		BearerToken: token,
		HTTPTimeout: time.Duration(kubeletTimeout) * time.Second,
	})
	if err != nil {
		glog.Error(err)
		return
	}
	driver := csimanager.NewPodDriver(k8sClient)
	m := csimanager.Manager{
		KubeletClient: kubeletClient,
		Driver:        driver,
	}

	go m.Run(recoverFusePeriod, wait.NeverStop)

	return
}
