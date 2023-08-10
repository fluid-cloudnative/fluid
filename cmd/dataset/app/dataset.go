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
	"os"
	"time"

	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	databackupctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/databackup"
	dataflowctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/dataflow"
	dataloadctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/dataload"
	datamigratectl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/datamigrate"
	dataprocessctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/dataprocess"
	datasetctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/dataset"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*alluxio.AlluxioEngine)(nil)

	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	pprofAddr               string
	maxConcurrentReconciles int
)

var datasetCmd = &cobra.Command{
	Use:   "start",
	Short: "start dataset-controller in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	datasetCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	datasetCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	datasetCmd.Flags().StringVarP(&leaderElectionNamespace, "leader-election-namespace", "", "fluid-system", "The namespace in which the leader election resource will be created.")
	datasetCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	datasetCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	datasetCmd.Flags().IntVar(&maxConcurrentReconciles, "reconcile-workers", 3, "Set the number of max concurrent workers for reconciling dataset and dataset operations")
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
		LeaderElectionID:        "dataset.data.fluid.io",
		Port:                    9443,
		NewCache:                NewCache(scheme),
		NewClient:               controllers.NewFluidControllerClient,
	})
	if err != nil {
		setupLog.Error(err, "unable to start dataset manager")
		os.Exit(1)
	}

	controllerOptions := controller.Options{
		MaxConcurrentReconciles: maxConcurrentReconciles,
	}

	if err = (&datasetctl.DatasetReconciler{
		Client:       mgr.GetClient(),
		Log:          ctrl.Log.WithName("datasetctl").WithName("Dataset"),
		Scheme:       mgr.GetScheme(),
		Recorder:     mgr.GetEventRecorderFor("Dataset"),
		ResyncPeriod: time.Duration(5 * time.Second),
	}).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Dataset")
		os.Exit(1)
	}

	if err = (dataloadctl.NewDataLoadReconciler(mgr.GetClient(),
		ctrl.Log.WithName("dataloadctl").WithName("DataLoad"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("DataLoad"),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataLoad")
		os.Exit(1)
	}

	if err = (databackupctl.NewDataBackupReconciler(mgr.GetClient(),
		ctrl.Log.WithName("databackupctl").WithName("DataBackup"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("DataBackup"),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataBackup")
		os.Exit(1)
	}

	if err = (datamigratectl.NewDataMigrateReconciler(mgr.GetClient(),
		ctrl.Log.WithName("datamigratectl").WithName("DataMigrate"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("DataMigrate"),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataMigrate")
		os.Exit(1)
	}

	if err = (dataprocessctl.NewDataProcessReconciler(mgr.GetClient(),
		ctrl.Log.WithName("dataprocessctl").WithName("DataProcess"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("DataProcess"),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DataProcess")
		os.Exit(1)
	}

	if err = (dataflowctl.NewDataFlowReconciler(mgr.GetClient(),
		ctrl.Log.WithName("dataflowctl"),
		mgr.GetEventRecorderFor("DataFlow"),
		time.Duration(5*time.Second),
	)).SetupWithManager(mgr, controllerOptions); err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	setupLog.Info("starting dataset-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running dataset-controller")
		os.Exit(1)
	}
}

func NewCache(scheme *runtime.Scheme) cache.NewCacheFunc {
	selectors := make(cache.SelectorsByObject, 1)

	if compatibility.IsBatchV1CronJobSupported() {
		selectors[&batchv1.CronJob{}] = cache.ObjectSelector{Label: labels.SelectorFromSet(labels.Set{
			common.JobPolicy: common.CronPolicy,
		})}
	} else {
		selectors[&batchv1beta1.CronJob{}] = cache.ObjectSelector{Label: labels.SelectorFromSet(labels.Set{
			common.JobPolicy: common.CronPolicy,
		})}
	}

	return cache.BuilderWithOptions(cache.Options{
		Scheme:            scheme,
		SelectorsByObject: selectors,
	})
}
