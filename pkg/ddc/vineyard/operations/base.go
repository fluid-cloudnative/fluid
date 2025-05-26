/*
Copyright 2024 The Fluid Authors.
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

package operations

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/validation"
)

type VineyardFileUtils struct {
	podNamePrefix string
	port          int32
	replicas      int32
	namespace     string
	log           logr.Logger
}

func NewVineyardFileUtils(podNamePrefix string, port int32, replicas int32, namespace string, log logr.Logger) VineyardFileUtils {
	return VineyardFileUtils{
		podNamePrefix: podNamePrefix,
		port:          port,
		replicas:      replicas,
		namespace:     namespace,
		log:           log,
	}
}

func validateResourceName(name string) error {
	if errs := validation.IsDNS1035Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid DNS-1035 label: %s", name)
	}
	return nil
}

// Get summary info of the Vineyard Engine
func (a VineyardFileUtils) ReportSummary() (summary []string, err error) {
	var resp *http.Response
	var body []byte
	if err := validateResourceName(a.podNamePrefix); err != nil {
		return nil, err
	}
	if err := validateResourceName(a.namespace); err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	for i := int32(0); i < a.replicas; i++ {
		u := url.URL{
			Scheme: "http",
			Host: net.JoinHostPort(fmt.Sprintf("%s-%d.%s.%s.svc", a.podNamePrefix, i,
				a.podNamePrefix, a.namespace), strconv.Itoa(int(a.port))),
			Path: "/metrics",
		}

		resp, err = client.Get(u.String())
		if err != nil {
			err = fmt.Errorf("failed to get metrics from %s, error: %v", u.String(), err)
			return summary, err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("failed to read response body from %s, error: %v", u.String(), err)
			return
		}

		summary = append(summary, string(body))
	}
	return summary, nil
}
