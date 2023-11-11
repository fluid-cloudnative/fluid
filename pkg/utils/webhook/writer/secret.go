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

package writer

import (
	"errors"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/webhook/generator"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// secretCertWriter provisions the certificate by reading and writing to the k8s secrets.
type secretCertWriter struct {
	*SecretCertWriterOptions

	// dnsName is the DNS name that the certificate is for.
	dnsName string
}

// SecretCertWriterOptions is options for constructing a secretCertWriter.
type SecretCertWriterOptions struct {
	// client talks to a kubernetes cluster for creating the secret.
	Client client.Client
	// certGenerator generates the certificates.
	CertGenerator generator.CertGenerator
	// secret points the secret that contains certificates that written by the CertWriter.
	Secret *types.NamespacedName
}

var _ CertWriter = &secretCertWriter{}

func (ops *SecretCertWriterOptions) setDefaults() {
	if ops.CertGenerator == nil {
		ops.CertGenerator = &generator.SelfSignedCertGenerator{}
	}
}

func (ops *SecretCertWriterOptions) validate() error {
	if ops.Client == nil {
		return errors.New("client must be set in SecretCertWriterOptions")
	}
	if ops.Secret == nil {
		return errors.New("secret must be set in SecretCertWriterOptions")
	}
	return nil
}

// NewSecretCertWriter constructs a CertWriter that persists the certificate in a k8s secret.
func NewSecretCertWriter(ops SecretCertWriterOptions) (CertWriter, error) {
	ops.setDefaults()
	err := ops.validate()
	if err != nil {
		return nil, err
	}
	return &secretCertWriter{
		SecretCertWriterOptions: &ops,
	}, nil
}

// EnsureCert provisions certificates for a webhookClientConfig by writing the certificates to a k8s secret.
func (s *secretCertWriter) EnsureCert(dnsName string) (*generator.Artifacts, bool, error) {
	// Create or refresh the certs based on clientConfig
	s.dnsName = dnsName
	return handleCommon(s.dnsName, s)
}

var _ certReadWriter = &secretCertWriter{}

func (s *secretCertWriter) buildSecret() (*corev1.Secret, *generator.Artifacts, error) {
	certs, err := s.CertGenerator.Generate(s.dnsName)
	if err != nil {
		return nil, nil, err
	}
	secret := certsToSecret(certs, *s.Secret)
	return secret, certs, err
}

func (s *secretCertWriter) write() (*generator.Artifacts, error) {
	secret, certs, err := s.buildSecret()
	if err != nil {
		return nil, err
	}
	err = kubeclient.CreateSecret(s.Client, secret)
	if apierrors.IsAlreadyExists(err) {
		return nil, err
	}
	return certs, err
}

func (s *secretCertWriter) overwrite(resourceVersion string) (
	*generator.Artifacts, error) {
	secret, certs, err := s.buildSecret()
	if err != nil {
		return nil, err
	}
	secret.ResourceVersion = resourceVersion
	err = kubeclient.UpdateSecret(s.Client, secret)
	if err != nil {
		log.Info("Cert writer update secret failed: %v", err)
		return certs, err
	}
	log.Info("Cert writer update secret %s resourceVersion from %s to %s",
		secret.Name, resourceVersion, secret.ResourceVersion)
	return certs, err
}

func (s *secretCertWriter) read() (*generator.Artifacts, error) {
	secret, err := kubeclient.GetSecret(s.Client, s.Secret.Name, s.Secret.Namespace)
	if apierrors.IsNotFound(err) {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	certs := secretToCerts(secret)
	if certs != nil && certs.CACert != nil && certs.CAKey != nil {
		// Store the CA for next usage.
		s.CertGenerator.SetCA(certs.CAKey, certs.CACert)
	}
	return certs, nil
}

func secretToCerts(secret *corev1.Secret) *generator.Artifacts {
	if secret.Data == nil {
		return &generator.Artifacts{ResourceVersion: secret.ResourceVersion}
	}
	return &generator.Artifacts{
		CAKey:           secret.Data[CAKeyName],
		CACert:          secret.Data[CACertName],
		Cert:            secret.Data[ServerCertName],
		Key:             secret.Data[ServerKeyName],
		ResourceVersion: secret.ResourceVersion,
	}
}

func certsToSecret(certs *generator.Artifacts, sec types.NamespacedName) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sec.Namespace,
			Name:      sec.Name,
		},
		Data: map[string][]byte{
			CAKeyName:       certs.CAKey,
			CACertName:      certs.CACert,
			ServerKeyName:   certs.Key,
			ServerKeyName2:  certs.Key,
			ServerCertName:  certs.Cert,
			ServerCertName2: certs.Cert,
		},
	}
}
