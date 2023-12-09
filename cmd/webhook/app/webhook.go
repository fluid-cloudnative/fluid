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
	"flag"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	"os"

	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	fluidwebhook "github.com/fluid-cloudnative/fluid/pkg/webhook"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/handler"
)

const (
	webhookName = "webhook"
)

var (
	setupLog = ctrl.Log.WithName(webhookName)
	scheme   = runtime.NewScheme()
)

var (
	development   bool
	fullGoProfile bool
	metricsAddr   string
	webhookPort   int
	certDir       string
	pprofAddr     string
)

var webhookCmd = &cobra.Command{
	Use:   "start",
	Short: "fluid admission webhook server",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {

	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	webhookCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	webhookCmd.Flags().BoolVarP(&development, "development", "", false, "Enable development mode for fluid controller.")
	webhookCmd.Flags().BoolVarP(&fullGoProfile, "full-go-profile", "", false, "Enable full golang profile mode for fluid controller.")
	webhookCmd.Flags().IntVar(&webhookPort, "webhook-port", 9443, "Admission webhook listen address.")
	webhookCmd.Flags().StringVar(&certDir, "webhook-cert-dir", "/etc/k8s-webhook-server/certs", "Admission webhook cert/key dir.")
	webhookCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	webhookCmd.Flags().AddGoFlagSet(flag.CommandLine)
}

func handle() {

	// print fluid version
	fluid.LogVersion()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = development
	}, func(o *zap.Options) {
		o.ZapOpts = append(o.ZapOpts, zapOpt.AddCaller())
	}, func(o *zap.Options) {
		var encCfg zapcore.EncoderConfig
		if !development {
			encCfg = zapOpt.NewProductionEncoderConfig()
		} else {
			encCfg = zapOpt.NewDevelopmentEncoderConfig()
		}
		encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		o.Encoder = zapcore.NewConsoleEncoder(encCfg)
	}))

	cfg, err := ctrl.GetConfig()
	if err != nil {
		setupLog.Error(err, "can not get kube config")
		os.Exit(1)
	}

	utils.NewPprofServer(setupLog, pprofAddr, fullGoProfile)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               webhookPort,
		CertDir:            certDir,
		LeaderElection:     false,
		LeaderElectionID:   "webhook.data.fluid.io",
		NewCache: cache.BuilderWithOptions(cache.Options{
			Scheme: scheme,
			SelectorsByObject: cache.SelectorsByObject{
				&admissionregistrationv1.MutatingWebhookConfiguration{}: {
					Field: fields.SelectorFromSet(fields.Set{"metadata.name": common.WebhookName}),
				},
			},
		}),
	})

	if err != nil {
		setupLog.Error(err, "initialize controller manager failed")
		os.Exit(1)
	}

	// get client from mgr
	client, err := client.New(cfg, client.Options{})
	if err != nil {
		setupLog.Error(err, "initialize kube client failed")
		os.Exit(1)
	}

	// initialize the cert files
	certBuilder := fluidwebhook.NewCertificateBuilder(client, setupLog)
	caCert, err := certBuilder.BuildOrSyncCABundle(common.WebhookServiceName, certDir)
	if err != nil || len(caCert) == 0 {
		setupLog.Error(err, "initialize webhook CABundle failed")
		os.Exit(1)
	}

	// watch the WebhookConfiguration to patch it
	if err = watch.SetupWatcherForWebhook(mgr, certBuilder, caCert); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "webhook")
		os.Exit(1)
	}

	// register admission handlers
	handler.Register(mgr, mgr.GetClient(), setupLog)

	// register pod mutating handlers
	err = plugins.RegisterMutatingHandlers(mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "get the register plugins from configmap occurs error")
		os.Exit(1)
	}

	setupLog.Info("Register Handler")

	setupLog.Info("starting webhook-manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "start webhook handler failed")
		os.Exit(1)
	}

}
