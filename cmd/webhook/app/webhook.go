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
	"os"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	fluidwebhook "github.com/fluid-cloudnative/fluid/pkg/webhook"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	webhookName = "setup"
)

var (
	setupLog = ctrl.Log.WithName(webhookName)
	scheme   = runtime.NewScheme()
)

var (
	development bool
	metricsAddr string
	webhookPort int
	certDir     string
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
	webhookCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	webhookCmd.Flags().IntVar(&webhookPort, "webhook-port", 9443, "Admission webhook listen address.")
	webhookCmd.Flags().StringVar(&certDir, "webhook-cert-dir", "/etc/k8s-webhook-server/certs", "Admission webhook cert/key dir.")

	webhookCmd.Flags().AddGoFlagSet(flag.CommandLine)
}

func handle() {

	// print fluid version
	fluid.LogVersion()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		setupLog.Error(err, "can not get kube config")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               webhookPort,
		CertDir:            certDir,
	})

	if err != nil {
		setupLog.Error(err, "initialize controller manager failed")
		os.Exit(1)
	}

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

	client, err := client.New(cfg, client.Options{})
	if err != nil {
		setupLog.Error(err, "initialize kube client failed")
		os.Exit(1)
	}

	// patch adminssion web hook  ca bundle
	certBuilder := fluidwebhook.NewCertificateBuilder(client, setupLog)
	if err := certBuilder.BuildAndSyncCABundle(common.WebhookServiceName, common.WebhookName, certDir); err != nil {
		setupLog.Error(err, "patch webhook CABundle failed")
		os.Exit(1)
	}

	if err := fluidwebhook.Register(mgr); err != nil {
		setupLog.Error(err, "register fluid webhook handler failed")
		os.Exit(1)
	}
	setupLog.Info("starting webhook-manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "start webhook handler failed")
		os.Exit(1)
	}

}
