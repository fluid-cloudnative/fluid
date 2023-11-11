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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/webhook/generator"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// CAKeyName is the name of the CA private key
	CAKeyName = "ca-key.pem"
	// CACertName is the name of the CA certificate
	CACertName = "ca-cert.pem"
	// ServerKeyName is the name of the server private key
	ServerKeyName  = "key.pem"
	ServerKeyName2 = "tls.key"
	// ServerCertName is the name of the serving certificate
	ServerCertName  = "cert.pem"
	ServerCertName2 = "tls.crt"
)

var log = ctrl.Log.WithName("webhookWriter")

// CertWriter provides method to handle webhooks.
type CertWriter interface {
	// EnsureCert provisions the cert for the webhookClientConfig.
	EnsureCert(dnsName string) (*generator.Artifacts, bool, error)
}

// handleCommon ensures the given webhook has a proper certificate.
// It uses the given certReadWriter to read and (or) write the certificate.
func handleCommon(dnsName string, ch certReadWriter) (*generator.Artifacts, bool, error) {
	if len(dnsName) == 0 {
		return nil, false, errors.New("dnsName should not be empty")
	}
	if ch == nil {
		return nil, false, errors.New("certReaderWriter should not be nil")
	}

	certs, changed, err := createIfNotExists(ch)
	if err != nil {
		return nil, changed, err
	}

	// Recreate the cert if it's invalid.
	valid := validCert(certs, dnsName)
	if !valid {
		log.Info("cert is invalid or expiring, regenerating a new one")
		certs, err = ch.overwrite(certs.ResourceVersion)
		if err != nil {
			return nil, false, err
		}
		changed = true
	}
	return certs, changed, nil
}

func createIfNotExists(ch certReadWriter) (*generator.Artifacts, bool, error) {
	// Try to read first
	certs, err := ch.read()
	if apierrs.IsNotFound(err) {
		// Create if not exists
		certs, err = ch.write()
		switch {
		// This may happen if there is another racer.
		case apierrs.IsAlreadyExists(err):
			certs, err = ch.read()
			return certs, true, err
		default:
			return certs, true, err
		}
	}
	return certs, false, err
}

// certReadWriter provides methods for reading and writing certificates.
type certReadWriter interface {
	// read a webhook name and returns the certs for it.
	read() (*generator.Artifacts, error)
	// write the certs and return the certs it wrote.
	write() (*generator.Artifacts, error)
	// overwrite the existing certs and return the certs it wrote.
	overwrite(resourceVersion string) (*generator.Artifacts, error)
}

func validCert(certs *generator.Artifacts, dnsName string) bool {
	if certs == nil || certs.Cert == nil || certs.Key == nil || certs.CACert == nil {
		return false
	}

	// Verify key and cert are valid pair
	_, err := tls.X509KeyPair(certs.Cert, certs.Key)
	if err != nil {
		return false
	}

	// Verify cert is good for desired DNS name and signed by CA and will be valid for desired period of time.
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certs.CACert) {
		return false
	}
	block, _ := pem.Decode(certs.Cert)
	if block == nil {
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	ops := x509.VerifyOptions{
		DNSName:     dnsName,
		Roots:       pool,
		CurrentTime: time.Now().AddDate(0, 6, 0),
	}
	_, err = cert.Verify(ops)
	return err == nil
}
