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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	datasetUFSFileNum = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dataset_ufs_file_num",
		Help: "Total num of files of a specific dataset",
	}, []string{"dataset"})

	datasetUFSTotalSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "dataset_ufs_total_size",
		Help: "Total size of files in dataset",
	}, []string{"dataset"})
)

var datasetMetricsMap map[string]*datasetMetrics

type datasetMetrics struct {
	datasetKey string
	labels     prometheus.Labels
}

func GetDatasetMetrics(namespace, name string) *datasetMetrics {
	key := labelKeyFunc(namespace, name)

	if m, exists := datasetMetricsMap[key]; exists {
		return m
	}

	ret := &datasetMetrics{
		datasetKey: key,
		labels:     prometheus.Labels{"dataset": key},
	}
	datasetMetricsMap[key] = ret

	return ret
}

func (m *datasetMetrics) SetUFSTotalSize(size float64) {
	datasetUFSTotalSize.With(m.labels).Set(size)
}

func (m *datasetMetrics) SetUFSFileNum(num float64) {
	datasetUFSFileNum.With(m.labels).Set(num)
}

func (m *datasetMetrics) Forget() {
	datasetUFSTotalSize.Delete(m.labels)
	datasetUFSFileNum.Delete(m.labels)

	delete(datasetMetricsMap, m.datasetKey)
}

func init() {
	metrics.Registry.MustRegister(datasetUFSFileNum, datasetUFSTotalSize)
	datasetMetricsMap = map[string]*datasetMetrics{}
}
