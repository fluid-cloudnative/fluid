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
	"flag"
	"os"

	zapOpt "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports

	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	alluxioctl "github.com/cloudnativefluid/fluid/pkg/controllers/v1alpha1/alluxio"
	dataloadctl "github.com/cloudnativefluid/fluid/pkg/controllers/v1alpha1/dataload"
	datasetctl "github.com/cloudnativefluid/fluid/pkg/controllers/v1alpha1/dataset"
	"github.com/cloudnativefluid/fluid/pkg/ddc/alluxio"
	"github.com/cloudnativefluid/fluid/pkg/ddc/base"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*alluxio.AlluxioEngine)(nil)
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = datav1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		development          bool
	)

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&development, "development", true,
		"Enable development mode for pillar controller.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = development
	}, func(o *zap.Options) {
		o.ZapOpts = append(o.ZapOpts, zapOpt.AddCaller())
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		// MaxConcurrentReconciles: 5,
		Port: 9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&datasetctl.DatasetReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("datasetctl").WithName("Dataset"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Dataset")
		os.Exit(1)
	}
	// if err = (&alluxioctl.RuntimeReconciler{
	// 	Client: mgr.GetClient(),
	// 	Log:    ctrl.Log.WithName("alluxioctl").WithName("AlluxioRuntime"),
	// 	Scheme: mgr.GetScheme(),
	// }).SetupWithManager(mgr); err != nil {
	// 	setupLog.Error(err, "unable to create controller", "controller", "AlluxioRuntime")
	// 	os.Exit(1)
	// }
	if err = (alluxioctl.NewRuntimeReconciler(mgr.GetClient(),
		ctrl.Log.WithName("alluxioctl").WithName("AlluxioRuntime"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("AlluxioRuntime"),
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AlluxioRuntime")
		os.Exit(1)
	}
	if err = (dataloadctl.NewDataLoadReconciler(mgr.GetClient(),
		ctrl.Log.WithName("alluxioctl").WithName("AlluxioDataLoad"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("AlluxioDataLoad"),
	)).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AlluxioDataLoad")
		os.Exit(1)
	}
	//if err = (&dataload.DataLoadReconciler{
	//	Client: mgr.GetClient(),
	//	Log:    ctrl.Log.WithName("alluxioctl").WithName("AlluxioDataLoad"),
	//	Scheme: mgr.GetScheme(),
	//}).SetupWithManager(mgr); err != nil {
	//	setupLog.Error(err, "unable to create controller", "controller", "AlluxioDataLoad")
	//	os.Exit(1)
	//}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
