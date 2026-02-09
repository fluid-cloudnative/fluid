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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testNodeNamespace     = "big-data"
	testNodeLabelApp      = "goosefs"
	testNodeLabelRole     = "goosefs-worker"
	testNodeAPIVersion    = "apps/v1"
	testNodeKindSts       = "StatefulSet"
	testNodeLabelDataset  = "fluid.io/dataset"
	testNodeLabelSelector = "%s=true"
)

func getTestGooseFSEngineNode(c client.Client, name string, namespace string, withRunTime bool) *GooseFSEngine {
	engine := &GooseFSEngine{
		runtime:     nil,
		name:        name,
		namespace:   namespace,
		Client:      c,
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
			nodes     []*v1.Node
			name      string
			namespace string
		}

		testcaseCnt := 0
		makeDatasetResources := func(dsName string, dsNamespace string, stsPodNodeNames []string) fields {
			testcaseCnt++
			ret := fields{
				name:      dsName,
				namespace: dsNamespace,
				worker: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       testNodeKindSts,
						APIVersion: testNodeAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      dsName + "-worker",
						Namespace: dsNamespace,
						UID:       types.UID(fmt.Sprintf("uid%d", testcaseCnt)),
					},
					Spec: appsv1.StatefulSetSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":     testNodeLabelApp,
								"role":    testNodeLabelRole,
								"release": dsName,
							},
						},
					},
				},
				pods: []*v1.Pod{},
			}

			for idx, nodeName := range stsPodNodeNames {
				ret.pods = append(ret.pods, &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("%s-worker-%d", dsName, idx),
						Namespace: dsNamespace,
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       testNodeKindSts,
							APIVersion: testNodeAPIVersion,
							Name:       dsName + "-worker",
							UID:        types.UID(fmt.Sprintf("uid%d", testcaseCnt)),
							Controller: ptr.To(true),
						}},
						Labels: map[string]string{
							"app":                testNodeLabelApp,
							"role":               testNodeLabelRole,
							"release":            dsName,
							testNodeLabelDataset: fmt.Sprintf("%s-%s", dsNamespace, dsName),
						},
					},
					Spec: v1.PodSpec{
						NodeName: nodeName,
					},
				})
			}

			return ret
		}

		fields1 := makeDatasetResources("spark", testNodeNamespace, []string{"node1"})
		fields1.nodes = append(fields1.nodes, &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})

		fields2 := makeDatasetResources("hbase", testNodeNamespace, []string{"node2", "node3"})
		fields2.nodes = append(fields2.nodes,
			&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node3"}},
			&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"fluid.io/s-big-data-hbase": "true"}}},
		)

		fields3 := makeDatasetResources("hbase-a", testNodeNamespace, []string{"node4", "node5"})
		fields3.pods[1].OwnerReferences = []metav1.OwnerReference{}
		fields3.nodes = append(fields3.nodes,
			&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node5"}},
			&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node4", Labels: map[string]string{"fluid.io/s-big-data-hbase-a": "true"}}},
		)

		fields4 := makeDatasetResources("hbase-b", testNodeNamespace, []string{})
		fields4.nodes = append(fields4.nodes,
			&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node6", Labels: map[string]string{
				"fluid.io/s-big-data-hbase-b":         "true",
				"fluid.io/s-goosefs-big-data-hbase-b": "true",
			}}},
		)

		DescribeTable("should sync schedule info to cache nodes correctly",
			func(f fields, expectedNodeNames []string) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, f.worker)

				for _, pod := range f.pods {
					runtimeObjs = append(runtimeObjs, pod)
				}
				for _, node := range f.nodes {
					runtimeObjs = append(runtimeObjs, node)
				}

				c := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)
				engine := getTestGooseFSEngineNode(c, f.name, f.namespace, true)

				err := engine.SyncScheduleInfoToCacheNodes()
				Expect(err).NotTo(HaveOccurred())

				nodeList := &v1.NodeList{}
				datasetLabels, err := labels.Parse(fmt.Sprintf(testNodeLabelSelector, engine.runtimeInfo.GetCommonLabelName()))
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
			Entry("create", fields1, []string{"node1"}),
			Entry("add", fields2, []string{"node2", "node3"}),
			Entry("noController", fields3, []string{"node4"}),
			Entry("remove", fields4, []string{}),
		)
	})
})
