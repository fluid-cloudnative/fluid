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
<<<<<<< HEAD
	"os"

	"github.com/fluid-cloudnative/fluid/cmd/dataset/app"
)

func main() {
	command := app.NewDatasetCommand()
	if err := command.Execute(); err != nil {
=======
	"fmt"
	"github.com/fluid-cloudnative/fluid"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	databackupctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/databackup"
	dataloadctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/dataload"
	datasetctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/dataset"
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
	pprofAddr            string
)

var cmd = &cobra.Command{
	Use:   "dataset-controller",
	Short: "controller for dataset",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start dataset-controller in Kubernetes",
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
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
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

	utils.NewPprofServer(setupLog, pprofAddr)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "89759796.data.fluid.io",
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start dataset manager")
		os.Exit(1)
	}

	if err = (&datasetctl.DatasetReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("datasetctl").WithName("Dataset"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("Dataset"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Dataset")
		os.Exit(1)
	}

	if err = (dataloadctl.NewDataLoadReconciler(mgr.GetClient(),
		ctrl.Log.WithName("dataloadctl").WithName("DataLoad"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("DataLoad"),
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataLoad")
		os.Exit(1)
	}

	if err = (databackupctl.NewDataBackupReconciler(mgr.GetClient(),
		ctrl.Log.WithName("databackupctl").WithName("DataBackup"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("DataBackup"),
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataBackup")
		os.Exit(1)
	}

	setupLog.Info("starting dataset-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running dataset-controller")
>>>>>>> master
		os.Exit(1)
	}
}
