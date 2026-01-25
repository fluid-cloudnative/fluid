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
	"reflect"
	"testing"

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

func TestSyncScheduleInfoToCacheNodes(t *testing.T) {
	type fields struct {
		worker    *appsv1.StatefulSet
		pods      []*v1.Pod
		nodes     []*v1.Node
		name      string
		namespace string
	}
	testcases := []struct {
		name      string
		fields    fields
		nodeNames []string
	}{}

	testcaseCnt := 0
	makeDatasetResourcesFn := func(dsName string, dsNamespace string, stsPodNodeNames []string) fields {
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

	fields1 := makeDatasetResourcesFn("spark", testNodeNamespace, []string{"node1"})
	fields1.nodes = append(fields1.nodes, &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}})
	testcases = append(testcases, struct {
		name      string
		fields    fields
		nodeNames []string
	}{
		name:      "create",
		fields:    fields1,
		nodeNames: []string{"node1"},
	})

	fields2 := makeDatasetResourcesFn("hbase", testNodeNamespace, []string{"node2", "node3"})
	fields2.nodes = append(fields2.nodes,
		&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node3"}},
		&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node2", Labels: map[string]string{"fluid.io/s-big-data-hbase": "true"}}},
	)
	testcases = append(testcases, struct {
		name      string
		fields    fields
		nodeNames []string
	}{
		name:      "add",
		fields:    fields2,
		nodeNames: []string{"node2", "node3"},
	})

	fields3 := makeDatasetResourcesFn("hbase-a", testNodeNamespace, []string{"node4", "node5"})
	fields3.pods[1].OwnerReferences = []metav1.OwnerReference{}
	fields3.nodes = append(fields3.nodes,
		&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node5"}},
		&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node4", Labels: map[string]string{"fluid.io/s-big-data-hbase-a": "true"}}},
	)
	testcases = append(testcases, struct {
		name      string
		fields    fields
		nodeNames []string
	}{
		name:      "noController",
		fields:    fields3,
		nodeNames: []string{"node4"},
	})

	fields4 := makeDatasetResourcesFn("hbase-b", testNodeNamespace, []string{})
	fields4.nodes = append(fields4.nodes,
		&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node6", Labels: map[string]string{"fluid.io/s-big-data-hbase-b": "true", "fluid.io/s-goosefs-big-data-hbase-b": "true"}}},
	)
	testcases = append(testcases, struct {
		name      string
		fields    fields
		nodeNames []string
	}{
		name:      "remove",
		fields:    fields4,
		nodeNames: []string{},
	})

	runtimeObjs := []runtime.Object{}

	for _, testcase := range testcases {
		runtimeObjs = append(runtimeObjs, testcase.fields.worker)

		for _, pod := range testcase.fields.pods {
			runtimeObjs = append(runtimeObjs, pod)
		}

		for _, node := range testcase.fields.nodes {
			runtimeObjs = append(runtimeObjs, node)
		}
	}
	c := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)

	for _, testcase := range testcases {
		engine := getTestGooseFSEngineNode(c, testcase.fields.name, testcase.fields.namespace, true)
		err := engine.SyncScheduleInfoToCacheNodes()
		if err != nil {
			t.Errorf("testcase %s: Got error %v", testcase.name, err)
			continue
		}

		nodeList := &v1.NodeList{}
		datasetLabels, parseErr := labels.Parse(fmt.Sprintf(testNodeLabelSelector, engine.runtimeInfo.GetCommonLabelName()))
		if parseErr != nil {
			t.Fatalf("testcase %s: Got error parsing labels: %v", testcase.name, parseErr)
		}

		listErr := c.List(context.TODO(), nodeList, &client.ListOptions{
			LabelSelector: datasetLabels,
		})

		if listErr != nil {
			t.Errorf("testcase %s: Got error listing nodes: %v", testcase.name, listErr)
			continue
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
