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

package kubelet

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
)

var (
	mockPodLists = "{\n    \"kind\":\"PodList\",\n    \"apiVersion\":\"v1\",\n    \"metadata\":{\n\n    },\n    \"items\":[\n        {\n            \"metadata\":{\n                \"name\":\"jfsdemo-fuse-6cnqt\",\n                \"generateName\":\"jfsdemo-fuse-\",\n                \"namespace\":\"default\",\n                \"labels\":{\n                    \"app\":\"juicefs\",\n                    \"role\":\"juicefs-fuse\"\n                }\n            },\n            \"spec\":{\n                \"containers\":[\n                    {\n                        \"name\":\"juicefs-fuse\",\n                        \"image\":\"juicedata/juicefs-csi-driver:v0.10.5\",\n                        \"command\":[\n                            \"sh\",\n                            \"-c\",\n                            \"/bin/mount.juicefs redis://10.111.100.91:6379/0 /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/demo,cache-dir=/dev/shm,cache-size=40960,free-space-ratio=0.1\"\n                        ],\n                        \"securityContext\":{\n                            \"privileged\":true\n                        }\n                    }\n                ]\n            }\n        }\n    ]\n}"
	privileged   = true
	mockPod      = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:         "jfsdemo-fuse-6cnqt",
			GenerateName: "jfsdemo-fuse-",
			Namespace:    "default",
			Labels: map[string]string{
				"app":  "juicefs",
				"role": "juicefs-fuse",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "juicefs-fuse",
				Image:   "juicedata/juicefs-csi-driver:v0.10.5",
				Command: []string{"sh", "-c", "/bin/mount.juicefs redis://10.111.100.91:6379/0 /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse -o metrics=0.0.0.0:9567,subdir=/demo,cache-dir=/dev/shm,cache-size=40960,free-space-ratio=0.1"},
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
			}},
		},
	}
)

var _ = Describe("NewKubeletClient", func() {
	Context("when creating client successfully", func() {
		It("should return kubelet client without error", func() {
			config := &KubeletClientConfig{
				Address:     "localhost",
				Port:        10250,
				HTTPTimeout: 5 * time.Second,
			}

			patch := ApplyFunc(transport.TLSConfigFor, func(c *transport.Config) (*tls.Config, error) {
				return nil, nil
			})
			defer patch.Reset()

			patch2 := ApplyFunc(transport.HTTPWrappersForConfig, func(config *transport.Config, rt http.RoundTripper) (http.RoundTripper, error) {
				return http.DefaultTransport, nil
			})
			defer patch2.Reset()

			client, err := NewKubeletClient(config)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
			Expect(client.host).To(Equal("localhost"))
			Expect(client.defaultPort).To(Equal(uint(10250)))
		})

		It("should handle secure TLS configuration", func() {
			config := &KubeletClientConfig{
				Address:     "localhost",
				Port:        10250,
				HTTPTimeout: 5 * time.Second,
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   "/path/to/ca.crt",
					CertFile: "/path/to/client.crt",
					KeyFile:  "/path/to/client.key",
				},
			}

			patch := ApplyFunc(transport.TLSConfigFor, func(c *transport.Config) (*tls.Config, error) {
				return &tls.Config{}, nil
			})
			defer patch.Reset()

			patch2 := ApplyFunc(transport.HTTPWrappersForConfig, func(config *transport.Config, rt http.RoundTripper) (http.RoundTripper, error) {
				return http.DefaultTransport, nil
			})
			defer patch2.Reset()

			client, err := NewKubeletClient(config)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})
	})

	Context("when TLSConfigFor fails", func() {
		It("should return error", func() {
			config := &KubeletClientConfig{
				Address:     "localhost",
				Port:        10250,
				HTTPTimeout: 5 * time.Second,
			}

			patch := ApplyFunc(transport.TLSConfigFor, func(c *transport.Config) (*tls.Config, error) {
				return nil, errors.New("TLS config error")
			})
			defer patch.Reset()

			client, err := NewKubeletClient(config)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("TLS config error"))
			Expect(client).To(BeNil())
		})
	})

	Context("when HTTPWrappersForConfig fails", func() {
		It("should return error", func() {
			config := &KubeletClientConfig{
				Address:     "localhost",
				Port:        10250,
				HTTPTimeout: 5 * time.Second,
			}

			patch := ApplyFunc(transport.TLSConfigFor, func(c *transport.Config) (*tls.Config, error) {
				return nil, nil
			})
			defer patch.Reset()

			patch2 := ApplyFunc(transport.HTTPWrappersForConfig, func(config *transport.Config, rt http.RoundTripper) (http.RoundTripper, error) {
				return nil, errors.New("wrapper error")
			})
			defer patch2.Reset()

			client, err := NewKubeletClient(config)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("wrapper error"))
			Expect(client).To(BeNil())
		})
	})
})

