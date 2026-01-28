/*
Copyright 2023 The Fluid Author.

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

package kubeclient

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("GetNode", func() {
	var (
		testNodeInputs []*corev1.Node
		testNodes      []runtime.Object
		mockClient     client.Client
	)

	BeforeEach(func() {
		testNodeInputs = []*corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "test1"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "test2"},
			},
		}

		testNodes = []runtime.Object{}
		for _, ns := range testNodeInputs {
			testNodes = append(testNodes, ns.DeepCopy())
		}

		mockClient = fake.NewFakeClientWithScheme(testScheme, testNodes...)
	})

	Context("when node doesn't exist", func() {
		It("should return nil", func() {
			result, err := GetNode(mockClient, "notExist")

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("when node exists", func() {
		It("should return the node", func() {
			result, err := GetNode(mockClient, "test1")

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("test1"))
		})
	})
})

var _ = Describe("IsReady", func() {
	var (
		readyNode    corev1.Node
		notReadyNode corev1.Node
	)

	BeforeEach(func() {
		readyNode = corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test1"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:               corev1.NodeReady,
						Status:             corev1.ConditionTrue,
						Reason:             "FakeReady",
						LastTransitionTime: metav1.Now(),
						LastHeartbeatTime:  metav1.Now(),
					},
				},
			},
		}

		notReadyNode = corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test2"},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:               corev1.NodeReady,
						Status:             corev1.ConditionFalse,
						Reason:             "FakePending",
						LastTransitionTime: metav1.Now(),
						LastHeartbeatTime:  metav1.Now(),
					},
				},
			},
		}
	})

	Context("when node is ready", func() {
		It("should return true", func() {
			result := IsReady(readyNode)
			Expect(result).To(BeTrue())
		})
	})

	Context("when node is not ready", func() {
		It("should return false", func() {
			result := IsReady(notReadyNode)
			Expect(result).To(BeFalse())
		})
	})
})
