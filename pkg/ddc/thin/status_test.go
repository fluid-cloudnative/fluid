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

package thin

import (
	"context"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/utils/pointer"
)

func TestThinEngine_CheckAndUpdateRuntimeStatus(t *testing.T) {
	Convey("Test CheckAndUpdateRuntimeStatus ", t, func() {
		Convey("CheckAndUpdateRuntimeStatus success", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("thin", "fluid", "thin", datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("fail to create the runtimeInfo with error %v", err)
			}
			runtimeInfo.SetupFuseDeployMode(false, nil)

			var workerInputs = []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin2-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      2,
						ReadyReplicas: 2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-fuse-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: utilpointer.Int32Ptr(1),
					},
				},
			}

			var fuseInputs = []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1-fuse",
						Namespace: "fluid",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin2-fuse",
						Namespace: "fluid",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-worker-fuse",
						Namespace: "fluid",
					},
				},
			}

			runtimeInputs := []*datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 3, // 2
					},
					Status: datav1alpha1.RuntimeStatus{
						CurrentWorkerNumberScheduled: 2,
						CurrentMasterNumberScheduled: 2, // 0
						CurrentFuseNumberScheduled:   2,
						DesiredMasterNumberScheduled: 3,
						DesiredWorkerNumberScheduled: 2,
						DesiredFuseNumberScheduled:   3,
						Conditions: []datav1alpha1.RuntimeCondition{
							utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
							utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
						},
						WorkerPhase: "NotReady",
						FusePhase:   "NotReady",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin2",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 2,
					},
					Status: datav1alpha1.RuntimeStatus{
						CurrentWorkerNumberScheduled: 3,
						CurrentMasterNumberScheduled: 3,
						CurrentFuseNumberScheduled:   3,
						DesiredMasterNumberScheduled: 2,
						DesiredWorkerNumberScheduled: 3,
						DesiredFuseNumberScheduled:   2,
						Conditions: []datav1alpha1.RuntimeCondition{
							utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
							utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
						},
						WorkerPhase: "NotReady",
						FusePhase:   "NotReady",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-worker",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 2,
					},
					Status: datav1alpha1.RuntimeStatus{
						CurrentWorkerNumberScheduled: 2,
						CurrentMasterNumberScheduled: 2,
						CurrentFuseNumberScheduled:   2,
						DesiredMasterNumberScheduled: 2,
						DesiredWorkerNumberScheduled: 2,
						DesiredFuseNumberScheduled:   2,
						WorkerPhase:                  "NotReady",
						FusePhase:                    "NotReady",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-fuse",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 2,
					},
					Status: datav1alpha1.RuntimeStatus{
						CurrentWorkerNumberScheduled: 2,
						CurrentMasterNumberScheduled: 2,
						CurrentFuseNumberScheduled:   2,
						DesiredMasterNumberScheduled: 2,
						DesiredWorkerNumberScheduled: 2,
						DesiredFuseNumberScheduled:   2,
						WorkerPhase:                  "NotReady",
						FusePhase:                    "NotReady",
					},
				},
			}

			var datasetInputs = []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "ceph://myceph",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin2",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://my-pvc",
							},
						},
					},
				},
			}

			objs := []runtime.Object{}

			for _, workerInput := range workerInputs {
				objs = append(objs, workerInput.DeepCopy())
			}

			for _, runtimeInput := range runtimeInputs {
				objs = append(objs, runtimeInput.DeepCopy())
			}

			for _, fuseInput := range fuseInputs {
				objs = append(objs, fuseInput.DeepCopy())
			}

			for _, datasetInput := range datasetInputs {
				objs = append(objs, datasetInput)
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

			testCases := []struct {
				testName   string
				name       string
				namespace  string
				isErr      bool
				deprecated bool
			}{
				{
					testName:  "thin1",
					name:      "thin1",
					namespace: "fluid",
				},
				{
					testName:  "thin2",
					name:      "thin2",
					namespace: "fluid",
				},
				{
					testName:  "no-fuse",
					name:      "no-fuse",
					namespace: "fluid",
					isErr:     true,
				},
				{
					testName:  "no-worker",
					name:      "no-worker",
					namespace: "fluid",
					isErr:     true,
				},
			}

			for _, testCase := range testCases {
				engine := newThinEngineREP(fakeClient, testCase.name, testCase.namespace)

				_, err := engine.CheckAndUpdateRuntimeStatus()
				if err != nil && !testCase.isErr {
					t.Errorf("testcase %s Failed due to %v", testCase.testName, err)
				}
			}
		})
	})

}