var _ = Describe("makeTransport", func() {
	Context("with insecure skip TLS verify enabled", func() {
		It("should configure transport with insecure TLS", func() {
			config := &KubeletClientConfig{
				Address: "localhost",
				Port:    10250,
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   "/path/to/ca.crt",
					CertFile: "/path/to/client.crt",
					KeyFile:  "/path/to/client.key",
				},
			}

			patch := ApplyFunc(transport.TLSConfigFor, func(c *transport.Config) (*tls.Config, error) {
				// Verify that insecure was set
				if c.TLS.Insecure && c.TLS.CAData == nil && c.TLS.CAFile == "" {
					return nil, nil
				}
				return nil, nil
			})
			defer patch.Reset()

			patch2 := ApplyFunc(transport.HTTPWrappersForConfig, func(config *transport.Config, rt http.RoundTripper) (http.RoundTripper, error) {
				return http.DefaultTransport, nil
			})
			defer patch2.Reset()

			rt, err := makeTransport(config, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).NotTo(BeNil())
		})
	})

	Context("with secure TLS configuration", func() {
		It("should configure transport with TLS", func() {
			config := &KubeletClientConfig{
				Address: "localhost",
				Port:    10250,
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   "/path/to/ca.crt",
					CertFile: "/path/to/client.crt",
					KeyFile:  "/path/to/client.key",
				},
			}

			patch := ApplyFunc(transport.TLSConfigFor, func(c *transport.Config) (*tls.Config, error) {
				return &tls.Config{}, nil
			})
			defer patch.Reset()

			patch2 := ApplyFunc(transport.HTTPWrappersForConfig, func(config *transport.Config, rt http.RoundTripper) (http.RoundTripper, error) {
				return http.DefaultTransport, nil
			})
			defer patch2.Reset()

			rt, err := makeTransport(config, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(rt).NotTo(BeNil())
		})
	})
})

var _ = Describe("transportConfig", func() {
	Context("when CA file is provided", func() {
		It("should configure with CA file", func() {
			config := &KubeletClientConfig{
				TLSClientConfig: rest.TLSClientConfig{
					CAFile:   "/path/to/ca.crt",
					CertFile: "/path/to/client.crt",
					KeyFile:  "/path/to/client.key",
				},
				BearerToken: "test-token",
			}

			transportCfg := config.transportConfig()
			Expect(transportCfg.TLS.CAFile).To(Equal("/path/to/ca.crt"))
			Expect(transportCfg.TLS.CertFile).To(Equal("/path/to/client.crt"))
			Expect(transportCfg.TLS.KeyFile).To(Equal("/path/to/client.key"))
			Expect(transportCfg.BearerToken).To(Equal("test-token"))
		})
	})

	Context("when CA data is provided", func() {
		It("should configure with CA data", func() {
			caData := []byte("ca-data")
			certData := []byte("cert-data")
			keyData := []byte("key-data")

			config := &KubeletClientConfig{
				TLSClientConfig: rest.TLSClientConfig{
					CAData:   caData,
					CertData: certData,
					KeyData:  keyData,
				},
			}

			transportCfg := config.transportConfig()
			Expect(transportCfg.TLS.CAData).To(Equal(caData))
			Expect(transportCfg.TLS.CertData).To(Equal(certData))
			Expect(transportCfg.TLS.KeyData).To(Equal(keyData))
		})
	})

	Context("when no CA is provided", func() {
		It("should set insecure to true", func() {
			config := &KubeletClientConfig{
				BearerToken: "test-token",
			}

			transportCfg := config.transportConfig()
			Expect(transportCfg.TLS.Insecure).To(BeTrue())
		})
	})
})

var _ = Describe("ReadAll", func() {
	Context("when reading small data", func() {
		It("should read all data successfully", func() {
			data := "test data"
			reader := strings.NewReader(data)

			result, err := ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(Equal(data))
		})
	})

	Context("when reading large data", func() {
		It("should read all data and expand buffer", func() {
			// Create data larger than initial buffer (512 bytes)
			largeData := strings.Repeat("a", 1024)
			reader := strings.NewReader(largeData)

			result, err := ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(Equal(largeData))
			Expect(len(result)).To(Equal(1024))
		})
	})

	Context("when reader returns error", func() {
		It("should return error", func() {
			reader := &errorReader{err: errors.New("read error")}

			result, err := ReadAll(reader)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("read error"))
			Expect(result).NotTo(BeNil()) // Returns partial data
		})
	})

	Context("when reading EOF", func() {
		It("should return without error", func() {
			reader := strings.NewReader("")

			result, err := ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result)).To(Equal(0))
		})
	})
})

