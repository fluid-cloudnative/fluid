/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CacheEngine.Shutdown", func() {
	It("cleans fuse and worker labels from nodes", func() {
		const (
			runtimeName      = "shutdown-runtime"
			runtimeNamespace = "default"
			nodeName         = "worker-node"
		)
		runtime := &datav1alpha1.CacheRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName, Namespace: runtimeNamespace},
		}
		client := fake.NewFakeClientWithScheme(CacheEngineTestScheme, runtime)
		engine := &CacheEngine{
			Client:      client,
			name:        runtimeName,
			namespace:   runtimeNamespace,
			runtimeType: common.CacheRuntime,
			Log:         fake.NullLogger(),
		}
		info, err := engine.getRuntimeInfo()
		Expect(err).NotTo(HaveOccurred())

		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
				Labels: map[string]string{
					info.GetFuseLabelName():         "true",
					info.GetCommonLabelName():       "true",
					info.GetRuntimeLabelName():      "true",
					info.GetLabelNameForMemory():    "1",
					info.GetLabelNameForDisk():      "1",
					info.GetLabelNameForTotal():     "1",
					utils.GetExclusiveKey():         info.GetExclusiveLabelValue(),
					common.GetDatasetNumLabelName(): "1",
					"unrelated":                     "keep",
				},
			},
		}
		Expect(client.Create(context.Background(), node)).To(Succeed())

		Expect(engine.Shutdown()).To(Succeed())

		updatedNode := &corev1.Node{}
		Expect(client.Get(context.Background(), types.NamespacedName{Name: nodeName}, updatedNode)).To(Succeed())
		Expect(updatedNode.Labels).To(Equal(map[string]string{"unrelated": "keep"}))
	})

	It("returns an error when runtime information cannot be fetched", func() {
		engine := &CacheEngine{
			Client:    fake.NewFakeClientWithScheme(CacheEngineTestScheme),
			name:      "missing-runtime",
			namespace: "default",
			Log:       fake.NullLogger(),
		}

		Expect(engine.Shutdown()).To(HaveOccurred())
	})
})
