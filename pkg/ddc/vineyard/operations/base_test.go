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
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/go-logr/logr"
)

func TestVineyardFileUtils_ReportSummary(t *testing.T) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	patch := ApplyMethod(reflect.TypeOf(client), "Get",
		func(_ *http.Client, url string) (resp *http.Response, err error) {
			if url == "http://vineyard-0.vineyard.default.svc:8080/metrics" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString("metric for pod 1")),
				}, nil
			}
			if url == "http://vineyard-1.vineyard.default.svc:8080/metrics" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString("metric for pod 2")),
				}, nil
			}
			return nil, nil
		})
	defer patch.Reset()

	podNamePrefix := "vineyard"
	port := int32(8080)
	replicas := int32(2)
	namespace := "default"
	mockLogger := logr.Discard()

	vineyard := NewVineyardFileUtils(podNamePrefix, port, replicas, namespace, mockLogger)
	got, err := vineyard.ReportSummary()
	expected := []string{"metric for pod 1", "metric for pod 2"}
	if err != nil || !reflect.DeepEqual(got, expected) {
		t.Errorf("VineyardFileUtils.ReportSummary() got = %v, want %v, err = %v", got, expected, err)
	}

	// test the potenial risk args
	podNamePrefix = "curl -d @/etc/passwd 127.0.0.1 -o-"
	vineyard = NewVineyardFileUtils(podNamePrefix, port, replicas, namespace, mockLogger)
	_, err = vineyard.ReportSummary()
	if err == nil {
		t.Errorf("Expect error, get nil. as the string %s can't pass the regex validation. err = %v", podNamePrefix, err)
	}
}
