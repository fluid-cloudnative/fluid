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

package metrics

import (
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	runtimeSetupErrorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "runtime_setup_error_total",
		Help: "Total num of errors during runtime setup",
	}, []string{"runtime_type", "runtime"})

	runtimeHealthCheckErrorTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "runtime_sync_healthcheck_error_total",
		Help: "Total num of errors during runtime health check",
	}, []string{"runtime_type", "runtime"})
)

var runtimeMetricsMutex *sync.RWMutex // race condition protection for runtimeMetricsMap's concurrent writes
var runtimeMetricsMap map[string]*runtimeMetrics

// runtimeMetrics holds all the metrics related to a specific kind of runtime.
type runtimeMetrics struct {
	runtimeType string
	runtimeKey  string

	labels prometheus.Labels
}

func GetRuntimeMetrics(runtimeType, runtimeNamespace, runtimeName string) *runtimeMetrics {
	key := labelKeyFunc(runtimeNamespace, runtimeName)

	runtimeMetricsMutex.RLock()
	defer runtimeMetricsMutex.RUnlock()
	if m, exists := runtimeMetricsMap[key]; exists {
		return m
	}

	m := &runtimeMetrics{
		runtimeType: runtimeType,
		runtimeKey:  key,
		labels:      prometheus.Labels{"runtime_type": strings.ToLower(runtimeType), "runtime": key},
	}
	runtimeMetricsMutex.Lock()
	defer runtimeMetricsMutex.Unlock()
	runtimeMetricsMap[key] = m
	return m
}

func (m *runtimeMetrics) SetupErrorInc() {
	runtimeSetupErrorTotal.With(m.labels).Inc()
}

func (m *runtimeMetrics) HealthCheckErrorInc() {
	runtimeHealthCheckErrorTotal.With(m.labels).Inc()
}

func (m *runtimeMetrics) Forget() {
	runtimeMetricsMutex.Lock()
	defer runtimeMetricsMutex.Unlock()

	runtimeSetupErrorTotal.Delete(m.labels)
	runtimeHealthCheckErrorTotal.Delete(m.labels)

	delete(runtimeMetricsMap, m.runtimeKey)
}

func init() {
	metrics.Registry.MustRegister(runtimeSetupErrorTotal, runtimeHealthCheckErrorTotal)
	runtimeMetricsMap = map[string]*runtimeMetrics{}
	runtimeMetricsMutex = &sync.RWMutex{}
}
