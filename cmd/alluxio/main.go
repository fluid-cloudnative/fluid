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
	"fmt"
	"github.com/fluid-cloudnative/fluid"
	alluxioctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/alluxio"
	"github.com/spf13/cobra"
	"os"

	"go.uber.org/zap/zapcore"

	zapOpt "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*alluxio.AlluxioEngine)(nil)

	short                bool
	metricsAddr          string
	enableLeaderElection bool
	development          bool
)

var cmd = &cobra.Command{
	Use:   "alluxioruntime-controller",
	Short: "Controller for alluxioruntime",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start alluxioruntime-controller in Kubernetes",
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
	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	startCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	startCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	startCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	versionCmd.Flags().BoolVar(&short, "short", false, "print just the short version info")

	cmd.AddCommand(startCmd)
	cmd.AddCommand(versionCmd)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
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

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "7857424864.data.fluid.io",
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start alluxioruntime manager")
		os.Exit(1)
	}

	if err = (alluxioctl.NewRuntimeReconciler(mgr.GetClient(),
		ctrl.Log.WithName("alluxioctl").WithName("AlluxioRuntime"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("AlluxioRuntime"),
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AlluxioRuntime")
		os.Exit(1)
	}

	setupLog.Info("starting alluxioruntime-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem alluxioruntime-controller")
		os.Exit(1)
	}
}
