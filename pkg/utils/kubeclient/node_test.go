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

package kubeclient

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetNode(t *testing.T) {
	testNodeInputs := []*corev1.Node{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1"},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "test2"},
	}}

	testNodes := []runtime.Object{}

	for _, ns := range testNodeInputs {
		testNodes = append(testNodes, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNodes...)

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *corev1.Node
	}{
		{
			name: "Node doesn't exist",
			args: args{
				name: "notExist",
			},
			want: nil,
		},
		{
			name: "Node exists",
			args: args{
				name: "test1",
			},
			want: testNodeInputs[0].DeepCopy(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, _ := GetNode(client, tt.args.name)

			if tt.want == nil {
				if want != nil {
					t.Errorf("testcase %v GetNode()'s expected is %v, result is %v", tt.name, tt.want, want)
				}
			} else {
				if want == nil {
					t.Errorf("testcase %v GetNode()'s expected is %v, result is %v", tt.name, tt.want, want)
				} else if want.Name != tt.args.name {
					t.Errorf("testcase %v GetNode()'s expected is %v, result is %v", tt.name, tt.want.Name, want.Name)
				}
			}

		})
	}
}

func TestIsReady(t *testing.T) {

	testNodeInputs := []*corev1.Node{{
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
	}, {
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
	}}

	type args struct {
		node corev1.Node
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Node Ready",
			args: args{
				node: *testNodeInputs[0],
			},
			want: true,
		}, {
			name: "Node not Ready",
			args: args{
				node: *testNodeInputs[1],
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if want := IsReady(tt.args.node); want != tt.want {
				t.Errorf("testcase %v IsReady()'s wanted %v, actual %v", tt.args.node.Name, tt.want, want)
			}
		})
	}
}

func TestGetIpAddressesOfNodes(t *testing.T) {
	type args struct {
		nodes []corev1.Node
	}
	tests := []struct {
		name            string
		args            args
		wantIpAddresses []string
	}{
		// TODO: Add test cases.
		{
			name: "duplicateNodes",
			args: args{
				nodes: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.1.1",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.1-1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.1.1",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.4",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.1.4",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.2.101",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.2.101",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.51",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.2.51",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "10.152.16.23",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "10.152.16.23",
								},
							},
						},
					},
				},
			},
			wantIpAddresses: []string{
				"10.152.16.23",
				"192.168.1.1",
				"192.168.1.4",
				"192.168.2.51",
				"192.168.2.101",
			},
		},
		{
			name: "noDuplicateNodes",
			args: args{
				nodes: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.1.1",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.4",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.1.4",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.2.101",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.2.101",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "192.168.1.51",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "192.168.2.51",
								},
							},
						},
					}, {
						ObjectMeta: metav1.ObjectMeta{
							Name: "10.152.16.23",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Type:    corev1.NodeInternalIP,
									Address: "10.152.16.23",
								},
							},
						},
					},
				},
			},
			wantIpAddresses: []string{
				"10.152.16.23",
				"192.168.1.1",
				"192.168.1.4",
				"192.168.2.51",
				"192.168.2.101",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIpAddresses := GetIpAddressesOfNodes(tt.args.nodes); !reflect.DeepEqual(gotIpAddresses, tt.wantIpAddresses) {
				t.Errorf("Name %v GetIpAddressesOfNodes() = %v, want %v",
					tt.name,
					gotIpAddresses,
					tt.wantIpAddresses)
			}
		})
	}
}
