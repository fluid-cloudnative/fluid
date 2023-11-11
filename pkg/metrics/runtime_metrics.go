/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package metrics

import (
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

var runtimeMetricsMap map[string]*runtimeMetrics

// runtimeMetrics holds all the metrics related to a specific kind of runtime.
type runtimeMetrics struct {
	runtimeType string
	runtimeKey  string

	labels prometheus.Labels
}

func GetRuntimeMetrics(runtimeType, runtimeNamespace, runtimeName string) *runtimeMetrics {
	key := labelKeyFunc(runtimeNamespace, runtimeName)
	if m, exists := runtimeMetricsMap[key]; exists {
		return m
	}

	m := &runtimeMetrics{
		runtimeType: runtimeType,
		runtimeKey:  key,
		labels:      prometheus.Labels{"runtime_type": strings.ToLower(runtimeType), "runtime": key},
	}
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
	runtimeSetupErrorTotal.Delete(m.labels)
	runtimeHealthCheckErrorTotal.Delete(m.labels)

	delete(runtimeMetricsMap, m.runtimeKey)
}

func init() {
	metrics.Registry.MustRegister(runtimeSetupErrorTotal, runtimeHealthCheckErrorTotal)
	runtimeMetricsMap = map[string]*runtimeMetrics{}
}
