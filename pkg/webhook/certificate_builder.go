/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package webhook

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/utils/webhook/generator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/webhook/writer"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CertificateBuilder struct {
	log logr.Logger
	client.Client
}

func NewCertificateBuilder(c client.Client, log logr.Logger) *CertificateBuilder {
	ch := &CertificateBuilder{Client: c, log: log}
	return ch
}

// BuildOrSyncCABundle use service name and namespace to generate webhook certs
// or sync the certs from the secret
func (c *CertificateBuilder) BuildOrSyncCABundle(svcName, cerPath string) ([]byte, error) {

	ns, err := utils.GetEnvByKey(common.MyPodNamespace)
	if err != nil {
		return []byte{}, errors.Wrapf(err, "get namespace from env failed, env key:%s", common.MyPodNamespace)
	}
	c.log.Info("start generate certificate", "service", svcName, "namespace", ns, "cert dir", cerPath)

	certs, err := c.genCA(ns, svcName, cerPath)
	if err != nil {
		return []byte{}, err
	}

	return certs.CACert, nil
}

// genCA generate the caBundle and store it in secret and local path
func (c *CertificateBuilder) genCA(ns, svc, certPath string) (*generator.Artifacts, error) {

	certWriter, err := writer.NewSecretCertWriter(writer.SecretCertWriterOptions{
		Client: c.Client,
		Secret: &types.NamespacedName{Namespace: ns, Name: common.CertSecretName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new certWriter: %v", err)
	}

	dnsName := generator.ServiceToCommonName(ns, svc)

	certs, _, err := certWriter.EnsureCert(dnsName)
	if err != nil {
		return certs, fmt.Errorf("failed to ensure certs: %v", err)
	}

	if err := writer.WriteCertsToDir(certPath, certs); err != nil {
		return certs, fmt.Errorf("failed to WriteCertsToDir: %v", err)
	}
	return certs, nil
}

// PatchCABundle patch the caBundle to MutatingWebhookConfiguration
func (c *CertificateBuilder) PatchCABundle(webhookName string, ca []byte) error {

	var m v1.MutatingWebhookConfiguration

	c.log.Info("start patch MutatingWebhookConfiguration caBundle", "name", webhookName)

	ctx := context.Background()

	if err := c.Get(ctx, client.ObjectKey{Name: webhookName}, &m); err != nil {
		c.log.Error(err, "fail to get mutatingWebHook", "name", webhookName)
		return err
	}

	current := m.DeepCopy()
	for i := range m.Webhooks {
		m.Webhooks[i].ClientConfig.CABundle = ca
	}

	if reflect.DeepEqual(m.Webhooks, current.Webhooks) {
		c.log.Info("no need to patch the MutatingWebhookConfiguration", "name", webhookName)
		return nil
	}

	if err := c.Patch(ctx, &m, client.MergeFrom(current)); err != nil {
		c.log.Error(err, "fail to patch CABundle to mutatingWebHook", "name", webhookName)
		return err
	}

	c.log.Info("finished patch MutatingWebhookConfiguration caBundle", "name", webhookName)

	return nil
}
