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
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	runtimeschema "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTestGooseFSEngineNode(client client.Client, name string, namespace string, withRunTime bool) *GooseFSEngine {
	engine := &GooseFSEngine{
		runtime:     nil,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: nil,
		Log:         fake.NullLogger(),
	}
	if withRunTime {
		engine.runtime = &v1alpha1.GooseFSRuntime{}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo(name, namespace, common.GooseFSRuntime)
	}
	return engine
}

var _ = Describe("GooseFSEngine", func() {
	Describe("SyncScheduleInfoToCacheNodes", func() {
		type fields struct {
			worker    *appsv1.StatefulSet
			pods      []*v1.Pod
			ds        *appsv1.DaemonSet
			nodes     []*v1.Node
			name      string
			namespace string
		}

		DescribeTable("should sync schedule info to cache nodes correctly",
			func(fields fields, expectedNodeNames []string) {
				runtimeObjs := []runtimeschema.Object{}
				runtimeObjs = append(runtimeObjs, fields.worker)

				if fields.ds != nil {
					runtimeObjs = append(runtimeObjs, fields.ds)
				}
				for _, pod := range fields.pods {
					runtimeObjs = append(runtimeObjs, pod)
				}
				for _, node := range fields.nodes {
					runtimeObjs = append(runtimeObjs, node)
				}

				c := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
				engine := getTestGooseFSEngineNode(c, fields.name, fields.namespace, true)

				err := engine.SyncScheduleInfoToCacheNodes()
				Expect(err).NotTo(HaveOccurred())

				nodeList := &v1.NodeList{}
				datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", engine.runtimeInfo.GetCommonLabelName()))
				Expect(err).NotTo(HaveOccurred())

				err = c.List(context.TODO(), nodeList, &client.ListOptions{
					LabelSelector: datasetLabels,
				})
				Expect(err).NotTo(HaveOccurred())

				nodeNames := []string{}
				for _, node := range nodeList.Items {
					nodeNames = append(nodeNames, node.Name)
				}

				if len(expectedNodeNames) == 0 && len(nodeNames) == 0 {
					return
				}

				Expect(nodeNames).To(Equal(expectedNodeNames),
					fmt.Sprintf("wanted %v, got %v", expectedNodeNames, nodeNames))
			},
			Entry("create",
				fields{
					name:      "spark",
					namespace: "big-data",
					worker: &appsv1.StatefulSet{
						TypeMeta: metav1.TypeMeta{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spark-worker",
							Namespace: "big-data",
							UID:       "uid1",
						},
						Spec: appsv1.StatefulSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app":     "goosefs",
									"role":    "goosefs-worker",
									"release": "spark",
								},
							},
						},
					},
					pods: []*v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "spark-worker-0",
								Namespace: "big-data",
								OwnerReferences: []metav1.OwnerReference{{
									Kind:       "StatefulSet",
									APIVersion: "apps/v1",
									Name:       "spark-worker",
									UID:        "uid1",
									Controller: ptr.To(true),
								}},
								Labels: map[string]string{
									"app":              "goosefs",
									"role":             "goosefs-worker",
									"release":          "spark",
									"fluid.io/dataset": "big-data-spark",
								},
							},
							Spec: v1.PodSpec{
								NodeName: "node1",
							},
						},
					},
					nodes: []*v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node1",
							},
						},
					},
				},
				[]string{"node1"},
			),
			Entry("add",
				fields{
					name:      "hbase",
					namespace: "big-data",
					worker: &appsv1.StatefulSet{
						TypeMeta: metav1.TypeMeta{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase-worker",
							Namespace: "big-data",
							UID:       "uid2",
						},
						Spec: appsv1.StatefulSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app":     "goosefs",
									"role":    "goosefs-worker",
									"release": "hbase",
								},
							},
						},
					},
					pods: []*v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "hbase-worker-0",
								Namespace: "big-data",
								OwnerReferences: []metav1.OwnerReference{{
									Kind:       "StatefulSet",
									APIVersion: "apps/v1",
									Name:       "hbase-worker",
									UID:        "uid2",
									Controller: ptr.To(true),
								}},
								Labels: map[string]string{
									"app":              "goosefs",
									"role":             "goosefs-worker",
									"release":          "hbase",
									"fluid.io/dataset": "big-data-hbase",
								},
							},
							Spec: v1.PodSpec{
								NodeName: "node3",
							},
						},
					},
					nodes: []*v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node3",
							},
						}, {
							ObjectMeta: metav1.ObjectMeta{
								Name: "node2",
								Labels: map[string]string{
									"fluid.io/s-default-hbase": "true",
								},
							},
						},
					},
				},
				[]string{"node3"},
			),
			Entry("noController",
				fields{
					name:      "hbase-a",
					namespace: "big-data",
					worker: &appsv1.StatefulSet{
						TypeMeta: metav1.TypeMeta{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase-a-worker",
							Namespace: "big-data",
							UID:       "uid3",
						},
						Spec: appsv1.StatefulSetSpec{
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app":     "goosefs",
									"role":    "goosefs-worker",
									"release": "hbase-a",
								},
							},
						},
					},
					pods: []*v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "hbase-a-worker-0",
								Namespace: "big-data",
								Labels: map[string]string{
									"app":              "goosefs",
									"role":             "goosefs-worker",
									"release":          "hbase-a",
									"fluid.io/dataset": "big-data-hbase-a",
								},
							},
							Spec: v1.PodSpec{
								NodeName: "node5",
							},
						},
					},
					nodes: []*v1.Node{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "node5",
							},
						}, {
							ObjectMeta: metav1.ObjectMeta{
								Name: "node4",
								Labels: map[string]string{
									"fluid.io/s-default-hbase-a": "true",
								},
							},
						},
					},
				},
				[]string{},
			),
		)
	})
})