func TestThinEngine_UpdateRuntimeSetConfigIfNeeded(t *testing.T) {
	type fields struct {
		worker    *appsv1.StatefulSet
		pods      []*corev1.Pod
		ds        *appsv1.DaemonSet
		nodes     []*corev1.Node
		name      string
		namespace string
	}
	testcases := []struct {
		name        string
		fields      fields
		configMap   *corev1.ConfigMap
		want        string
		wantUpdated bool
	}{
		{
			name: "create",
			fields: fields{
				name:      "spark",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker",
						Namespace: "big-data",
						UID:       "uid1",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				pods: []*corev1.Pod{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-worker-0",
						Namespace: "big-data",
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "apps/v1",
							Name:       "spark-worker",
							UID:        "uid1",
							Controller: utilpointer.BoolPtr(true),
						}},
						Labels: map[string]string{
							"app":              "thin",
							"role":             "thin-worker",
							"fluid.io/dataset": "big-data-spark",
						},
					},
					Spec: corev1.PodSpec{
						NodeName: "node1",
					},
				}},
				nodes: []*corev1.Node{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
						Labels: map[string]string{
							"fluid.io/f-big-data-spark": "true",
						},
					}, Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "192.168.0.1",
							},
						},
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
						Labels: map[string]string{
							"fluid.io/f-big-data-spark":      "true",
							"fluid.io/s-big-data-spark":      "true",
							"fluid.io/s-thin-big-data-spark": "true",
						},
					}, Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "192.168.0.2",
							},
						},
					},
				}},
			},
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark-runtimeset",
					Namespace: "big-data",
				}, Data: map[string]string{
					"runtime.json": "",
				},
			}, want: "{\"workers\":[\"192.168.0.2\"],\"fuses\":[\"192.168.0.1\",\"192.168.0.2\"]}",
			wantUpdated: true,
		},
		{
			name: "nochange_configmap",
			fields: fields{
				name:      "hbase",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker",
						Namespace: "big-data",
						UID:       "uid2",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				pods: []*corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase-worker-0",
							Namespace: "big-data",
							OwnerReferences: []metav1.OwnerReference{{
								Kind:       "StatefulSet",
								APIVersion: "apps/v1",
								Name:       "hbase-worker",
								UID:        "uid2",
								Controller: utilpointer.BoolPtr(true),
							}},
							Labels: map[string]string{
								"app":              "thin",
								"role":             "thin-worker",
								"fluid.io/dataset": "big-data-hbase",
							},
						},
						Spec: corev1.PodSpec{NodeName: "node3"},
					},
				},
				nodes: []*corev1.Node{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							"fluid.io/f-big-data-hbase":      "true",
							"fluid.io/s-big-data-hbase":      "true",
							"fluid.io/s-thin-big-data-hbase": "true",
						},
					},
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "10.0.0.2",
							},
						},
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name: "node4",
						Labels: map[string]string{"fluid.io/s-default-hbase": "true",
							"fluid.io/s-thin-big-data-hbase": "true"},
					}, Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "172.17.0.9",
							},
						},
					},
				}},
			},
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-runtimeset",
					Namespace: "big-data",
				}, Data: map[string]string{
					"runtime.json": "{\"workers\":[\"10.0.0.2\",\"172.17.0.9\"],\"fuses\":[\"10.0.0.2\"]}",
				},
			}, want: "{\"workers\":[\"10.0.0.2\",\"172.17.0.9\"],\"fuses\":[\"10.0.0.2\"]}",
			wantUpdated: false,
		},
		{
			name: "nomatch",
			fields: fields{
				name:      "hbase-a",
				namespace: "big-data",
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-a-worker",
						Namespace: "big-data",
						UID:       "uid3",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				pods: []*corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase-a-worker-0",
							Namespace: "big-data",
							Labels: map[string]string{
								"app":              "thin",
								"role":             "thin-worker",
								"fluid.io/dataset": "big-data-hbase-a",
							},
						},
						Spec: corev1.PodSpec{NodeName: "node5"},
					},
				},
				nodes: []*corev1.Node{{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node5",
					}, Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
					},
				}, {
					ObjectMeta: metav1.ObjectMeta{
						Name: "node6",
						Labels: map[string]string{
							"fluid.io/s-default-hbase-a": "true",
						},
					}, Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: "10.0.0.2",
							},
						},
					},
				}},
			},
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-a-runtimeset",
					Namespace: "big-data",
				}, Data: map[string]string{
					"runtime.json": "{\"workers\":[],\"fuses\":[]}",
				},
			}, want: "{\"workers\":[],\"fuses\":[]}",
			wantUpdated: false,
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

		runtimeObjs = append(runtimeObjs, testcase.configMap)

		// runtimeObjs = append(runtimeObjs, testcase.fields.pods)
	}
	c := fake.NewFakeClientWithScheme(testScheme, runtimeObjs...)

	for _, testcase := range testcases {
		engine := getTestThinEngineNode(c, testcase.fields.name, testcase.fields.namespace, true)
		runtimeInfo, err := base.BuildRuntimeInfo(testcase.fields.name,
			testcase.fields.namespace,
			"thin",
			datav1alpha1.TieredStore{})
		if err != nil {
			t.Errorf("BuildRuntimeInfo() error = %v", err)
		}

		engine.Helper = ctrlhelper.BuildHelper(runtimeInfo, c, engine.Log)
		updated, err := engine.UpdateRuntimeSetConfigIfNeeded()
		if err != nil {
			t.Errorf("Got error %t.", err)
		}

		cm := corev1.ConfigMap{}
		err = c.Get(context.TODO(), types.NamespacedName{
			Namespace: testcase.configMap.Namespace,
			Name:      testcase.configMap.Name,
		}, &cm)
		if err != nil {
			t.Errorf("Got error %t.", err)
		}
		got := cm.Data["runtime.json"]
		if !reflect.DeepEqual(testcase.want, got) {
			t.Errorf("testcase %v UpdateRuntimeSetConfigIfNeeded()'s wanted %v, actual %v",
				testcase.name, testcase.want, got)
		}

		if testcase.wantUpdated != updated {
			t.Errorf("testcase %v UpdateRuntimeSetConfigIfNeeded()'s wantUpdated %v, actual %v",
				testcase.name, testcase.wantUpdated, updated)
		}

		// 2.Try the second time to make sure it idempotent and no update

		updated, err = engine.UpdateRuntimeSetConfigIfNeeded()
		if err != nil {
			t.Errorf("Got error %t.", err)
		}
		cm = corev1.ConfigMap{}
		err = c.Get(context.TODO(), types.NamespacedName{
			Namespace: testcase.configMap.Namespace,
			Name:      testcase.configMap.Name,
		}, &cm)
		if err != nil {
			t.Errorf("Got error %t.", err)
		}
		got = cm.Data["runtime.json"]
		if !reflect.DeepEqual(testcase.want, got) {
			t.Errorf("testcase %v UpdateRuntimeSetConfigIfNeeded()'s wanted %v, actual %v",
				testcase.name, testcase.want, got)
		}
		if updated {
			t.Errorf("testcase %v UpdateRuntimeSetConfigIfNeeded()'s wantUpdated false, actual %v",
				testcase.name, updated)
		}
		// if reflect.DeepEqual()

	}
}
