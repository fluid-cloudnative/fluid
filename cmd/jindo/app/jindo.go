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
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindocache"

	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	jindoctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/jindo"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*jindo.JindoEngine)(nil)
	_ base.Implement = (*jindofsx.JindoFSxEngine)(nil)
	_ base.Implement = (*jindocache.JindoCacheEngine)(nil)

	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	// The new mode
	eventDriven             bool
	maxConcurrentReconciles int
	pprofAddr               string
	portRange               string
	portAllocatePolicy      string

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

var jindoCmd = &cobra.Command{
	Use:   "start",
	Short: "start jindoruntime-controller in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	jindoCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	jindoCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	jindoCmd.Flags().StringVarP(&leaderElectionNamespace, "leader-election-namespace", "", "fluid-system", "The namespace in which the leader election resource will be created.")
	jindoCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	jindoCmd.Flags().StringVar(&portRange, "runtime-node-port-range", "18000-19999", "Set available port range for Jindo")
	jindoCmd.Flags().StringVar(&portAllocatePolicy, "port-allocate-policy", "random", "Set port allocating policy, available choice is bitmap or random(default random).")
	jindoCmd.Flags().IntVar(&maxConcurrentReconciles, "runtime-workers", 3, "Set max concurrent workers for JindoRuntime controller")
	jindoCmd.Flags().BoolVar(&eventDriven, "event-driven", true, "The reconciler's loop strategy. if it's false, it indicates period driven.")
	jindoCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	jindoCmd.Flags().Float32VarP(&kubeClientQPS, "kube-api-qps", "", 20, "QPS to use while talking with kubernetes apiserver.")   // 20 is the default qps in controller-runtime
	jindoCmd.Flags().IntVarP(&kubeClientBurst, "kube-api-burst", "", 30, "Burst to use while talking with kubernetes apiserver.") // 30 is the default burst in controller-runtime
	jindoCmd.Flags().StringVar(&controllerWorkqueueDefaultSyncBackoffStr, "workqueue-default-sync-backoff", "5ms", "base backoff period for failed reconciliation in controller's workqueue")
	jindoCmd.Flags().StringVar(&controllerWorkqueueMaxSyncBackoffStr, "workqueue-max-sync-backoff", "1000s", "max backoff period for failed reconciliation in controller's workqueue")
	jindoCmd.Flags().IntVar(&controllerWorkqueueQPS, "workqueue-qps", 10, "qps limit value for controller's workqueue")
	jindoCmd.Flags().IntVar(&controllerWorkqueueBurst, "workqueue-burst", 100, "burst limit value for controller's workqueue")
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
		LeaderElectionID:        "jindo.data.fluid.io",
		NewClient:               controllers.NewFluidControllerClient,
	})
	if err != nil {
		setupLog.Error(err, "unable to start jindoruntime manager")
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

	if err = (jindoctl.NewRuntimeReconciler(mgr.GetClient(),
		ctrl.Log.WithName("jindoctl").WithName("JindoRuntime"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("JindoRuntime"),
	)).SetupWithManager(mgr, controllerOptions, eventDriven); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "JindoRuntime")
		os.Exit(1)
	}

	pr, err := net.ParsePortRange(portRange)
	if err != nil {
		setupLog.Error(err, "can't parse port range. Port range must be like <min>-max")
		os.Exit(1)
	}
	setupLog.Info("port range parsed", "port range", pr.String())

	// Register with jindofsx.GetReservedPorts func by default. The function will only be called when users explicitly setting portAllocatePolicy to "bitmap", which
	// is a DEPRECATED port allocation policy that will be removed in the future. When using "bitmap", Fluid may allocate some used ports but have low possibility
	// affecting runtime's deployment.
	err = portallocator.SetupRuntimePortAllocator(mgr.GetClient(), pr, portAllocatePolicy, jindofsx.GetReservedPorts)
	if err != nil {
		setupLog.Error(err, "failed to setup runtime port allocator")
		os.Exit(1)
	}
	setupLog.Info("Set up runtime port allocator", "policy", portAllocatePolicy)

	setupLog.Info("starting jindoruntime-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem jindoruntime-controller")
		os.Exit(1)
	}
}
