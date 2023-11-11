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

package nodeaffinitywithcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// default tiered locality to be compatible with fluid 0.9 logic
	tieredLocalityConfigMap = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.TieredLocalityConfigMapName,
			Namespace: fluidNameSpace,
		},
		Data: map[string]string{
			"tieredLocality": "" +
				"preferred:\n" +
				"  # fluid existed node affinity, the name can not be modified.\n" +
				"  - name: fluid.io/node\n" +
				"    weight: 100\n" +
				"required:\n" +
				"  - fluid.io/node\n",
		},
	}
	alluxioRuntime = &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio-runtime",
			Namespace: "fluid-test",
		},
	}
)

func TestPlugin(t *testing.T) {
	var (
		client client.Client
	)
	plugin := NewPlugin(client)
	if plugin.GetName() != Name {
		t.Errorf("GetName expect %v, got %v", Name, plugin.GetName())
	}
}

func TestGetPreferredSchedulingTermWithGlobalMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Test case 1: Global fuse with selector enable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{"test1": "test1"})
	term := getPreferredSchedulingTerm(runtimeInfo, 100)

	expectTerm := corev1.PreferredSchedulingTerm{
		Weight: 100,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      runtimeInfo.GetCommonLabelName(),
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	}

	if !reflect.DeepEqual(*term, expectTerm) {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", expectTerm, term)
	}

	// Test case 2: Global fuse with selector disable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})
	term = getPreferredSchedulingTerm(runtimeInfo, 100)

	if !reflect.DeepEqual(*term, expectTerm) {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", expectTerm, term)
	}
}

func TestMutateOnlyRequired(t *testing.T) {
	schema := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(schema)
	_ = corev1.AddToScheme(schema)
	var (
		client   = fake.NewFakeClientWithScheme(schema, tieredLocalityConfigMap, alluxioRuntime)
		schedPod *corev1.Pod
	)

	plugin := NewPlugin(client)
	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio", datav1alpha1.TieredStore{})
	// enable Preferred scheduling
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	schedPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset.test10-ds.sched": "required",
			},
		},
	}

	// labeled dataset not exist, no err
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
	if err != nil {
		t.Errorf("expect error is nil, but get %v", err)
	}
	// reset injected scheduling terms
	schedPod.Spec = corev1.PodSpec{}

	// labeled dataset exist with nil value, not inject
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": nil})
	if err != nil {
		t.Errorf("expect error is nil")
	}
	// reset injected scheduling terms
	schedPod.Spec = corev1.PodSpec{}

	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": runtimeInfo})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if len(schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
		t.Errorf("fail to mutate pod, not add node affinity")
	}

	if schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		t.Errorf("fail to mutate pod, not need to add Preferred scheduling term")
	}
	// reset injected scheduling terms
	schedPod.Spec = corev1.PodSpec{}
}

func TestMutateOnlyPrefer(t *testing.T) {
	schema := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(schema)
	_ = corev1.AddToScheme(schema)
	var (
		client = fake.NewFakeClientWithScheme(schema, tieredLocalityConfigMap, alluxioRuntime)
		pod    *corev1.Pod
	)

	plugin := NewPlugin(client)
	if plugin.GetName() != Name {
		t.Errorf("GetName expect %v, got %v", Name, plugin.GetName())
	}

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio", datav1alpha1.TieredStore{})
	// enable Preferred scheduling
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if shouldStop {
		t.Errorf("expect shouldStop as false, but got %v", shouldStop)
	}

	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
	if err != nil {
		t.Errorf("expect error is nil")
	}

}

