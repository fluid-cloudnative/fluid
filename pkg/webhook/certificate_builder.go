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
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/api/admissionregistration/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	genCertFormat = "%s --service %s --namespace %s --certDir %s"
)

type CertificateBuilder struct {
	log logr.Logger
	client.Client
}

func NewCertificateBuilder(c client.Client, log logr.Logger) *CertificateBuilder {
	ch := &CertificateBuilder{Client: c, log: log}
	return ch
}

// BuildAndSyncCABundle use service name and namespace generate webhook caBundle
// and patch the caBundle to MutatingWebhookConfiguration
func (c *CertificateBuilder) BuildAndSyncCABundle(svcName, webhookName, cerPath string) error {

	ns, err := utils.GetEnvByKey(common.MyPodNamespace)
	if err != nil {
		return errors.Wrapf(err, "get namespace from env failed, env key:%s", common.MyPodNamespace)
	}
	c.log.Info("start generate certificate", "service", svcName, "namespace", ns, "cert dir", cerPath)

	ca, err := c.genCA(ns, svcName, common.CertificationGenerateFile, cerPath)
	if err != nil {
		return err
	}

	return c.PatchCABundle(webhookName, ca)
}

// genCA use shell script to generate the caBundle
func (c *CertificateBuilder) genCA(ns, svc, certFile, certPath string) ([]byte, error) {

	genCertCmd := fmt.Sprintf(genCertFormat, certFile, svc, ns, certPath)

	c.log.Info("generate certification file", "commands", genCertCmd)

	cmd := exec.Command("/bin/sh", "-c", genCertCmd)
	stdout, err := cmd.Output()
	if err != nil {
		c.log.Error(err, "fail to generate certificate", "output", stdout)
		return nil, err
	}

	ca, err := ioutil.ReadFile(certPath + "/ca.crt")

	if err != nil {
		c.log.Error(err, "fail to load certificate ca.crt file", "path", certPath)
		return nil, err
	}

	c.log.Info("generate and load certification ca.crt file success")
	return ca, nil
}

// PatchCABundle patch the caBundle to MutatingWebhookConfiguration
func (c *CertificateBuilder) PatchCABundle(webHookName string, ca []byte) error {

	var m v1beta1.MutatingWebhookConfiguration

	c.log.Info("start patch MutatingWebhookConfiguration caBundle", "name", webHookName)

	ctx := context.Background()

	if err := c.Get(ctx, client.ObjectKey{Name: webHookName}, &m); err != nil {
		c.log.Error(err, "fail to get mutatingWebHook", "name", webHookName)
		return err
	}

	current := m.DeepCopy()
	for i := range m.Webhooks {
		m.Webhooks[i].ClientConfig.CABundle = ca
	}

	if err := c.Patch(ctx, &m, client.MergeFrom(current)); err != nil {
		c.log.Error(err, "fail to patch CABundle to mutatingWebHook", "name", webHookName)
		return err
	}

	c.log.Info("finished patch MutatingWebhookConfiguration caBundle", "name", webHookName)

	return nil
}
