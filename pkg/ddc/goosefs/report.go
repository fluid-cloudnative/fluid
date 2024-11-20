/*
Copyright 2022 The Fluid Authors.

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

package goosefs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	ddctypes "github.com/fluid-cloudnative/fluid/pkg/ddc/types"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// reportSummary reports goosefs summary
func (e *GooseFSEngine) GetReportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// parse goosefs report summary to cacheStates
func (e GooseFSEngine) ParseReportSummary(s string) cacheStates {

	var states cacheStates

	strs := strings.Split(s, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if strings.HasPrefix(str, ddctypes.SummaryPrefixTotalCapacity) {
			totalCacheCapacityGooseFS, _ := utils.FromHumanSize(strings.TrimPrefix(str, ddctypes.SummaryPrefixTotalCapacity))
			// Convert GooseFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cacheCapacity = utils.BytesSize(float64(totalCacheCapacityGooseFS))
		}
		if strings.HasPrefix(str, ddctypes.SummaryPrefixUsedCapacity) {
			usedCacheCapacityGooseFS, _ := utils.FromHumanSize(strings.TrimPrefix(str, ddctypes.SummaryPrefixUsedCapacity))
			// Convert GooseFS's binary byte units to Fluid's binary byte units
			// e.g. 10KB -> 10KiB, 2GB -> 2GiB
			states.cached = utils.BytesSize(float64(usedCacheCapacityGooseFS))
		}
	}

	return states
}

// reportMetrics reports goosefs metrics
func (e *GooseFSEngine) GetReportMetrics() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportMetrics()
}

// parse goosefs report metric to cacheHitStates
func (e GooseFSEngine) ParseReportMetric(metrics string, cacheHitStates, lastCacheHitStates *cacheHitStates) {
	var localThroughput, remoteThroughput, ufsThroughput int64

	strs := strings.Split(metrics, "\n")
	for _, str := range strs {
		str = strings.TrimSpace(str)
		counterPattern := regexp.MustCompile(`\(Type:\sCOUNTER,\sValue:\s(.*)\)`)
		gaugePattern := regexp.MustCompile(`\(Type:\sGAUGE,\sValue:\s(.*)/MIN\)`)
		if strings.HasPrefix(str, MetricsPrefixBytesReadLocal) {
			cacheHitStates.bytesReadLocal, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, MetricsPrefixBytesReadRemote) {
			cacheHitStates.bytesReadRemote, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, MetricsPrefixBytesReadUfsAll) {
			cacheHitStates.bytesReadUfsAll, _ = utils.FromHumanSize(counterPattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, MetricsPrefixBytesReadLocalThroughput) {
			localThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, MetricsPrefixBytesReadRemoteThroughput) {
			remoteThroughput, _ = utils.FromHumanSize(gaugePattern.FindStringSubmatch(str)[1])
			continue
		}

		if strings.HasPrefix(str, MetricsPrefixBytesReadUfsThroughput) {
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

// reportCapacity reports goosefs capacity
func (e *GooseFSEngine) reportCapacity() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportCapacity()
}
