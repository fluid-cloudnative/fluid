/*
Copyright 2026 The Fluid Authors.

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
	workloadv1alpha1 "github.com/fluid-cloudnative/fluid/api/workload/v1alpha1"
	advancedstatefulset "github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/advancedstatefulset"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	pprofAddr               string
)

var workloadCmd = &cobra.Command{
	Use:   "start",
	Short: "start workload-controller in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = workloadv1alpha1.AddToScheme(scheme)

	workloadCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8084", "The address the metric endpoint binds to.")
	workloadCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager.")
	workloadCmd.Flags().StringVarP(&leaderElectionNamespace, "leader-election-namespace", "", "fluid-system", "The namespace in which the leader election resource will be created.")
	workloadCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for workload controller.")
	workloadCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
}

// NewWorkloadCommand creates the cobra command for the workload controller.
func NewWorkloadCommand() *cobra.Command {
	return workloadCmd
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
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionID:        "workload.fluid.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start workload controller manager")
		os.Exit(1)
	}

	if err := advancedstatefulset.Add(mgr); err != nil {
		setupLog.Error(err, "unable to add AdvancedStatefulSet controller")
		os.Exit(1)
	}

	setupLog.Info("starting workload-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running workload-controller")
		os.Exit(1)
	}
}
