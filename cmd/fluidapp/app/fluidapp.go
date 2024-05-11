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
	"os"

	"github.com/fluid-cloudnative/fluid"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/fluidapp"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	maxConcurrentReconciles int
	pprofAddr               string
)

var fluidAppCmd = &cobra.Command{
	Use:   "start",
	Short: "start fluidapp-controller in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	fluidAppCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	fluidAppCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	fluidAppCmd.Flags().StringVarP(&leaderElectionNamespace, "leader-election-namespace", "", "fluid-system", "The namespace in which the leader election resource will be created.")
	fluidAppCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	fluidAppCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	fluidAppCmd.Flags().IntVar(&maxConcurrentReconciles, "runtime-workers", 3, "Set max concurrent workers for Fluid App controller")
}

func handle() {
	fluid.LogVersion()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = development
	}, func(o *zap.Options) {
		o.ZapOpts = append(o.ZapOpts, zapOpt.AddCaller())
	}, func(o *zap.Options) {
		if !development {
			encCfg := zapOpt.NewProductionEncoderConfig()
			encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
			encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			o.Encoder = zapcore.NewConsoleEncoder(encCfg)
		}
	}))

	utils.NewPprofServer(setupLog, pprofAddr, development)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionID:        "fluidapp.data.fluid.io",
		Port:                    9443,
		NewCache:                fluidapp.NewCache(scheme),
	})
	if err != nil {
		setupLog.Error(err, "unable to start fluid app manager")
		os.Exit(1)
	}

	controllerOptions := controller.Options{
		MaxConcurrentReconciles: maxConcurrentReconciles,
		// Log:                     ctrl.Log.WithName("appctrl"),
	}
	if err = (fluidapp.NewFluidAppReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("appctrl"),
		mgr.GetEventRecorderFor("FluidApp"),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FluidApp")
		os.Exit(1)
	}

	setupLog.Info("starting fluidapp-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running fluidapp-controller")
		os.Exit(1)
	}
}
