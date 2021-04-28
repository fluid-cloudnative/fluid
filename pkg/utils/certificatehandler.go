/*
Copyright 2018 The Kubernetes Authors.

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

package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"sync"
	"time"
)

const (
	defaultClusterDomain = "cluster.local"
)

var setupLog = ctrl.Log.WithName("certificate")

type CertificateHandler struct {
	sync.Mutex
	tslCert     *tls.Certificate
	cert        []byte
	key         []byte
	ca          []byte
	serviceName string
	namespace   string
	notAfter    time.Time
	certPath    string
}

func NewCertificateHandler(serviceName, namespace, certPath string) *CertificateHandler {
	certHandler := &CertificateHandler{
		serviceName: serviceName,
		namespace:   namespace,
		certPath:    certPath,
	}
	return certHandler
}

func (handler *CertificateHandler) GetCaCert() []byte {
	return handler.ca
}

func (handler *CertificateHandler) GetNotAfter() time.Time {
	return handler.notAfter
}

// LoadCertificate generate and store the Certificate
func (handler *CertificateHandler) LoadCertificate() error {
	handler.Lock()
	defer handler.Unlock()

	err := handler.generateCertificate()
	if err != nil {
		setupLog.Error(err, "fail to generate certificate")
		return err
	}
	err = handler.storeCertificate()
	if err != nil {
		setupLog.Error(err, "fail to storeCertificate")
		return err
	}

	tmpCert, err := tls.X509KeyPair(handler.cert, handler.key)
	if err != nil {
		setupLog.Error(err, "fail to X509KeyPair")
		return err
	}
	handler.tslCert = &tmpCert
	return nil
}

// generateCertificate generate the cert and key file and so on
func (handler *CertificateHandler) generateCertificate() error {
	now := time.Now()
	notBefore := now.Add(-time.Hour)
	notAfter := now.Add(time.Hour * 24 * 365)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: handler.serviceName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return err
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName: handler.serviceName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              getSubjectAlternativeNames(handler.serviceName, handler.namespace),
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	// Server cert PEM
	// Add server cert
	serverCertPEM := bytes.Buffer{}
	err = pem.Encode(&serverCertPEM, &pem.Block{Type: certutil.CertificateBlockType, Bytes: serverCertDER})
	if err != nil {
		setupLog.Error(err, "fail to encode server cert pem")
		return err
	}

	// Add ca cert
	err = pem.Encode(&serverCertPEM, &pem.Block{Type: certutil.CertificateBlockType, Bytes: caCertDER})
	if err != nil {
		setupLog.Error(err, "fail to encode ca cert")
		return err
	}

	// Server key PEM
	serverKeyPEM := bytes.Buffer{}
	err = pem.Encode(&serverKeyPEM, &pem.Block{Type: keyutil.RSAPrivateKeyBlockType, Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	if err != nil {
		setupLog.Error(err, "fail to encode server key pem")
		return err
	}

	// CA cert PEM
	caCertPEM := bytes.Buffer{}
	err = pem.Encode(&caCertPEM, &pem.Block{Type: certutil.CertificateBlockType, Bytes: caCertDER})
	if err != nil {
		setupLog.Error(err, "fail to encode ca cert pem")
		return err
	}

	handler.cert = serverCertPEM.Bytes()
	handler.key = serverKeyPEM.Bytes()
	handler.ca = caCertPEM.Bytes()
	handler.notAfter = notAfter
	return nil
}

// storeCertificate store the cert and key file in certPath
func (handler *CertificateHandler) storeCertificate() error {
	_, err := os.Stat(handler.certPath)
	if os.IsNotExist(err) {
		os.Mkdir(handler.certPath, 0644)
	}

	certPath := filepath.Join(handler.certPath, "tls.crt")
	err = ioutil.WriteFile(certPath, handler.cert, 0644)
	if err != nil {
		setupLog.Error(err, "cannot write cert", "certPath", certPath)
		return err
	}
	keyPath := filepath.Join(handler.certPath, "tls.key")
	err = ioutil.WriteFile(keyPath, handler.key, 0644)
	if err != nil {
		setupLog.Error(err, "cannot write key", "keyPath", keyPath)
		return err
	}
	return nil
}

func getSubjectAlternativeNames(service string, namespace string) []string {
	serviceNamespace := strings.Join([]string{service, namespace}, ".")
	serviceNamespaceSvc := strings.Join([]string{serviceNamespace, "svc"}, ".")
	return []string{serviceNamespace, serviceNamespaceSvc,
		strings.Join([]string{serviceNamespaceSvc, getClusterDomain(serviceNamespaceSvc)}, ".")}
}

// getClusterDomain get the ClusterDomain from serviceName
func getClusterDomain(service string) string {
	clusterName, err := net.LookupCNAME(service)
	if err != nil {
		setupLog.Error(err, "fail to lookup CNAME, will use default ClusterDomain", "service", service)
		return defaultClusterDomain
	}

	domain := strings.TrimPrefix(clusterName, service)
	domain = strings.Trim(domain, ".")
	if len(domain) == 0 {
		setupLog.Info("len of domain is 0, will use default ClusterDomain", "service", service)
		return defaultClusterDomain
	}

	return domain
}
