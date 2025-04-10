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

package thin

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getTestThinEngineNode(client client.Client, name string, namespace string, withRunTime bool) *ThinEngine {
	engine := &ThinEngine{
		runtime:     nil,
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: nil,
		Log:         fake.NullLogger(),
	}
	if withRunTime {
		engine.runtime = &datav1alpha1.ThinRuntime{}
		engine.runtimeInfo, _ = base.BuildRuntimeInfo(name, namespace, common.ThinRuntime)
	}
	return engine
}

func TestSyncScheduleInfoToCacheNodes(t *testing.T) {
	type fields struct {
		worker    *appsv1.StatefulSet
		pods      []*v1.Pod
		ds        *appsv1.DaemonSet
		nodes     []*v1.Node
		name      string
		namespace string
	}
	testcases := []struct {
		name      string
		fields    fields
		nodeNames []string
	}{
		{
			name: "create",
			fields: fields{
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
								"app":     "thin",
								"role":    "thin-worker",
								"release": "spark",
							},
						},
					},
				},
				pods: []*v1.Pod{{
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
							"app":              "thin",
							"role":             "thin-worker",
							"release":          "spark",
							"fluid.io/dataset": "big-data-spark",
						},
					},
					Spec: v1.PodSpec{
						NodeName: "node1",
					},
				}},
				nodes: []*v1.Node{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				}},
			},
			nodeNames: []string{"node1"},
		},
		{
			name: "add",
			fields: fields{
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
								"app":     "thin",
								"role":    "thin-worker",
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
								"app":              "thin",
								"role":             "thin-worker",
								"release":          "hbase",
								"fluid.io/dataset": "big-data-hbase",
							},
						},
						Spec: v1.PodSpec{NodeName: "node3"},
					},
				},
				nodes: []*v1.Node{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name:   "node2",
						Labels: map[string]string{"fluid.io/s-default-hbase": "true"},
					},
				}},
			},
			nodeNames: []string{"node3"},
		},
		{
			name: "noController",
			fields: fields{
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
								"app":     "thin",
								"role":    "thin-worker",
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
								"app":              "thin",
								"role":             "thin-worker",
								"release":          "hbase-a",
								"fluid.io/dataset": "big-data-hbase-a",
							},
						},
						Spec: v1.PodSpec{NodeName: "node5"},
					},
				},
				nodes: []*v1.Node{{
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
				}},
			},
			nodeNames: []string{},
		},
	}

	runtimeObjs := []runtime.Object{}

	for _, testcase := range testcases {
		runtimeObjs = append(runtimeObjs, testcase.fields.worker)

		if testcase.fields.ds != nil {
			runtimeObjs = append(runtimeObjs, testcase.fields.ds)
		}
		for _, pod := range testcase.fields.pods {
			runtimeObjs = append(runtimeObjs, pod)
		}

		for _, node := range testcase.fields.nodes {
			runtimeObjs = append(runtimeObjs, node)
		}
		// runtimeObjs = append(runtimeObjs, testcase.fields.pods)
	}
	c := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)

	for _, testcase := range testcases {
		engine := getTestThinEngineNode(c, testcase.fields.name, testcase.fields.namespace, true)
		runtimeInfo, err := base.BuildRuntimeInfo(testcase.fields.name,
			testcase.fields.namespace,
			common.ThinRuntime)
		if err != nil {
			t.Errorf("BuildRuntimeInfo() error = %v", err)
		}

		engine.Helper = ctrlhelper.BuildHelper(runtimeInfo, c, engine.Log)
		err = engine.SyncScheduleInfoToCacheNodes()
		if err != nil {
			t.Errorf("Got error %t.", err)
		}

		nodeList := &v1.NodeList{}
		datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", engine.runtimeInfo.GetRuntimeLabelName()))
		if err != nil {
			return
		}

		err = c.List(context.TODO(), nodeList, &client.ListOptions{
			LabelSelector: datasetLabels,
		})

		if err != nil {
			t.Errorf("Got error %t.", err)
		}

		nodeNames := []string{}
		for _, node := range nodeList.Items {
			nodeNames = append(nodeNames, node.Name)
		}

		if len(testcase.nodeNames) == 0 && len(nodeNames) == 0 {
			continue
		}

		if !reflect.DeepEqual(testcase.nodeNames, nodeNames) {
			t.Errorf("test case %v fail to sync node labels, wanted %v, got %v", testcase.name, testcase.nodeNames, nodeNames)
		}

	}
}