func TestMutateBothRequiredAndPrefer(t *testing.T) {
	schema := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(schema)
	_ = corev1.AddToScheme(schema)
	var (
		client   = fake.NewFakeClientWithScheme(schema, tieredLocalityConfigMap, alluxioRuntime)
		schedPod *corev1.Pod
	)

	plugin := NewPlugin(client)
	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio", datav1alpha1.TieredStore{})
	// set global true to enable prefer
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	schedPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
				"fluid.io/dataset.no_exist.sched":                    "required",
			},
		},
	}
	runtimeInfos := map[string]base.RuntimeInfoInterface{
		alluxioRuntime.Name:   runtimeInfo,
		"prefer_dataset_name": runtimeInfo,
	}
	_, err = plugin.Mutate(schedPod, runtimeInfos)

	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if len(schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
		t.Errorf("fail to mutate pod, not add node required scheduling term")
	}

	if len(schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 {
		t.Errorf("fail to mutate pod, not add node Preferred scheduling term")
	}

	if len(runtimeInfos) != 2 {
		t.Errorf("mutate should not modify the parameter runtimeInfo")
	}
}

func TestTieredLocality(t *testing.T) {
	customizedTieredLocalityConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.TieredLocalityConfigMapName,
			Namespace: fluidNameSpace,
		},
		Data: map[string]string{
			common.TieredLocalityDataNameInConfigMap: "" +
				"preferred:\n" +
				"  # fluid existed node affinity, the name can not be modified.\n" +
				"  - name: fluid.io/node\n" +
				"    weight: 100\n" +
				"  # runtime worker's rack label name, can be changed according to k8s environment.\n" +
				"  - name: topology.kubernetes.io/rack\n" +
				"    weight: 50\n" +
				"  # runtime worker's zone label name, can be changed according to k8s environment.\n" +
				"  - name: topology.kubernetes.io/zone\n" +
				"    weight: 10\n" +
				"required:\n" +
				"  - fluid.io/node\n",
		},
	}

	alluxioRuntime = &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio-runtime",
			Namespace: "fluid-test",
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "topology.kubernetes.io/rack",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"rack-a"},
								},
								{
									Key:      "topology.kubernetes.io/zone",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"zone-a"},
								},
							},
						},
					},
				},
			},
		},
	}
	schema := runtime.NewScheme()
	_ = corev1.AddToScheme(schema)
	_ = datav1alpha1.AddToScheme(schema)
	client := fake.NewFakeClientWithScheme(schema, customizedTieredLocalityConfigMap, alluxioRuntime)

	runtimeInfo, _ := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio", datav1alpha1.TieredStore{})
	// set global true to enable prefer
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	type args struct {
		plugin       *NodeAffinityWithCache
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
	}
	type wanted struct {
		pod *corev1.Pod
	}

	var tests = []struct {
		name   string
		args   args
		wanted wanted
	}{
		{
			name: "tiered locality with dataset sched",
			args: args{
				plugin: NewPlugin(client),
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
						},
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{
					alluxioRuntime.Name: runtimeInfo,
				},
			},
			wanted: wanted{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
						},
					},
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
									NodeSelectorTerms: []corev1.NodeSelectorTerm{
										{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      runtimeInfo.GetCommonLabelName(),
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"true"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "tiered locality",
			args: args{
				plugin: NewPlugin(client),
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{
					alluxioRuntime.Name: runtimeInfo,
				},
			},
			wanted: wanted{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
									{
										Weight: 100,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      runtimeInfo.GetCommonLabelName(),
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"true"},
												},
											},
										},
									},
									{
										Weight: 50,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "topology.kubernetes.io/rack",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"rack-a"},
												},
											},
										},
									},
									{
										Weight: 10,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "topology.kubernetes.io/zone",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"zone-a"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "pod customized tiered locality",
			args: args{
				plugin: NewPlugin(client),
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
									{
										Weight: 100,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "topology.kubernetes.io/rack",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"rack-a"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{
					alluxioRuntime.Name: runtimeInfo,
				},
			},
			wanted: wanted{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
									{
										Weight: 100,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "topology.kubernetes.io/rack",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"rack-a"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.args.plugin.Mutate(tt.args.pod, tt.args.runtimeInfos)
			if err != nil {
				t.Errorf("get err %v", err)
			}
			if !reflect.DeepEqual(tt.args.pod.Spec.Affinity, tt.wanted.pod.Spec.Affinity) {
				t.Errorf("wanted %v, but get %v", tt.wanted.pod.Spec.Affinity, tt.args.pod.Spec.Affinity)
			}
		})
	}
}
