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

package webhook

import (
	"context"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admissionreg "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	"k8s.io/client-go/rest"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
)

var setupLog = ctrl.Log.WithName("certificate")

func UpdateCertificate(service, namespace, certPath, webHookName string) error {
	setupLog.Info("start to new a certificate")
	// TODO(LingWei Qiu): store the certificate and share with other replicas
	cmd := exec.Command("sh", "/certificate.sh", "--service", service, "--namespace", namespace, "--certDir", certPath)
	err := cmd.Run()
	if err != nil {
		setupLog.Error(err, "fail to load certificate")
		return err
	}
	caCert, err := ioutil.ReadFile(certPath + "/ca.crt")
	if err != nil {
		setupLog.Error(err, "fail to read ca file")
	}

	// Update MutatingWebHookConfiguration caBundle
	err = updateCaBundle(webHookName, caCert)
	if err != nil {
		setupLog.Error(err, "fail to update CaBundle")
		return err
	}
	return nil
}

func updateCaBundle(mutatingWebHookName string, ca []byte) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		setupLog.Error(err, "fail to get InClusterConfig")
		return err
	}
	admissionClient := admissionreg.NewForConfigOrDie(config)
	mutating := admissionClient.MutatingWebhookConfigurations()
	ctx := context.Background()
	obj, err := mutating.Get(ctx, mutatingWebHookName, metav1.GetOptions{})
	if err != nil {
		setupLog.Error(err, "fail to get mutatingWebHookConfig")
		return err
	}

	for ind := range obj.Webhooks {
		obj.Webhooks[ind].ClientConfig.CABundle = ca
	}
	_, err = mutating.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		setupLog.Error(err, "fail to update mutatingWebHookConfig")
		return err
	}
	return nil
}
