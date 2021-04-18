package webhook

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"golang.org/x/net/context"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	admissionreg "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

const (
	defaultClusterDomain = "cluster.local"
)

var setupLog = ctrl.Log.WithName("certificate")

func getClusterDomain(service string) string {
	clusterName, err := net.LookupCNAME(service)
	if err != nil {
		return defaultClusterDomain
	}

	domain := strings.TrimPrefix(clusterName, service)
	domain = strings.Trim(domain, ".")
	if len(domain) == 0 {
		return defaultClusterDomain
	}

	return domain
}

func getSubjectAlternativeNames(service string, namespace string) []string {
	serviceNamespace := strings.Join([]string{service, namespace}, ".")
	serviceNamespaceSvc := strings.Join([]string{serviceNamespace, "svc"}, ".")
	return []string{serviceNamespace, serviceNamespaceSvc,
		strings.Join([]string{serviceNamespaceSvc, getClusterDomain(serviceNamespaceSvc)}, ".")}
}

type CertificateHandler struct {
	sync.Mutex
	tslCert             *tls.Certificate
	cert                []byte
	key                 []byte
	ca                  []byte
	serviceName         string
	namespace           string
	notAfter            time.Time
	certPath            string
	mutatingWebHookName string
}

func newCertificateHandler(serviceName, namespace, certPath, mutatingWebHookName string) *CertificateHandler {
	certHandler := &CertificateHandler{
		serviceName:         serviceName,
		namespace:           namespace,
		certPath:            certPath,
		mutatingWebHookName: mutatingWebHookName,
	}
	return certHandler
}

func UpdateCertificate(deployment, namespace, certPath, webHookName string) error {
	setupLog.Info("start to new a certificate")
	certHandler := newCertificateHandler(deployment, namespace, certPath, webHookName)
	err := certHandler.loadCertificate()
	if err != nil {
		setupLog.Error(err, "fail to load certificate")
		return err
	}

	// Update ValidatingWebHookConfiguration caBundle
	err = certHandler.updateCaBundle()
	if err != nil {
		setupLog.Error(err, "fail to update CaBundle")
		return err
	}

	// Watch certificate in background
	go certHandler.renewCertificate()
	return nil
}

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

func (handler *CertificateHandler) loadCertificate() error {
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

func (handler *CertificateHandler) renewCertificate() {
	waitingTime := time.Until(handler.notAfter) - 24*time.Hour
	for {
		time.Sleep(waitingTime)

		err := handler.loadCertificate()
		if err != nil {
			waitingTime = time.Hour
			continue
		}

		err = handler.updateCaBundle()
		if err != nil {
			waitingTime = time.Hour
			continue
		}
		waitingTime = time.Until(handler.notAfter) - 24*time.Hour
	}
}

func (handler *CertificateHandler) updateCaBundle() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		//log.Message(log.Error, err.Error())
		return err
	}
	admissionClient := admissionreg.NewForConfigOrDie(config)
	mutating := admissionClient.MutatingWebhookConfigurations()
	ctx := context.Background()
	obj, err := mutating.Get(ctx, handler.mutatingWebHookName, metav1.GetOptions{})
	if err != nil {
		setupLog.Error(err, "fail to get mutatingWebHookConfig")
		return err
	}

	for ind := range obj.Webhooks {
		obj.Webhooks[ind].ClientConfig.CABundle = handler.ca
	}
	_, err = mutating.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		setupLog.Error(err, "fail to update mutatingWebHookConfig")
		return err
	}
	return nil
}
