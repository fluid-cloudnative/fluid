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
	"net/http"
	"regexp"
	"time"

	"github.com/go-logr/logr"
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
	validNameRegex := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	matched, err := regexp.MatchString(validNameRegex, name)
	if err != nil {
		return fmt.Errorf("failed to validate name: %v", err)
	}
	if !matched {
		return fmt.Errorf("invalid Kubernetes naming: %s", name)
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
		url := fmt.Sprintf("http://%s-%d.%s-svc.%s.svc.cluster.local:%d/metrics", a.podNamePrefix, i, a.podNamePrefix, a.namespace, a.port)

		resp, err = client.Get(url)
		if err != nil {
			err = fmt.Errorf("failed to get metrics from %s, error: %v", url, err)
			return summary, err
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("failed to read response body from %s, error: %v", url, err)
			return
		}

		summary = append(summary, string(body))
	}
	return summary, nil
}