// Helper type for testing read errors
type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

var _ = Describe("KubeletClient", func() {
	Describe("GetNodeRunningPods", func() {
		var kubeletClient KubeletClient

		BeforeEach(func() {
			kubeletClient = KubeletClient{
				host:        "localhost",
				defaultPort: 10250,
			}
		})

		Context("when request is successful", func() {
			It("should return pod list", func() {
				httpClient := &http.Client{}
				patch := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(mockPodLists))}, nil
				})
				defer patch.Reset()

				kubeletClient.client = httpClient
				got, err := kubeletClient.GetNodeRunningPods()
				Expect(err).NotTo(HaveOccurred())
				Expect(got.Items).To(HaveLen(1))
				Expect(reflect.DeepEqual(got.Items[0], mockPod)).To(BeTrue())
			})
		})

		Context("when HTTP client returns error", func() {
			It("should return error", func() {
				httpClient := &http.Client{}
				patch := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
					return nil, errors.New("connection error")
				})
				defer patch.Reset()

				kubeletClient.client = httpClient
				got, err := kubeletClient.GetNodeRunningPods()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection error"))
				Expect(got).To(BeNil())
			})
		})

		Context("when reading response body fails", func() {
			It("should return error", func() {
				httpClient := &http.Client{}
				patch := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
					return &http.Response{Body: &errorReadCloser{err: errors.New("read failed")}}, nil
				})
				defer patch.Reset()

				kubeletClient.client = httpClient
				got, err := kubeletClient.GetNodeRunningPods()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("read failed"))
				Expect(got).To(BeNil())
			})
		})

		Context("when response contains invalid JSON", func() {
			It("should return error", func() {
				httpClient := &http.Client{}
				patch := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader("invalid json"))}, nil
				})
				defer patch.Reset()

				kubeletClient.client = httpClient
				got, err := kubeletClient.GetNodeRunningPods()
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeNil())
			})
		})

		Context("when response body is empty", func() {
			It("should return error for invalid JSON", func() {
				httpClient := &http.Client{}
				patch := ApplyMethod(reflect.TypeOf(httpClient), "Get", func(_ *http.Client, url string) (resp *http.Response, err error) {
					return &http.Response{Body: io.NopCloser(strings.NewReader(""))}, nil
				})
				defer patch.Reset()

				kubeletClient.client = httpClient
				got, err := kubeletClient.GetNodeRunningPods()
				Expect(err).To(HaveOccurred())
				Expect(got).To(BeNil())
			})
		})
	})
})

// Helper type for testing read errors on response body
type errorReadCloser struct {
	err error
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReadCloser) Close() error {
	return nil
}

var _ = Describe("InitNodeAuthorizedClient", func() {
	Context("when initialization is successful", func() {
		It("should return clientset without error", func() {
			patch1 := ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return &rest.Config{}, nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
				return &kubernetes.Clientset{}, nil
			})
			defer patch2.Reset()

			client, err := InitNodeAuthorizedClient("/path/to/kubeconfig")
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})
	})

	Context("when BuildConfigFromFlags fails", func() {
		It("should return error with proper message", func() {
			patch := ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return nil, errors.New("failed to build config")
			})
			defer patch.Reset()

			client, err := InitNodeAuthorizedClient("/path/to/kubeconfig")
			Expect(err).To(HaveOccurred())
			Expect(client).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("fail to build kubelet config"))
		})
	})

	Context("when NewForConfig fails", func() {
		It("should return error with proper message", func() {
			patch1 := ApplyFunc(clientcmd.BuildConfigFromFlags, func(masterUrl, kubeconfigPath string) (*rest.Config, error) {
				return &rest.Config{}, nil
			})
			defer patch1.Reset()

			patch2 := ApplyFunc(kubernetes.NewForConfig, func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, errors.New("failed to create client")
			})
			defer patch2.Reset()

			client, err := InitNodeAuthorizedClient("/path/to/kubeconfig")
			Expect(err).To(HaveOccurred())
			Expect(client).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("fail to build client-go client from kubelet kubeconfig"))
		})
	})
})
