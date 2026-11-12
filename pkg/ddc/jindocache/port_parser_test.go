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

package jindocache

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var cfg = `
[bigboot]
logger.dir =  /dev/shm/default/oss-tf-dataset/bigboot/log
logger.cleanner.enable = true

[jindocache-common]
jfs.namespaces = spark
jfs.namespaces.spark.oss.uri = oss://tensorflow-datasets.oss-cn-shanghai-internal.aliyuncs.com/
namespace.backend.type = rocksdb
namespace.blocklet.cache.size = 1000000
namespace.filelet.cache.size = 100000
namespace.meta-dir = /dev/shm/default/oss-tf-dataset/bigboot/server
namespace.rpc.port = 18000
namespace.filelet.atime.enable = false

[jindocache-storage]
namespace.meta-dir = /dev/shm/default/oss-tf-dataset/bigboot/bignode
storage.data-dirs = /dev/shm/default/oss-tf-dataset/bigboot
storage.data-dirs.capacities = 10g
storage.ram.cache.size = 10g
storage.rpc.port = 18001
namespace.meta-dir = /dev/shm/default/oss-tf-dataset/bigboot/bignode
storage.compaction.enable = false

[jindocache-namespace]
client.oss.upload.queue.size = 5
client.oss.upload.threads = 4
client.storage.rpc.port = 18001
`

var _ = Describe("parsePortsFromConfigMap", func() {
	It("should parse ports from configMap", func() {
		configMap := &v1.ConfigMap{
			Data: map[string]string{
				"jindocache.cfg": cfg,
			},
		}

		gotPorts, err := parsePortsFromConfigMap(configMap)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotPorts).To(Equal([]int{18000, 18001}))
	})

	It("should collect reserved ports from jindo runtime configmaps", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "spark", Namespace: "fluid"},
			Status: datav1alpha1.DatasetStatus{
				Runtimes: []datav1alpha1.Runtime{{
					Category:  common.AccelerateCategory,
					Name:      "spark",
					Namespace: "fluid",
					Type:      "jindo",
				}},
			},
		}
		nonJindoDataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "fluid"},
			Status: datav1alpha1.DatasetStatus{
				Runtimes: []datav1alpha1.Runtime{{
					Name:      "other",
					Namespace: "fluid",
					Type:      "alluxio",
				}},
			},
		}
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "spark-jindofs-config", Namespace: "fluid"},
			Data:       map[string]string{"jindocache.cfg": cfg},
		}

		objects := []runtime.Object{dataset, nonJindoDataset, configMap}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, objects...)

		ports, err := GetReservedPorts(fakeClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(ports).To(Equal([]int{18000, 18001}))
	})
})
