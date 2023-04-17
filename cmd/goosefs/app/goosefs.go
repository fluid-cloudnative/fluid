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
	"os"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	goosefsctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/goosefs"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs"
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
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*goosefs.GooseFSEngine)(nil)

	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	portRange               string
	maxConcurrentReconciles int
	pprofAddr               string
	portAllocatePolicy      string
)

var cmd = &cobra.Command{
	Use:   "goosefsruntime-controller",
	Short: "Controller for goosefsruntime",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start goosefsruntime-controller in Kubernetes",
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
	startCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	startCmd.Flags().StringVar(&portRange, "runtime-node-port-range", "20000-25000", "Set available port range for GooseFS")
	startCmd.Flags().StringVar(&portAllocatePolicy, "port-allocate-policy", "random", "Set port allocating policy, available choice is bitmap or random(default random).")
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	startCmd.Flags().IntVar(&maxConcurrentReconciles, "runtime-workers", 3, "Set max concurrent workers for GooseFSRuntime controller")
	cmd.AddCommand(startCmd)
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
		LeaderElectionID:        "goosefs.data.fluid.io",
		Port:                    9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start goosefsruntime manager")
		os.Exit(1)
	}

	controllerOptions := controller.Options{
		MaxConcurrentReconciles: maxConcurrentReconciles,
	}

	if err = (goosefsctl.NewRuntimeReconciler(mgr.GetClient(),
		ctrl.Log.WithName("goosefsctl").WithName("GooseFSRuntime"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("GooseFSRuntime"),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GooseFSRuntime")
		os.Exit(1)
	}

	pr, err := net.ParsePortRange(portRange)
	if err != nil {
		setupLog.Error(err, "can't parse port range. Port range must be like <min>-<max>")
		os.Exit(1)
	}
	setupLog.Info("port range parsed", "port range", pr.String())

	err = portallocator.SetupRuntimePortAllocator(mgr.GetClient(), pr, portAllocatePolicy, goosefs.GetReservedPorts)
	if err != nil {
		setupLog.Error(err, "failed to setup runtime port allocator")
		os.Exit(1)
	}
	setupLog.Info("Set up runtime port allocator", "policy", portAllocatePolicy)

	setupLog.Info("starting goosefsruntime-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem goosefsruntime-controller")
		os.Exit(1)
	}

}
