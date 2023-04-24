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

