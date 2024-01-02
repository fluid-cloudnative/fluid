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

package vineyard

import (
	"strconv"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/vineyard/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// reportSummary reports vineyard summary
func (e *VineyardEngine) GetReportSummary() (summary []string, err error) {
	podNamePrefix := e.getWorkerName()
	port := e.getWorkerPodExporterPort()
	replicas := e.getWorkerReplicas()
	fileUtils := operations.NewVineyardFileUtils(podNamePrefix, port, replicas, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// parse vineyard report summary to cached
func (e VineyardEngine) ParseReportSummary(s []string) string {
	var cached float64
	var lastCached string
	for _, str := range s {
		lines := strings.Split(str, "\n")

		// the metrics of vineyard is like:
		// grok_exporter_lines_processing_time_microseconds_total{metric="instances_memory_usage_bytes"} 3329
		// instances_memory_usage_bytes{instance="0",user="vineyard-worker-0"} 1.4662007808e+10
		for _, line := range lines {
			if strings.Contains(line, "instances_memory_usage_bytes") {
				lastSpaceIndex := strings.LastIndex(line, " ")
				if lastSpaceIndex == -1 {
					continue
				}
				lastCached = line[lastSpaceIndex+1:]
			}
		}

		cache, err := strconv.ParseFloat(lastCached, 64)
		if err != nil {
			e.Log.Error(err, "parse cached error")
		}
		cached += float64(cache)
	}

	return utils.BytesSize(cached)
}
