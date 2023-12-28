/*
Copyright 2023 The Fluid Authors.

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
	"io/ioutil"
	"net/http"

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

// Get summary info of the Vineyard Engine
func (a VineyardFileUtils) ReportSummary() (summary []string, err error) {
	var resp *http.Response
	var body []byte
	for i := int32(0); i < a.replicas; i++ {
		url := fmt.Sprintf("http://%s-%d.%s-svc.%s.svc.cluster.local:%d/metrics", a.podNamePrefix, i, a.podNamePrefix, a.namespace, a.port)

		resp, err = http.Get(url)
		if err != nil {
			err = fmt.Errorf("failed to get metrics from %s, error: %v", url, err)
			return summary, err
		}
		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("failed to read response body from %s, error: %v", url, err)
			return
		}

		summary = append(summary, string(body))
	}
	return summary, nil
}
