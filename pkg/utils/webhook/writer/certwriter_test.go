/*
Copyright 2021 The Fluid Authors.

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

package writer

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils/webhook/generator"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCertWriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CertWriter Suite")
}

var _ = Describe("CertWriter", func() {
	var (
		testDNSName = "test-service.test-namespace.svc"
	)

	Describe("handleCommon", func() {
		Context("when input validation fails", func() {
			It("should return error when dnsName is empty", func() {
				mockWriter := &mockCertReadWriter{}
				artifacts, changed, err := handleCommon("", mockWriter)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("dnsName should not be empty"))
				Expect(artifacts).To(BeNil())
				Expect(changed).To(BeFalse())
			})

			It("should return error when certReadWriter is nil", func() {
				artifacts, changed, err := handleCommon(testDNSName, nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("certReaderWriter should not be nil"))
				Expect(artifacts).To(BeNil())
				Expect(changed).To(BeFalse())
			})
		})

		Context("when certificates need to be created", func() {
			It("should create new certificates when they don't exist", func() {
				validCerts := generateValidTestCerts(testDNSName)
				mockWriter := &mockCertReadWriter{
					readErr:   apierrs.NewNotFound(schema.GroupResource{}, "test"),
					writeResp: validCerts,
				}

				artifacts, changed, err := handleCommon(testDNSName, mockWriter)

				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeTrue())
				Expect(artifacts).NotTo(BeNil())
				Expect(mockWriter.writeCalled).To(BeTrue())
			})

			It("should handle already exists error during write", func() {
				validCerts := generateValidTestCerts(testDNSName)
				mockWriter := &mockCertReadWriter{
					readErr:   apierrs.NewNotFound(schema.GroupResource{}, "test"),
					writeErr:  apierrs.NewAlreadyExists(schema.GroupResource{}, "test"),
					readResp2: validCerts,
					writeResp: validCerts,
				}

				artifacts, changed, err := handleCommon(testDNSName, mockWriter)

				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeTrue())
				Expect(artifacts).NotTo(BeNil())
				Expect(mockWriter.readCount).To(Equal(2))
			})
		})

		Context("when certificates exist but are invalid", func() {
			It("should regenerate certificates when they are invalid", func() {
				invalidCerts := &generator.Artifacts{
					Cert:   []byte("invalid-cert"),
					Key:    []byte("invalid-key"),
					CACert: []byte("invalid-ca"),
				}
				validCerts := generateValidTestCerts(testDNSName)

				mockWriter := &mockCertReadWriter{
					readResp:      invalidCerts,
					overwriteResp: validCerts,
				}

				artifacts, changed, err := handleCommon(testDNSName, mockWriter)

				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeTrue())
				Expect(artifacts).NotTo(BeNil())
				Expect(mockWriter.overwriteCalled).To(BeTrue())
			})

			It("should return error when overwrite fails", func() {
				invalidCerts := &generator.Artifacts{
					Cert:   []byte("invalid-cert"),
					Key:    []byte("invalid-key"),
					CACert: []byte("invalid-ca"),
				}

				mockWriter := &mockCertReadWriter{
					readResp:     invalidCerts,
					overwriteErr: apierrs.NewInternalError(errors.New("overwrite failed")),
				}

				artifacts, changed, err := handleCommon(testDNSName, mockWriter)

				Expect(err).To(HaveOccurred())
				Expect(changed).To(BeFalse())
				Expect(artifacts).To(BeNil())
			})
		})

		Context("when certificates exist and are valid", func() {
			It("should return existing certificates without changes", func() {
				validCerts := generateValidTestCerts(testDNSName)
				mockWriter := &mockCertReadWriter{
					readResp: validCerts,
				}

				artifacts, changed, err := handleCommon(testDNSName, mockWriter)

				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeFalse())
				Expect(artifacts).NotTo(BeNil())
				Expect(mockWriter.overwriteCalled).To(BeFalse())
			})
		})
	})

	Describe("createIfNotExists", func() {
		It("should return existing certificates when they exist", func() {
			existingCerts := generateValidTestCerts(testDNSName)
			mockWriter := &mockCertReadWriter{
				readResp: existingCerts,
			}

			artifacts, changed, err := createIfNotExists(mockWriter)

			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeFalse())
			Expect(artifacts).To(Equal(existingCerts))
			Expect(mockWriter.writeCalled).To(BeFalse())
		})

		It("should create certificates when they don't exist", func() {
			newCerts := generateValidTestCerts(testDNSName)
			mockWriter := &mockCertReadWriter{
				readErr:   apierrs.NewNotFound(schema.GroupResource{}, "test"),
				writeResp: newCerts,
			}

			artifacts, changed, err := createIfNotExists(mockWriter)

			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(artifacts).To(Equal(newCerts))
			Expect(mockWriter.writeCalled).To(BeTrue())
		})

		It("should handle race condition with AlreadyExists error", func() {
			raceCerts := generateValidTestCerts(testDNSName)
			mockWriter := &mockCertReadWriter{
				readErr:   apierrs.NewNotFound(schema.GroupResource{}, "test"),
				writeErr:  apierrs.NewAlreadyExists(schema.GroupResource{}, "test"),
				readResp2: raceCerts,
			}

			artifacts, changed, err := createIfNotExists(mockWriter)

			Expect(err).NotTo(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(artifacts).To(Equal(raceCerts))
			Expect(mockWriter.readCount).To(Equal(2))
		})

		It("should propagate read errors other than NotFound", func() {
			mockWriter := &mockCertReadWriter{
				readErr: apierrs.NewInternalError(errors.New("read failed")),
			}

			artifacts, changed, err := createIfNotExists(mockWriter)

			Expect(err).To(HaveOccurred())
			Expect(changed).To(BeFalse())
			Expect(artifacts).To(BeNil())
		})

		It("should propagate write errors other than AlreadyExists", func() {
			mockWriter := &mockCertReadWriter{
				readErr:  apierrs.NewNotFound(schema.GroupResource{}, "test"),
				writeErr: apierrs.NewInternalError(errors.New("write failed")),
			}

			artifacts, changed, err := createIfNotExists(mockWriter)

			Expect(err).To(HaveOccurred())
			Expect(changed).To(BeTrue())
			Expect(artifacts).To(BeNil())
		})
	})

	Describe("validCert", func() {
		Context("when certificates are nil or incomplete", func() {
			It("should return false when artifacts is nil", func() {
				valid := validCert(nil, testDNSName)
				Expect(valid).To(BeFalse())
			})

			It("should return false when Cert is nil", func() {
				artifacts := &generator.Artifacts{
					Key:    []byte("key"),
					CACert: []byte("cacert"),
				}
				valid := validCert(artifacts, testDNSName)
				Expect(valid).To(BeFalse())
			})

			It("should return false when Key is nil", func() {
				artifacts := &generator.Artifacts{
					Cert:   []byte("cert"),
					CACert: []byte("cacert"),
				}
				valid := validCert(artifacts, testDNSName)
				Expect(valid).To(BeFalse())
			})

			It("should return false when CACert is nil", func() {
				artifacts := &generator.Artifacts{
					Cert: []byte("cert"),
					Key:  []byte("key"),
				}
				valid := validCert(artifacts, testDNSName)
				Expect(valid).To(BeFalse())
			})
		})

		Context("when certificates are malformed", func() {
			It("should return false when cert and key don't form a valid pair", func() {
				artifacts := &generator.Artifacts{
					Cert:   []byte("invalid-cert"),
					Key:    []byte("invalid-key"),
					CACert: []byte("invalid-ca"),
				}
				valid := validCert(artifacts, testDNSName)
				Expect(valid).To(BeFalse())
			})

			It("should return false when CA cert is invalid PEM", func() {
				// Generate valid key pair but invalid CA
				key, cert := generateValidKeyPair(testDNSName)
				artifacts := &generator.Artifacts{
					Cert:   cert,
					Key:    key,
					CACert: []byte("invalid-ca-pem"),
				}
				valid := validCert(artifacts, testDNSName)
				Expect(valid).To(BeFalse())
			})

			It("should return false when cert PEM cannot be decoded", func() {
				validCerts := generateValidTestCerts(testDNSName)
				artifacts := &generator.Artifacts{
					Cert:   []byte("invalid-pem-format"),
					Key:    validCerts.Key,
					CACert: validCerts.CACert,
				}
				valid := validCert(artifacts, testDNSName)
				Expect(valid).To(BeFalse())
			})
		})

		Context("when certificates are valid", func() {
			It("should return true for valid certificates", func() {
				validCerts := generateValidTestCerts(testDNSName)
				valid := validCert(validCerts, testDNSName)
				Expect(valid).To(BeTrue())
			})
		})

		Context("when certificates are expiring soon", func() {
			It("should return false for certificates expiring within 6 months", func() {
				// Generate cert that expires in 3 months
				expiringSoonCerts := generateCertsWithExpiry(testDNSName, 90*24*time.Hour)
				valid := validCert(expiringSoonCerts, testDNSName)
				Expect(valid).To(BeFalse())
			})
		})

		Context("when DNS name doesn't match", func() {
			It("should return false when DNS name doesn't match certificate", func() {
				validCerts := generateValidTestCerts("different-service.test-namespace.svc")
				valid := validCert(validCerts, testDNSName)
				Expect(valid).To(BeFalse())
			})
		})
	})
})

// mockCertReadWriter is a mock implementation of certReadWriter for testing
type mockCertReadWriter struct {
	readResp        *generator.Artifacts
	readResp2       *generator.Artifacts
	readErr         error
	writeResp       *generator.Artifacts
	writeErr        error
	overwriteResp   *generator.Artifacts
	overwriteErr    error
	writeCalled     bool
	overwriteCalled bool
	readCount       int
}

func (m *mockCertReadWriter) read() (*generator.Artifacts, error) {
	m.readCount++
	if m.readCount == 1 {
		return m.readResp, m.readErr
	}
	return m.readResp2, nil
}

func (m *mockCertReadWriter) write() (*generator.Artifacts, error) {
	m.writeCalled = true
	return m.writeResp, m.writeErr
}

func (m *mockCertReadWriter) overwrite(resourceVersion string) (*generator.Artifacts, error) {
	m.overwriteCalled = true
	return m.overwriteResp, m.overwriteErr
}

// Helper functions to generate test certificates

func generateValidTestCerts(dnsName string) *generator.Artifacts {
	// Generate CA
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

	// Generate server cert
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: dnsName,
		},
		DNSNames:    []string{dnsName},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return &generator.Artifacts{
		Cert:   serverCertPEM,
		Key:    serverKeyPEM,
		CACert: caCertPEM,
	}
}

func generateCertsWithExpiry(dnsName string, validDuration time.Duration) *generator.Artifacts {
	// Generate CA
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(validDuration),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})

	// Generate server cert with same expiry
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: dnsName,
		},
		DNSNames:    []string{dnsName},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(validDuration),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		panic(err)
	}

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})

	return &generator.Artifacts{
		Cert:   serverCertPEM,
		Key:    serverKeyPEM,
		CACert: caCertPEM,
	}
}

func generateValidKeyPair(dnsName string) ([]byte, []byte) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: dnsName,
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return keyPEM, certPEM
}
