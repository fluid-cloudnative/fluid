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

package vineyard

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func TestParseReportSummary(t *testing.T) {
	testCases := map[string]struct {
		summary []string
		want    string
	}{
		"test ParseReportSummary case 1": {
			summary: mockVineyardReportSummaryForParseReport(),
			want:    utils.BytesSize(5555),
		},
	}

	for k, item := range testCases {
		got := VineyardEngine{}.ParseReportSummary(item.summary)
		if !reflect.DeepEqual(item.want, got) {
			t.Errorf("%s check failure,want:%+v,got:%+v", k, item.want, got)
		}
	}
}

func mockVineyardReportSummaryForParseReport() []string {
	summary := []string{
		`grok_exporter_lines_processing_time_microseconds_total{metric="instances_memory_usage_bytes"} 1234`,
		`instances_memory_usage_bytes{instance="0",user="vineyard-worker-0"} 4321`,
	}

	return summary
}
