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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	v1 "k8s.io/api/core/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

// KubeletClientConfig defines config parameters for the kubelet client
type KubeletClientConfig struct {
	// Address specifies the kubelet address
	Address string

	// Port specifies the default port - used if no information about Kubelet port can be found in Node.NodeStatus.DaemonEndpoints.
	Port uint

	// TLSClientConfig contains settings to enable transport layer security
	restclient.TLSClientConfig

	// Server requires Bearer authentication
	BearerToken string

	// HTTPTimeout is used by the client to timeout http requests to Kubelet.
	HTTPTimeout time.Duration
}

type KubeletClient struct {
	defaultPort uint
	host        string
	client      *http.Client
}

func NewKubeletClient(config *KubeletClientConfig) (*KubeletClient, error) {
	trans, err := makeTransport(config, config.Insecure)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: trans,
		Timeout:   config.HTTPTimeout,
	}
	return &KubeletClient{
		host:        config.Address,
		defaultPort: config.Port,
		client:      client,
	}, nil
}

// makeTransport creates a RoundTripper for HTTP Transport.
func makeTransport(config *KubeletClientConfig, insecureSkipTLSVerify bool) (http.RoundTripper, error) {
	// do the insecureSkipTLSVerify on the pre-transport *before* we go get a potentially cached connection.
	// transportConfig always produces a new struct pointer.
	preTLSConfig := config.transportConfig()
	if insecureSkipTLSVerify && preTLSConfig != nil {
		preTLSConfig.TLS.Insecure = true
		preTLSConfig.TLS.CAData = nil
		preTLSConfig.TLS.CAFile = ""
	}

	tlsConfig, err := transport.TLSConfigFor(preTLSConfig)
	if err != nil {
		return nil, err
	}

	rt := http.DefaultTransport
	if tlsConfig != nil {
		// If SSH Tunnel is turned on
		rt = utilnet.SetOldTransportDefaults(&http.Transport{
			TLSClientConfig: tlsConfig,
		})
	}

	return transport.HTTPWrappersForConfig(config.transportConfig(), rt)
}

// transportConfig converts a client config to an appropriate transport config.
func (c *KubeletClientConfig) transportConfig() *transport.Config {
	cfg := &transport.Config{
		TLS: transport.TLSConfig{
			CAFile:   c.CAFile,
			CAData:   c.CAData,
			CertFile: c.CertFile,
			CertData: c.CertData,
			KeyFile:  c.KeyFile,
			KeyData:  c.KeyData,
		},
		BearerToken: c.BearerToken,
	}
	if !cfg.HasCA() {
		cfg.TLS.Insecure = true
	}
	return cfg
}

func ReadAll(r io.Reader) ([]byte, error) {
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}
	}
}

func (k *KubeletClient) GetNodeRunningPods() (*v1.PodList, error) {
	resp, err := k.client.Get(fmt.Sprintf("https://%v:%d/pods/", k.host, k.defaultPort))
	if err != nil {
		return nil, err
	}

	body, err := ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	podList := &v1.PodList{}
	if err = json.Unmarshal(body, &podList); err != nil {
		return nil, err
	}
	return podList, err
}
