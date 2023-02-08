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
	"fmt"
	"strings"

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

// RuntimeMetrics holds all the metrics related to a specific kind of runtime.
type RuntimeMetrics struct {
	runtimeType string
	runtimeKey  string

	setupErrorTotal       prometheus.Counter
	healthCheckErrorTotal prometheus.Counter
}

func NewRuntimeMetrics(runtimeType, runtimeNamespace, runtimeName string) *RuntimeMetrics {
	key := fmt.Sprintf("%s/%s", runtimeNamespace, runtimeName)
	label := prometheus.Labels{"runtime_type": strings.ToLower(runtimeType), "runtime": key}
	metrics := &RuntimeMetrics{
		runtimeType:           runtimeType,
		runtimeKey:            key,
		setupErrorTotal:       runtimeSetupErrorTotal.With(label),
		healthCheckErrorTotal: runtimeHealthCheckErrorTotal.With(label),
	}

	return metrics
}

func (m *RuntimeMetrics) SetupErrorInc() {
	m.setupErrorTotal.Inc()
}

func (m *RuntimeMetrics) HealthCheckErrorInc() {
	m.healthCheckErrorTotal.Inc()
}

func init() {
	metrics.Registry.MustRegister(runtimeSetupErrorTotal, runtimeHealthCheckErrorTotal)
}
