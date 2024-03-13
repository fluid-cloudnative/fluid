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
	"sync"

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

var datasetMetricsMap sync.Map // race condition protection for datasetMetricsMap's concurrent writes

type datasetMetrics struct {
	datasetKey string
	labels     prometheus.Labels
}

func GetOrCreateDatasetMetrics(namespace, name string) *datasetMetrics {
	key := labelKeyFunc(namespace, name)
	m := &datasetMetrics{
		datasetKey: key,
		labels:     prometheus.Labels{"dataset": key},
	}

	ret, _ := datasetMetricsMap.LoadOrStore(key, m)

	return ret.(*datasetMetrics)
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

	datasetMetricsMap.Delete(m.datasetKey)
}

func init() {
	metrics.Registry.MustRegister(datasetUFSFileNum, datasetUFSTotalSize)
	datasetMetricsMap = sync.Map{}
}
