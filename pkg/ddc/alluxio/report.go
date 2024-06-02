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

package alluxio

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// reportSummary reports alluxio summary
func (e *AlluxioEngine) GetReportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// parse alluxio report summary to cacheStates
func (e AlluxioEngine) ParseReportSummary(s string) cacheStates {

	var states cacheStates

	strs := strings.Split(s, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, summaryPrefixTotalCapacity) {
			totalCacheCapacityAlluxio, _ := utils.FromHumanSize(strings.TrimPrefix(str, summaryPrefixTotalCapacity))
			// Convert Alluxio's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cacheCapacity = utils.BytesSize(float64(totalCacheCapacityAlluxio))
		}
		if strings.HasPrefix(str, summaryPrefixUsedCapacity) {
			usedCacheCapacityAlluxio, _ := utils.FromHumanSize(strings.TrimPrefix(str, summaryPrefixUsedCapacity))
			// Convert Alluxio's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cached = utils.BytesSize(float64(usedCacheCapacityAlluxio))
		}
	}

	return states
}

// reportMetrics reports alluxio metrics
func (e *AlluxioEngine) GetReportMetrics() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportMetrics()
}

// parse alluxio report metric to cacheHitStates
func (e AlluxioEngine) ParseReportMetric(metrics string, cacheHitStates, lastCacheHitStates *cacheHitStates) {
	var localThroughput, remoteThroughput, ufsThroughput int64

	strs := strings.Split(metrics, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		counterPattern := regexp.MustCompile(`\(Type:\sCOUNTER,\sValue:\s(.*)\)`)
		gaugePattern := regexp.MustCompile(`\(Type:\sGAUGE,\sValue:\s(.*)/MIN\)`)
		if strings.HasPrefix(str, metricsPrefixBytesReadLocal) {
			cacheHitStates.bytesReadLocal, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, metricsPrefixBytesReadRemote) {
			cacheHitStates.bytesReadRemote, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, metricsPrefixBytesReadUfsAll) {
			cacheHitStates.bytesReadUfsAll, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, metricsPrefixBytesReadLocalThroughput) {
			localThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, metricsPrefixBytesReadRemoteThroughput) {
			remoteThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, metricsPrefixBytesReadUfsThroughput) {
			ufsThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
		}
	}

	if lastCacheHitStates == nil {
		return
	}

	// Summarize local/remote cache hit ratio
	deltaReadLocal := cacheHitStates.bytesReadLocal - lastCacheHitStates.bytesReadLocal
	deltaReadRemote := cacheHitStates.bytesReadRemote - lastCacheHitStates.bytesReadRemote
	deltaReadUfsAll := cacheHitStates.bytesReadUfsAll - lastCacheHitStates.bytesReadUfsAll
	deltaReadTotal := deltaReadLocal + deltaReadRemote + deltaReadUfsAll

	if deltaReadTotal != 0 {
		cacheHitStates.localHitRatio = fmt.Sprintf("%.1f%%", float64(deltaReadLocal)*100.0/float64(deltaReadTotal))
		cacheHitStates.remoteHitRatio = fmt.Sprintf("%.1f%%", float64(deltaReadRemote)*100.0/float64(deltaReadTotal))
		cacheHitStates.cacheHitRatio = fmt.Sprintf("%.1f%%", float64(deltaReadLocal+deltaReadRemote)*100.0/float64(deltaReadTotal))
	} else {
		// No data is requested in last minute
		cacheHitStates.localHitRatio = "0.0%"
		cacheHitStates.remoteHitRatio = "0.0%"
		cacheHitStates.cacheHitRatio = "0.0%"
	}

	// Summarize local/remote throughput ratio
	totalThroughput := localThroughput + remoteThroughput + ufsThroughput
	if totalThroughput != 0 {
		cacheHitStates.localThroughputRatio = fmt.Sprintf("%.1f%%", float64(localThroughput)*100.0/float64(totalThroughput))
		cacheHitStates.remoteThroughputRatio = fmt.Sprintf("%.1f%%", float64(remoteThroughput)*100.0/float64(totalThroughput))
		cacheHitStates.cacheThroughputRatio = fmt.Sprintf("%.1f%%", float64(localThroughput+remoteThroughput)*100.0/float64(totalThroughput))
	} else {
		cacheHitStates.localThroughputRatio = "0.0%"
		cacheHitStates.remoteThroughputRatio = "0.0%"
		cacheHitStates.cacheThroughputRatio = "0.0%"
	}

}

// reportCapacity reports alluxio capacity
func (e *AlluxioEngine) reportCapacity() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportCapacity()
}
