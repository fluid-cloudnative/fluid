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

package app

import (
	"os"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc"
	"k8s.io/apimachinery/pkg/util/net"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	efcctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/efc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*efc.EFCEngine)(nil)

	eventDriven             bool
	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	maxConcurrentReconciles int
	pprofAddr               string
	portRange               string

	kubeClientQPS   float32
	kubeClientBurst int
)

// configuration for controllers' rate limiter
var (
	controllerWorkqueueDefaultSyncBackoffStr string
	controllerWorkqueueMaxSyncBackoffStr     string
	controllerWorkqueueQPS                   int
	controllerWorkqueueBurst                 int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start efcruntime-controller in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	startCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	startCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	startCmd.Flags().StringVarP(&leaderElectionNamespace, "leader-election-namespace", "", "fluid-system", "The namespace in which the leader election resource will be created.")
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	startCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	startCmd.Flags().BoolVar(&eventDriven, "event-driven", true, "The reconciler's loop strategy. if it's false, it indicates period driven.")
	startCmd.Flags().StringVar(&portRange, "runtime-node-port-range", "16000-17999", "Set available port range for EFC")
	startCmd.Flags().Float32VarP(&kubeClientQPS, "kube-api-qps", "", 20, "QPS to use while talking with kubernetes apiserver.")   // 20 is the default qps in controller-runtime
	startCmd.Flags().IntVarP(&kubeClientBurst, "kube-api-burst", "", 30, "Burst to use while talking with kubernetes apiserver.") // 30 is the default burst in controller-runtime
	startCmd.Flags().StringVar(&controllerWorkqueueDefaultSyncBackoffStr, "workqueue-default-sync-backoff", "5ms", "base backoff period for failed reconciliation in controller's workqueue")
	startCmd.Flags().StringVar(&controllerWorkqueueMaxSyncBackoffStr, "workqueue-max-sync-backoff", "1000s", "max backoff period for failed reconciliation in controller's workqueue")
	startCmd.Flags().IntVar(&controllerWorkqueueQPS, "workqueue-qps", 10, "qps limit value for controller's workqueue")
	startCmd.Flags().IntVar(&controllerWorkqueueBurst, "workqueue-burst", 100, "burst limit value for controller's workqueue")
	startCmd.Flags().IntVar(&maxConcurrentReconciles, "runtime-workers", 3, "Set max concurrent workers for efc controller")
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

	// the default webhook server port is 9443, no need to set
	mgr, err := ctrl.NewManager(controllers.GetConfigOrDieWithQPSAndBurst(kubeClientQPS, kubeClientBurst), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionID:        "efc.data.fluid.io",
		NewClient:               controllers.NewFluidControllerClient,
	})
	if err != nil {
		setupLog.Error(err, "unable to start efcruntime manager")
		os.Exit(1)
	}

	defaultSyncBackoff, err := time.ParseDuration(controllerWorkqueueDefaultSyncBackoffStr)
	if err != nil {
		setupLog.Error(err, "workqueue-default-sync-backoff is not a valid duration, please use string like \"100ms\", \"5s\", \"3m\", ...")
		os.Exit(1)
	}

	maxSyncBackoff, err := time.ParseDuration(controllerWorkqueueMaxSyncBackoffStr)
	if err != nil {
		setupLog.Error(err, "workqueue-max-sync-backoff is not a valid duration, please use string like \"100ms\", \"5s\", \"3m\", ...)")
		os.Exit(1)
	}

	controllerOptions := controller.Options{
		MaxConcurrentReconciles: maxConcurrentReconciles,
		RateLimiter:             controllers.NewFluidControllerRateLimiter(defaultSyncBackoff, maxSyncBackoff, controllerWorkqueueQPS, controllerWorkqueueBurst),
	}

	if err = (efcctl.NewRuntimeReconciler(mgr.GetClient(),
		ctrl.Log.WithName("efcctl").WithName("EFCRuntime"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("EFCRuntime"),
	)).SetupWithManager(mgr, controllerOptions, eventDriven); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "EFCRuntime")
		os.Exit(1)
	}

	pr, err := net.ParsePortRange(portRange)
	if err != nil {
		setupLog.Error(err, "can't parse port range. Port range must be like <min>-<max>")
		os.Exit(1)
	}
	setupLog.Info("port range parsed", "port range", pr.String())

	// portallocator.SetupRuntimePortAllocator(mgr.GetClient(), pr, efc.GetReservedPorts)
	portallocator.SetupRuntimePortAllocatorWithType(mgr.GetClient(),
		pr,
		portallocator.Random,
		efc.GetReservedPorts)

	if err := mgr.Add(efc.NewSessMgrInitializer(mgr.GetClient())); err != nil {
		setupLog.Error(err, "fail to add SessMgrInitializer to controller manager")
		os.Exit(1)
	}

	setupLog.Info("starting efcruntime-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem efcruntime-controller")
		os.Exit(1)
	}
}
